package tables

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type User struct {
	Email     string    `bson:"email"`
	Confirmed bool      `bson:"confirmed"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	Token     string    `bson:"token"`
	Provider  string    `bson:"provider"`
}

func (user *User) Store(client *mongo.Client) *User {
	collection := client.Database("starpeace").Collection("users")

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := collection.InsertOne(context.TODO(), user)

	if err != nil {
		log.Fatalf("user: %s could not be stored as exists already")
	}

	return user
}

func (user *User) Update(client *mongo.Client) *User {
	collection := client.Database("starpeace").Collection("users")

	user.UpdatedAt = time.Now()

	filter := bson.D{{"email", user.Email}}

	_, err := collection.UpdateOne(context.TODO(), filter, user)

	if err != nil {
		log.Fatalf("unable to update user %v", user)
	}
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
