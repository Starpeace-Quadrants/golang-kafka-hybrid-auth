package apiserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kamva/mgm/v3"
	"github.com/ronappleton/gk-kafka/storage"
	"github.com/ronappleton/golang-kafka-hybrid-authentication/storage/mongo"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

// httpClientWithCerts creates an HTTP client that trusts certificates mounted in
// the provided directory. All files in the directory are read and any valid PEM
// data is added to the certificate pool.
func httpClientWithCerts(dir string) (*http.Client, error) {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return nil
		}

		pemData, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		pool.AppendCertsFromPEM(pemData)
		return nil
	})
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool},
	}

	return &http.Client{Transport: transport}, nil
}

func NewAPIServer() (*APIServer, error) {
	host := os.Getenv("API_SERVER_HOST")
	port := os.Getenv("API_SERVER_PORT")

	addr := fmt.Sprintf("%s:%s", host, port)

	return &APIServer{
		addr: addr,
	}, nil
}

// Start starts a server with a stop channel
func (s *APIServer) Start() {
	srv := &http.Server{
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.router(),
		Addr:         s.addr,
	}

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

	clientOrigin := fmt.Sprintf("%s://%s:%s", os.Getenv("CLIENT_PROTOCOL"), os.Getenv("CLIENT_DOMAIN"), os.Getenv("CLIENT_PORT"))

	c := cors.New(cors.Options{
		AllowedOrigins:         []string{clientOrigin},
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
	client, err := httpClientWithCerts("/app/certs")
	if err != nil {
		log.Println("failed to load certs from /app/certs:", err)
		client = http.DefaultClient
	}

	validator, err := idtoken.NewValidator(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("failed to create validator", err)
		return
	}

	payload, err := validator.Validate(context.Background(), token, "939293573845-e3e5t507011f13rid8ccu4iv4p6be2i8.apps.googleusercontent.com")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("verification failed with error ", err)
		w.Write([]byte(fmt.Sprintf("Authentication failed: %s", err)))
		return
	}

	data := storage.New()
	data.Populate(payload.Claims)

	if data.GetBool("email_verified") == false {
		w.WriteHeader(http.StatusUnauthorized)

		err, _ := json.Marshal(&ErrorResponse{Error: "Your Google email address must be verified!"})
		w.Write(err)
	}

	u := uuid.New()

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	fmt.Println("Host: ", host)
	user := mongo.NewUser(data.GetString("email"), u.String(), "google", host)
	options := mgm.UpsertTrueOption()

	_ = mgm.Coll(user).Update(user, options)

	w.WriteHeader(http.StatusOK)

	sessionResponse, _ := json.Marshal(SessionResponse{
		SessionId: u.String(),
	})

	w.Write(sessionResponse)

	t := &mongo.User{}

	_ = mgm.Coll(t).First(bson.M{"email": data.GetString("email")}, t)

	fmt.Sprintf("Stored User Record: %v", t)
}
