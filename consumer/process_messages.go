package consumer

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/event"
	kafka "github.com/ronappleton/gk-kafka"
	transport "github.com/ronappleton/gk-message-transport"
	"github.com/ronappleton/golang-kafka-hybrid-authentication/storage/mongo/tables"
	"github.com/ronappleton/golang-kafka-hybrid-authentication/utilities"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strconv"
	"strings"
	"time"
)

func ProcessMessage(event event.Event, db *mongo.Client) {
	message := fmt.Sprint(event.Get("message"))
	key := fmt.Sprint(event.Get("key"))

	serviceMessage, err := transport.DecodeIncomingServiceMessage([]byte(message))
	if err != nil {
		log.Fatalf("message could not be decoded: %v", fmt.Sprint(message))
		return
	}

	switch command := serviceMessage.Command; command {
	case "authenticate":
		processAuthenticationRequest(key, db, *serviceMessage)
	case "ban_list":
		processBanListRequest(key, db, *serviceMessage)
	}
}

func processAuthenticationRequest(key string, db *mongo.Client, message transport.IncomingServiceMessage) {
	bannedIps := tables.BannedIp{Host: message.Arguments["host"]}
	banned := false
	if bannedIps.Fetch(db) {
		banned = true
	}

	user := &tables.User{SessionId: message.Arguments["sessionId"]}
	valid := user.ValidateSession(db)

	if !valid {
		processBan(db, message)
	}

	reply := transport.OutgoingServiceMessage{
		Topic: message.Topic,
		Reply: strconv.FormatBool(valid && !banned),
	}

	replyBytes, _ := reply.Encode()
	topic := kafka.Topic{
		Topic:     "authentication_out",
		Leader:    "kafka:9092",
		Partition: 0,
	}

	kafka.Produce([]byte(key), replyBytes, topic, time.Now().Add(10*time.Second))
}

func processBan(db *mongo.Client, message transport.IncomingServiceMessage) {
	banned := tables.BannedIp{
		Host:     message.Arguments["host"],
		BannedAt: time.Time{},
	}

	banned.Upsert(db)
}

// processBanListRequest We get all banned ips, split into chunks of 100, we then stringify the chunks
// encode them and send them back to the relay one by one to prevent huge messages, this is used on
// startup of the relay server to ensure we keep banned ips from using the services
func processBanListRequest(key string, db *mongo.Client, message transport.IncomingServiceMessage) {
	bannedIps := tables.BannedIp{}
	ips := bannedIps.FetchAll(db)
	chunks := utilities.ChunkStringSlice(ips, 100)

	for _, chunk := range chunks {
		sliceString := strings.Join(chunk, ",")
		encodedString, _ := json.Marshal(sliceString)
		reply := transport.OutgoingServiceMessage{
			Topic: message.Topic,
			Reply: string(encodedString),
		}

		replyBytes, _ := reply.Encode()
		topic := kafka.Topic{
			Topic:     "authentication_out",
			Leader:    "kafka:9092",
			Partition: 0,
		}

		go kafka.Produce([]byte(key), replyBytes, topic, time.Now().Add(10*time.Second))
	}
}
