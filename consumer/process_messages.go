package consumer

import (
	"github.com/gookit/event"
	"github.com/kamva/mgm/v3"
	kafka "github.com/ronappleton/gk-kafka"
	"github.com/ronappleton/gk-kafka/storage"
	transport "github.com/ronappleton/gk-message-transport"
	"github.com/ronappleton/golang-kafka-hybrid-authentication/storage/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"time"
)

func ProcessMessage(event event.Event) {
	data := storage.New()
	data.Populate(event.Data())

	message := data.GetMessage()

	serviceMessage := transport.BytesToServiceMessage(message.Value)

	switch command := serviceMessage.Command; command {
	case "authenticate":
		processAuthenticationRequest(string(message.Key), serviceMessage)
	case "ban_list":
		processBanListRequest(string(message.Key), serviceMessage)
	}
}

func processAuthenticationRequest(key string, message transport.ServiceMessage) {
	banned := &mongo.BannedIp{}
	_ = mgm.Coll(banned).First(bson.M{"host": message.ArgumentStore.GetString("host")}, banned)

	if len(banned.Host) == 0 {
		HandleAuthenticationFailed(key, message)
	}

	user := &mongo.User{}
	_ = mgm.Coll(user).First(bson.M{"session_id": message.ArgumentStore.GetString("sessionId")}, user)

	if len(user.Email) == 0 {
		processBan(message)
		HandleAuthenticationFailed(key, message)
	}

	HandleAuthenticationSuccess(key, message)
}

func processBan(message transport.ServiceMessage) {
	banned := mongo.NewBannedIp(message.ArgumentStore.GetString("host"))
	_ = mgm.Coll(banned).Create(banned)
}

func HandleAuthenticationFailed(key string, message transport.ServiceMessage) {
	reply := transport.NewClientMessage()
	reply.Command = "authenticate"
	reply.Topic = message.Topic
	reply.Results["authentication"] = false

	replyBytes := reply.ToBytes()

	topic, _ := kafka.GetTopicByName("authentication", "out")

	kafka.Produce([]byte(key), replyBytes, topic, time.Now().Add(10*time.Second))
}

func HandleAuthenticationSuccess(key string, message transport.ServiceMessage) {
	reply := transport.NewClientMessage()
	reply.Command = "authenticate"
	reply.Topic = message.Topic
	reply.Results["authentication"] = true

	replyBytes := reply.ToBytes()

	topic, _ := kafka.GetTopicByName("authentication", "out")

	kafka.Produce([]byte(key), replyBytes, topic, time.Now().Add(10*time.Second))
}

// processBanListRequest We get all banned ips, split into chunks of 100, we then stringify the chunks
// encode them and send them back to the relay one by one to prevent huge messages, this is used on
// startup of the relay server to ensure we keep banned ips from using the services
func processBanListRequest(key string, message transport.ServiceMessage) {
	var banned []mongo.BannedIp

	_ = mgm.Coll(&mongo.BannedIp{}).SimpleFind(banned, bson.M{})

	chunks := mongo.ChunkBanned(banned, 100)

	for _, chunk := range chunks {
		var stringSlice []string
		for _, banned := range chunk {
			stringSlice = append(stringSlice, banned.Host)
		}

		sliceString := strings.Join(stringSlice, ",")

		reply := transport.NewClientMessage()
		reply.Topic = message.Topic
		reply.Results["hosts"] = sliceString

		replyBytes := reply.ToBytes()

		topic, _ := kafka.GetTopicByName("authentication", "out")

		go kafka.Produce([]byte(key), replyBytes, topic, time.Now().Add(10*time.Second))
	}
}
