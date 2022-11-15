package tables

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type BannedIp struct {
	Host     string    `bson:"host" json:"host"`
	BannedAt time.Time `bson:"banned_at" json:"banned_at"`
}

func (bannedIp *BannedIp) Store(client *mongo.Client) *BannedIp {
	collection := client.Database("starpeace").Collection("banned_ips")

	bannedIp.BannedAt = time.Now().UTC()

	_, err := collection.InsertOne(context.TODO(), bannedIp)

	if err != nil {
		log.Fatalf("banned_ip: %s could not be stored as exists already")
	}

	return bannedIp
}

func (bannedIp *BannedIp) Upsert(client *mongo.Client) *BannedIp {
	collection := client.Database("starpeace").Collection("banned_ips")

	bannedIp.BannedAt = time.Now().UTC()

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"host", bannedIp.Host}}
	update := bson.D{{"$set", bannedIp.Host}}

	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)

	if err != nil {
		log.Fatalf("unable to update banned ip %v", bannedIp)
	}

	return bannedIp
}

func (bannedIp *BannedIp) Fetch(client *mongo.Client) bool {
	collection := client.Database("starpeace").Collection("banned_ips")

	filter := bson.D{{"host", bannedIp.Host}}

	err := collection.FindOne(context.TODO(), filter).Decode(bannedIp)

	if err != nil {
		return false
	}

	return true
}

func (bannedIp *BannedIp) FetchAll(client *mongo.Client) []string {
	var output []string

	collection := client.Database("starpeace").Collection("banned_ips")
	cursor, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return output
	}

	for cursor.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var elem BannedIp
		err := cursor.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		output = append(output, elem.Host)
	}

	return output
}
