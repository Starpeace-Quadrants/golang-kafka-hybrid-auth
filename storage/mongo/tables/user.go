package tables

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type User struct {
	Email     string    `bson:"email" json:"email"`
	SessionId string    `bson:"session_id" json:"sessionId"`
	Provider  string    `bson:"provider" json:"provider"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
	CreatedIp string    `bson:"created_ip" json:"createdIp"`
}

func (user *User) Store(client *mongo.Client) *User {
	collection := client.Database("starpeace").Collection("users")

	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now()

	_, err := collection.InsertOne(context.TODO(), user)

	if err != nil {
		log.Fatalf("user: %s could not be stored as exists already")
	}

	return user
}

func (user *User) UpsertUser(client *mongo.Client) *User {
	collection := client.Database("starpeace").Collection("users")

	user.UpdatedAt = time.Now().UTC()

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"email", user.Email}}
	update := bson.D{{"$set", user}}

	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)

	if err != nil {
		log.Fatalf("unable to update user %v", user)
	}

	return user
}

func (user *User) Delete(client *mongo.Client) bool {
	collection := client.Database("starpeace").Collection("users")

	filter := bson.D{{"email", user.Email}}

	result, err := collection.DeleteOne(context.TODO(), filter)

	if err != nil {
		log.Fatalf("Unable to delete user: %v", err)
	}

	return result != nil
}

func (user *User) Fetch(client *mongo.Client) *User {
	collection := client.Database("starpeace").Collection("users")

	filter := bson.D{{"email", user.Email}}

	err := collection.FindOne(context.TODO(), filter).Decode(user)

	if err != nil {
		log.Fatalf("Unable to find user: %v", user)
	}

	return user
}

func (user *User) ValidateSession(client *mongo.Client) bool {
	collection := client.Database("starpeace").Collection("users")

	filter := bson.D{{"session_id", user.SessionId}}

	err := collection.FindOne(context.TODO(), filter).Decode(user)

	if err != nil {
		return false
	}

	return true
}
