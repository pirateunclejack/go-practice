package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pirateunclejack/go-practice/mongo-golang/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type UserController struct {
	client *mongo.Client
}

func NewUserController(c *mongo.Client) *UserController {
	return &UserController{c}
}

func (uc UserController) GetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params)  {
	id := p.ByName("id")

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	u := models.User{}

	if err := uc.client.Database("mongo-golang").Collection("users").FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&u); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	uj, err := json.Marshal(u)
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", uj)
}

func (uc UserController) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {
	u := models.User{}

	json.NewDecoder(r.Body).Decode(&u)

	u.Id = primitive.NewObjectID()

	result, err := uc.client.Database("mongo-golang").Collection("users").InsertOne(context.TODO(), u)
	if err != nil {
		fmt.Println(err)
	}
	if roid, ok := result.InsertedID.(primitive.ObjectID);ok {
		fmt.Println(roid.String())
	}
	uj, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s\n", uj)

}

func (uc UserController) DeleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params)  {
	id := p.ByName("id")

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	result, err := uc.client.Database("mongo-golang").Collection("users").DeleteOne(context.TODO(),  bson.M{"_id": oid})

	if  err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	uj, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", uj)
}
