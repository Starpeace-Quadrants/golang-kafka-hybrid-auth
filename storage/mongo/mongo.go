package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

type Database struct {
	addr   string
	Client *mongo.Client
}

func NewDatabase() (*Database, error) {
	addr := os.Getenv("MONGO_SERVER_ADDR")
	if addr == "" {
		return nil, errors.New("mongo addr cannot be blank")
	}

	return &Database{
		addr: addr,
	}, nil
}

func (db *Database) Start() {
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://mongo%s", db.addr))
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	db.Client = client

	fmt.Println("Connected to MongoDB!")
}
