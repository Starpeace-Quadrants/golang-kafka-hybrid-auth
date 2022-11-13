package consumer

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func startConsumer() {
	for {
		result := consume()
		if result == false {
			break
		}
	}
}

func consume() bool {

	conn, err := kafka.DialLeader(context.Background(), "tcp", "kafka:9092", "authentication_in", 0)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetReadDeadline(time.Time{})

	log.Println(fmt.Sprintf("%s consumer running", topic.Topic))

	batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max

	b := make([]byte, 10e3) // 10KB max per message
	for {
		n, err := batch.Read(b)
		if err != nil {
			break
		}
		fmt.Println(string(b[:n]))
	}

	if err := batch.Close(); err != nil {
		log.Fatal("failed to close batch:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close connection:", err)
	}

	serverReply, err := transport.DecodeServerMessage(b)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
