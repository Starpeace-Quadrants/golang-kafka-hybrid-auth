package apiserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
)

type APIServer struct {
	addr     string
	dbClient *mongodb.Client
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
		Addr:    s.addr,
		Handler: s.router(),
	}

	logrus.WithField("addr", srv.Addr).Info("starting server")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("listen: %s\n", err)
	}
}

func (s *APIServer) router() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/", s.defaultRoute)
	router.HandleFunc("/add/trainers", s.addTrainers)
	return router
}

func (s *APIServer) addTrainers(w http.ResponseWriter, r *http.Request) {
	collection := s.dbClient.Database("test").Collection("trainers")

	type Trainer struct {
		Name string
		Age  int
		City string
	}

	ash := Trainer{"Ash", 10, "Pallet Town"}
	misty := Trainer{"Misty", 10, "Cerulean City"}
	brock := Trainer{"Brock", 15, "Pewter City"}

	trainers := []interface{}{ash, misty, brock}

	insertManyResult, err := collection.InsertMany(context.TODO(), trainers)
	if err != nil {
		log.Fatal(err)
	}

	output := fmt.Sprint("Inserted multiple documents: ", insertManyResult.InsertedIDs)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}

func (s *APIServer) defaultRoute(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello Ron"))
}
