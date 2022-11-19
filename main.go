package main

import (
	"github.com/gookit/event"
	"github.com/ronappleton/gk-kafka"
	"github.com/ronappleton/golang-kafka-hybrid-authentication/apiserver"
	"github.com/ronappleton/golang-kafka-hybrid-authentication/consumer"
)

func main() {
	event.On("messageReceived", event.ListenerFunc(func(e event.Event) error {
		consumer.ProcessMessage(e)
		return nil
	}), event.Normal)

	kafka.InitTopics("kafka", 9092)
	go kafka.SaramaConsume("kafka:9092", "auth", "authentication_in")

	api, err := apiserver.NewAPIServer()
	if err != nil {
		panic(err.Error())
	}

	api.Start()
}
