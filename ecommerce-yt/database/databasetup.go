package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSet() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("failed to load.env file: %v", err)
	}
	db := os.Getenv("MONGODB")
	if db == "" {
		log.Fatalf("failed to load MONGODB environment variable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	opts := options.Client().ApplyURI(db)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}

	defer cancel()

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("failed to test connection to mongodb: %v", err)
		return nil
	}

	fmt.Println("successfully connected to mongdb")

	return client
}

var Client *mongo.Client = DBSet()

func UserData(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("Ecommerce").Collection(collectionName)
	return collection
}

func ProductData(client *mongo.Client, collectionName string) *mongo.Collection {
	var productCollection *mongo.Collection = client.Database("Ecommerce").Collection(collectionName)
	return productCollection
}
