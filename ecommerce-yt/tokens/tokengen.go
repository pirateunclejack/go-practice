package token

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pirateunclejack/go-practice/ecommerce-yt/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email string
	First_Name string
	Last_Name string
	Uid string
	jwt.RegisteredClaims
}

var UserData *mongo.Collection = database.UserData(database.Client, "User")

var SECRET_KEY = os.Getenv("SECRET_KEY")

func TokenGenerator(
	email string, firstname string, lastname string, uid string) (
		signedtoken string, signedrefreshtoken string, err error) {
		claims := &SignedDetails{
			Email: email,
			First_Name: firstname,
			Last_Name: lastname,
			Uid: uid,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
		}
		
		refreshclaims := &SignedDetails{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
		}

		token, err := jwt.NewWithClaims(
			jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
		if err != nil {
			return "", "", err
		}
		log.Printf("new token: %v", token)

		refreshtoken, err := jwt.NewWithClaims(
			jwt.SigningMethodHS256, refreshclaims).SignedString([]byte(SECRET_KEY))
		if err != nil {
			log.Panic(err)
			return "", "", err
		}
		log.Printf("new refresh token: %v", refreshtoken)


		return token, refreshtoken, nil
}

func ValidateToken(signedtoken string) (claims *SignedDetails, msg string)  {
	token, err := jwt.ParseWithClaims(
		signedtoken, &SignedDetails{}, func(token *jwt.Token)(interface{}, error){
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		msg := err.Error()
		return nil, msg
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg := "the token is invalid"
		return nil, msg
	}
	if claims.ExpiresAt.Before(time.Now()) {
		msg := "the token has expired"
		return nil, msg
	}

	return claims, msg
	
}

func UpdateAllTokens(signedtoken string, signedrefreshtoken string, userid string)  {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateobj primitive.D

	updateobj = append(updateobj, bson.E{Key: "token", Value: signedtoken})
	updateobj = append(updateobj, bson.E{
		Key: "refresh_token", Value: signedrefreshtoken})
	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateobj = append(updateobj, bson.E{Key: "updatedat", Value: updated_at})

	upsert := true

	filter := bson.M{"user_id": userid}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := UserData.UpdateOne(ctx, filter, bson.D{{
		Key: "$set", Value: updateobj,
	}}, &opt)
	defer cancel()
	if err != nil {
		log.Panic(err)
		return
	}
}