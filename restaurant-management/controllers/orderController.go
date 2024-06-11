package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pirateunclejack/go-practice/restaurant-management/database"
	"github.com/pirateunclejack/go-practice/restaurant-management/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

func GetOrders() gin.HandlerFunc {
    return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        result, err := orderCollection.Find(context.TODO(), bson.M{})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "failed to get orders: " + err.Error(),
            })
            return
        }

        var allOrders []bson.M
        if err = result.All(ctx, &allOrders); err != nil {
            log.Fatal("failed to get orders: ", err.Error())
            return
        }

        c.JSON(http.StatusOK, allOrders)
    }
}

func GetOrder() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        orderId := c.Param("order_id")
        var order models.Order

        err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "failed to get order: " + err.Error(),
            })
            return
        }

        c.JSON(http.StatusOK, order)
    }
}

func CreateOrder() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        var table models.Table
        var order models.Order

        if err := c.BindJSON(&order); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "failed to parse order: " + err.Error(),
            })
            return
        }

        validationErr := validate.Struct(order)
        if validationErr != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "failed to validate order: " + validationErr.Error(),
            })
            return
        }
        
        if order.Table_id != nil {
            err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
            if err != nil {
                msg := fmt.Sprintf("failed to find table when create order: %v", err.Error())
                c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
                return
            }

            order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
            order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
            order.ID = primitive.NewObjectID()
            order.Order_id = order.ID.Hex()

            result, insertErr := orderCollection.InsertOne(ctx, order)
            if insertErr != nil {
                msg := fmt.Sprintf("failed to create order: %v", insertErr.Error())
                c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
                return
            }

            c.JSON(http.StatusOK, result)
        }
    }
}

func UpdateOrder() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        var table models.Table
        var order models.Order
        var updateObj primitive.D

        orderId := c.Param("order_id")
        if err := c.BindJSON(&order); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "failed to parse order: " + err.Error(),
            })
            return
        }

        if order.Table_id != nil {
            err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
            if err != nil {
                msg := fmt.Sprintf("failed to find table: %v", err.Error())
                c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
                return
            }
            updateObj = append(updateObj, bson.E{Key: "table_id", Value: order.Table_id})
        }

        if !order.Order_Date.IsZero() {
            updateObj = append(updateObj, bson.E{Key: "order_date", Value: order.Order_Date})
            
        }

        order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
        updateObj = append(updateObj, bson.E{Key: "updated_at", Value: order.Updated_at})
    
    
        upsert := true

        filter := bson.M{"order_id": orderId}
        opt := options.UpdateOptions{
            Upsert: &upsert,
        }

        result, err := orderCollection.UpdateOne(
            ctx,
            filter,
            bson.D{{Key: "$set", Value: updateObj}},
            &opt,
        )
        if err != nil {
            msg := fmt.Sprintf("failed to update order: %v", err.Error())
            c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
            return
        }

        c.JSON(http.StatusOK, result)
    }
}

func OrderItemOrderCreator(order models.Order) string {
    var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	orderCollection.InsertOne(ctx, order)
	return order.Order_id
}
