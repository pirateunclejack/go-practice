package database

import (
	"context"
	"log"
	"time"

	"github.com/pirateunclejack/go-practice/gql-yt/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var connectionString string = "mongodb://root:example@localhost:27017/"

type DB struct {
	client *mongo.Client
}

func Connect() *DB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v\n", err)
	}
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("failed to test connect to mongodb: %v", err)
	}

	return &DB{
		client: client,
	}
}

func (db *DB) GetJob(id string) *model.JobListing {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	var jobListing model.JobListing

	err := jobCollec.FindOne(ctx, filter).Decode(&jobListing)
	if err != nil {
		log.Fatalf("failed to find job: %v\n", err)
	}
	return &jobListing
}

func (db *DB) GetJobs() []*model.JobListing {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var jobListings []*model.JobListing
	cursor, err := jobCollec.Find(ctx, bson.D{})
	if err != nil {
		log.Fatalf("failed to list jobs: %v\n", err)
	}

	if err := cursor.All(context.TODO(), &jobListings); err != nil {
		log.Fatalf("failed to decode jobs: %v\n", err)
	}

	return jobListings
}

func (db *DB) CreateJobListing(jobInfo model.CreateJobListingInput) *model.JobListing {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	inserted, err := jobCollec.InsertOne(ctx, bson.M{
		"title": jobInfo.Title,
		"description": jobInfo.Description,
		"url": jobInfo.URL,
		"company": jobInfo.Company,
	})
	if err != nil {
		log.Fatalf("failed to insert job: %v", err)
	}

	insertedID := inserted.InsertedID.(primitive.ObjectID).Hex()
	jobListing := model.JobListing {
		ID: insertedID,
		Title: jobInfo.Title,
		Company: jobInfo.Company,
		Description: jobInfo.Description,
		URL: jobInfo.URL,
	}

	return &jobListing
}

func (db *DB) UpdateJobListing(jobID string, jobInfo *model.UpdateJobListingInput) *model.JobListing {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	updateJobInfo := bson.M{}
	if jobInfo.Title != nil {
		updateJobInfo["title"] = jobInfo.Title
	}
	if jobInfo.Description != nil {
		updateJobInfo["description"] = jobInfo.Description
	}
	if jobInfo.URL != nil {
		updateJobInfo["url"] = jobInfo.URL
	}

	_id, _ := primitive.ObjectIDFromHex(jobID)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": updateJobInfo}

	results := jobCollec.FindOneAndUpdate(
		ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(1))
	var jobListing model.JobListing
	if err := results.Decode(&jobListing); err != nil {
		log.Fatalf("failed to decode jobListing: %v", err)
	}

	return &jobListing
}

func (db *DB) DeleteJobListing(jobID string) *model.DeleteJobResponse {
	jobCollec := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(jobID)
	filter := bson.M{"_id": _id}

	_, err := jobCollec.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatalf("failed to delete joblisting: %v", err)
	}

	return &model.DeleteJobResponse{
		DeleteJobID: jobID,
	}
}