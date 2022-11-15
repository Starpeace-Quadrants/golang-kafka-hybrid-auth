package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/ronappleton/golang-kafka-hybrid-authentication/storage/mongo/tables"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/idtoken"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type APIServer struct {
	addr     string
	dbClient *mongodb.Client
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SessionResponse struct {
	SessionId string `json:"sessionId"`
}

func NewAPIServer(client *mongodb.Client) (*APIServer, error) {
	addr := os.Getenv("API_SERVER_ADDR")
	if addr == "" {
		return nil, errors.New("api addr cannot be blank")
	}

	return &APIServer{
		addr:     addr,
		dbClient: client,
	}, nil
}

// Start starts a server with a stop channel
func (s *APIServer) Start() {
	srv := &http.Server{
		Addr: s.addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.router(),
	}

	logrus.WithField("addr", srv.Addr).Info("starting server")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("listen: %s\n", err)
	}
}

func (s *APIServer) router() http.Handler {
	router := mux.NewRouter()

	// New rate limiter = 1 request per second per ip limited to get requests
	rateLimiter := tollbooth.NewLimiter(1, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	rateLimiter.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"})
	rateLimiter.SetMethods([]string{"GET"})

	router.Handle("/google/verify", tollbooth.LimitFuncHandler(rateLimiter, s.GoogleVerify))

	c := cors.New(cors.Options{
		AllowedOrigins:         []string{"http://localhost:4000"},
		AllowOriginFunc:        nil,
		AllowOriginRequestFunc: nil,
		AllowedMethods:         nil,
		AllowedHeaders:         []string{"Authorization"},
		ExposedHeaders:         nil,
		MaxAge:                 0,
		AllowCredentials:       true,
		OptionsPassthrough:     false,
		OptionsSuccessStatus:   0,
		Debug:                  false,
	})

	return c.Handler(router)
}

func (s *APIServer) GoogleVerify(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	payload, err := idtoken.Validate(context.Background(), token, "939293573845-e3e5t507011f13rid8ccu4iv4p6be2i8.apps.googleusercontent.com")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Fatal("verification failed with error", err)

		return
	}

	if payload.Claims["email_verified"] == false {
		w.WriteHeader(http.StatusUnauthorized)

		err, _ := json.Marshal(&ErrorResponse{Error: "Your Google email address must be verified!"})
		w.Write(err)

		return
	}

	u := uuid.New()

	user := &tables.User{
		Email:     fmt.Sprintf("%s", payload.Claims["email"]),
		SessionId: u.String(),
		Provider:  "google",
		CreatedAt: time.Now().UTC(),
		CreatedIp: r.RemoteAddr,
	}

	user.UpsertUser(s.dbClient)

	w.WriteHeader(http.StatusOK)

	sessionResponse, _ := json.Marshal(SessionResponse{
		SessionId: u.String(),
	})

	w.Write(sessionResponse)

	t := &tables.User{Email: fmt.Sprintf("%s", payload.Claims["email"])}

	fmt.Sprintf("Stored User Record: %v", t.Fetch(s.dbClient))
}
