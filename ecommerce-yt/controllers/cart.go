package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pirateunclejack/go-practice/ecommerce-yt/database"
	"github.com/pirateunclejack/go-practice/ecommerce-yt/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApplication(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")
			_ = c.AbortWithError(
				http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		
		UserQueryID := c.Query("userID")
		if UserQueryID == "" {
			log.Println("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Printf("failed to get productid from hex: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.AddProductToCart(
			ctx, app.prodCollection, app.userCollection, productID, UserQueryID)
		if err != nil {
			log.Printf("failed to add product to cart: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.IndentedJSON(http.StatusOK, "successfully added to cart")
	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")
			_ = c.AbortWithError(
				http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		
		userQueryID := c.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.RemoveCartItem(
			ctx, app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(http.StatusOK, "successfully removed from cart")
	}
}

func GetItemFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")
		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{
				"error": "invalid id",
			})
			c.Abort()
			return
		}

		usert_id, _ := primitive.ObjectIDFromHex(user_id)

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var filledcart models.User
		err := UserCollection.FindOne(
			ctx, bson.D{primitive.E{Key: "_id", Value: usert_id}}).Decode(&filledcart)
		if err != nil {
			log.Printf("failed to get user when getting cart: %v", err)
			c.IndentedJSON(500, "failed to get user when getting cart")
			return
		}
		// fmt.Printf("filled cart: %v\n", filledcart)

		filter_match := bson.D{{Key: "$match", Value: bson.D{
			primitive.E{Key: "_id", Value: usert_id},
		}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{
			primitive.E{Key: "path", Value: "$usercart"},
		}}}
		grouping := bson.D{{Key: "$group", Value: bson.D{
			primitive.E{Key: "_id", Value: "$_id"},
			{Key: "total", Value: primitive.D{
				primitive.E{Key: "$sum", Value: "$usercart.price"}}},
		}}}

		pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{
			filter_match, unwind, grouping,
		})
		if err != nil {
			log.Println(err)
		}

		var listing []bson.M
		if err = pointcursor.All(ctx, &listing); err != nil {
			log.Printf("failed to add user cart to list: %v", err)
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		// log.Printf("found products in cart: %v", listing)

		for _, json := range listing {
			c.JSON(200, gin.H{
				"total": json["total"],
				"cart": filledcart.UserCart,
			})
		}
		ctx.Done()
	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("id")

		if userQueryID == "" {
			log.Panicln("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err := database.BuyItemFromCart(ctx, app.userCollection, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		c.IndentedJSON(http.StatusOK, "successfully placed the order")
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("id")
		if productQueryID == "" {
			log.Println("product id is empty")
			_ = c.AbortWithError(
				http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		
		userQueryID := c.Query("userID")
		if userQueryID == "" {
			log.Println("user id is empty")
			_ = c.AbortWithError(
				http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.InstantBuyer(
			ctx, app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(http.StatusOK, "successfully placed the order")
	}
}
