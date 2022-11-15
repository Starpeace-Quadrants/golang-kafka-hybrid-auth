package main

import (
	"github.com/gookit/event"
	"github.com/ronappleton/gk-kafka"
	"github.com/ronappleton/golang-kafka-hybrid-authentication/apiserver"
	"github.com/ronappleton/golang-kafka-hybrid-authentication/consumer"
	"github.com/ronappleton/golang-kafka-hybrid-authentication/storage/mongo"
)

func main() {
	db, err := mongo.NewDatabase()
	if err != nil {
		panic(err.Error())
	}

	db.Start()

	event.On("messageReceived", event.ListenerFunc(func(e event.Event) error {
		consumer.ProcessMessage(e, db.Client)
		return nil
	}), event.Normal)

	go kafka.SaramaConsume("kafka:9092", "auth", "authentication_in")

	api, err := apiserver.NewAPIServer(db.Client)
	if err != nil {
		panic(err.Error())
	}

	api.Start()
}
