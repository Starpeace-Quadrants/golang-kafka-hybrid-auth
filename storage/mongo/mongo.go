package mongo

import (
	"fmt"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

func init() {
	host := os.Getenv("MONGO_SERVER_HOST")
	port := os.Getenv("MONGO_SERVER_PORT")
	database := os.Getenv("MONGO_SERVER_DATABASE")

	if err := mgm.SetDefaultConfig(nil, database, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", host, port))); err != nil {
		panic(err)
	}
}

type User struct {
	mgm.DefaultModel
	Email     string `bson:"email" json:"email"`
	SessionId string `bson:"session_id" json:"sessionId"`
	Provider  string `bson:"provider" json:"provider"`
	CreatedIp string `bson:"created_ip" json:"createdIp"`
}

func NewUser(email string, sessionId string, provider string, createdIp string) *User {
	return &User{
		Email:     email,
		SessionId: sessionId,
		Provider:  provider,
		CreatedIp: createdIp,
	}
}

type BannedIp struct {
	mgm.DefaultModel
	Host string `bson:"host" json:"host"`
}

func NewBannedIp(host string) *BannedIp {
	return &BannedIp{
		Host: host,
	}
}

func ChunkBanned(items []BannedIp, chunkSize int) (chunks [][]BannedIp) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}

	return append(chunks, items)
}
