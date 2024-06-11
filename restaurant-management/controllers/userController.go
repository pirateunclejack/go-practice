package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pirateunclejack/go-practice/restaurant-management/database"
	helper "github.com/pirateunclejack/go-practice/restaurant-management/helpers"
	"github.com/pirateunclejack/go-practice/restaurant-management/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

        recordPerPage := 10
        page := 1
        startIndex := 0
        var err error

        if c.Query("recordPerPage") != "" {
            recordPerPage, err = strconv.Atoi(c.Query("recordPerPage"))
            if err != nil || recordPerPage < 1 {
                recordPerPage = 10
            }
        }

        if c.Query("page") != "" {
            page, err = strconv.Atoi(c.Query("page"))
            if err != nil || page < 1 {
                page = 1
            }
        }

        startIndex = (page - 1) * recordPerPage
        if c.Query("startIndex") != "" {
            startIndex, err = strconv.Atoi(c.Query("startIndex"))
            if err != nil {
                msg := "failed to convert startIndex to int: " + err.Error()
                c.JSON(http.StatusBadRequest, gin.H{"error": msg})
                return
            }
        }


        // matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		// projectStage := bson.D{
		// 	{Key: "$project", Value: bson.D{
		// 		{Key: "_id", Value: 0},
		// 		{Key: "total_count", Value: 1},
		// 		{Key: "user_items", Value: bson.D{
        //             {Key: "$slice", Value:[]interface{}{"$data", startIndex, recordPerPage}}}},
		// 	}}}

		// result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
		// 	matchStage, projectStage})
		// defer cancel()
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
		// }

		// var allUsers []bson.M
		// if err = result.All(ctx, &allUsers); err != nil {
		// 	log.Fatal(err)
		// }
		// c.JSON(http.StatusOK, allUsers[0])


        // Set the options for pagination
        findOptions := options.Find()
        findOptions.SetLimit(int64(recordPerPage))
        findOptions.SetSkip(int64(startIndex))
        
        // Find users with pagination
        cursor, err := userCollection.Find(ctx, bson.M{}, findOptions)
        if err != nil {
            log.Fatal("failed to define users cursor: ", err.Error())
        }
        defer cursor.Close(ctx)
    
        // Iterate through the cursor and decode each document into a User struct
        var users []models.User

        for cursor.Next(ctx) {
            var user models.User
            err := cursor.Decode(&user)
            if err != nil {
                log.Fatal("failed to find next user: ", err.Error())
            }
            users = append(users, user)
        }
    
        if err := cursor.Err(); err != nil {
            log.Fatal("users cursor error: ", err.Error())
        }
    
        if len(users) > 0 {
            c.JSON(http.StatusOK, users)
        } else {
            c.JSON(http.StatusOK, gin.H{"message": "No users found"})
        }

	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		userId := c.Param("user_id")

		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
		}
		c.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		//convert the JSON data coming from postman to something that golang understands
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//validate the data based on user struct

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		//you'll check if the email has already been used by another user

		_, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}
		//hash password

		password := HashPassword(*user.Password)
		user.Password = &password

		//you'll also check if the phone no. has already been used by another user

		count, err := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exsits"})
			return
		}

		//create some extra details for the user object - created_at, updated_at, ID

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		//generate token and refersh token (generate all tokens function from helper)

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
		user.Token = &token
		user.Refresh_Token = &refreshToken
		//if all ok, then you insert this new user into the user collection

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created: %v", insertErr.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		//return status OK and send the result back

		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		var foundUser models.User

		//convert the login data from postman which is in JSON to golang readable format

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//find a user with that email and see if that user even exists

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found, login seems to be incorrect"})
			return
		}

		//then you will verify the password

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "password is not valid: "+msg})
			return
		}

		//if all goes well, then you'll generate tokens

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)

		//update tokens - token and refersh token
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		//return statusOK
		c.JSON(http.StatusOK, foundUser)
	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or password is incorrect: %v", err.Error())
		check = false
	}
	return check, msg
}
