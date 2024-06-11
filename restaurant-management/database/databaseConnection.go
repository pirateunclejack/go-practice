package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client {
    MongoDb := "mongodb://root:example@localhost:27017/"
    fmt.Println("Connecting to mongodb...")

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDb))
    if err != nil {
        log.Fatal("failed to connect mongodb: ", err)
    }

    fmt.Println("connected to mongodb")
    return client
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
    var collection *mongo.Collection = 
        client.Database("restaurant").Collection((collectionName))

    return collection
}
