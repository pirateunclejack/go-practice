package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/pirateunclejack/go-practice/restaurant-management/database"
	"github.com/pirateunclejack/go-practice/restaurant-management/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
    return func (c *gin.Context)  {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        var recordPerPage int
        var page int
        var startIndex int
        if c.Query("recordPerPage") == "" {
            recordPerPage = 10
        } else {
            recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
            if err != nil || recordPerPage < 1 {
                recordPerPage = 10
            }
        }

        if c.Query("page") == "" {
			page = 1
        } else {
            page, err := strconv.Atoi(c.Query("page"))
            if err != nil || page < 1 {
                page = 1
            }
        }

        if c.Query("startIndex") == "" {
            startIndex = (page - 1) * recordPerPage
        } else {
            var err error
            startIndex, err = strconv.Atoi(c.Query("startIndex"))
            if err != nil {
                msg := "failed to convert startIndex to int: " + err.Error()
                c.JSON(
                    http.StatusBadRequest,
                    gin.H{"error": msg},
                )
                return
            }
        }

        // matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
        // groupStage := bson.D{{Key: "$group", Value: bson.D{
        //     {Key: "_id",         Value: bson.D{{Key: "_id",   Value: "null"}}},
        //     {Key: "total_count", Value: bson.D{{Key: "$sum",  Value: 1}}},
        //     {Key: "data",        Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
        // }}}
        // progectStage := bson.D{{Key: "$project", Value: bson.D{
        //     {Key: "_id", Value: 0},
        //     {Key: "total_count", Value: 1},
        //     {Key: "food_items", Value: bson.D{
        //         {Key: "$slice", Value: []interface{}{
        //             "$data", startIndex, recordPerPage,
        //         }},
        //     }},
        // }}}

        // result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
        //     matchStage, groupStage, progectStage,
        // })
        // if err != nil {
        //     c.JSON(http.StatusInternalServerError, gin.H{
        //         "error": "failed to get foods: " + err.Error(),
        //     })
        //     return
        // }

        // var allFoods []bson.M
        // if err = result.All(ctx, &allFoods); err != nil {
        //     log.Fatal("failed to save allFoods: " + err.Error())
        //     return
        // }


        // Set the options for pagination
        findOptions := options.Find()
        findOptions.SetLimit(int64(recordPerPage))
        findOptions.SetSkip(int64(startIndex))

        // Find foods with pagination
        cursor, err := foodCollection.Find(ctx, bson.M{}, findOptions)
        if err != nil {
            log.Fatal("failed to define foods cursor: ", err.Error())
        }
        defer cursor.Close(ctx)
    
        // Iterate through the cursor and decode each document into a Food struct
        var foods []models.Food

        for cursor.Next(ctx) {
            var food models.Food
            err := cursor.Decode(&food)
            if err != nil {
                log.Fatal("failed to find next food: ", err.Error())
            }
            foods = append(foods, food)
        }
    
        if err := cursor.Err(); err != nil {
            log.Fatal("foods cursor error: ", err.Error())
        }
    

        if len(foods) == 0 {
            c.JSON(http.StatusOK, nil)
        } else {
            c.JSON(http.StatusOK, foods[0])
        }
    }
}

func GetFood() gin.HandlerFunc {
    return func (c *gin.Context)  {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        foodId := c.Param("food_id")
        var food models.Food

        err := foodCollection.FindOne(ctx, bson.M{"food_id" : foodId}).Decode(&food)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "failed to get food: " + err.Error(),
            })
            return
        }

        c.JSON(http.StatusOK, food)
    }
}


func round(num float64) int {
    return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
    output := math.Pow(10, float64(precision))
    return float64(round(num*output)) / output
}

func CreateFood() gin.HandlerFunc {
    return func (c *gin.Context)  {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        var menu models.Menu
        var food models.Food

        if err := c.BindJSON(&food); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "failed to parse food: " + err.Error(),
            })
            return
        }

        validationErr := validate.Struct(food)
        if validationErr != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error" : "failed to validate food: " + validationErr.Error(),
            })
            return
        }

        err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
        if err != nil {
            msg := fmt.Sprintf("field to get menu when create food: %v", err.Error())
            c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
            return
        }

        food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
        food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
        food.ID = primitive.NewObjectID()
        food.Food_id = food.ID.Hex()
        var num = toFixed(*food.Price, 2)
        food.Price = &num

        result, insertErr := foodCollection.InsertOne(ctx ,food)
        if insertErr != nil {
            msg := fmt.Sprintf("failed to insert food: %v", insertErr)
            c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
            return
        }

        c.JSON(http.StatusOK, result)
    }
}

func UpdateFood() gin.HandlerFunc {
    return func (c *gin.Context)  {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        var menu models.Menu
        var food models.Food

        foodId := c.Param("food_id")

        if err := c.BindJSON(&food); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "failed to parse food: " + err.Error(),
            })
            return
        }

        var updateObj primitive.D

        if food.Name != nil {
            updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
        }
        if food.Price != nil {
            updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
        }
        if food.Food_image != nil {
            updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.Food_image})
        }
        if food.Menu_id != nil {
            err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
            if err != nil {
                msg := fmt.Sprintf("failed to get menu when update food: %v", err.Error())
                c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
                return
            }
            updateObj = append(updateObj, bson.E{Key: "menu_id", Value: food.Menu_id})
        }
        food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
        updateObj = append(updateObj, bson.E{Key: "updated_at",  Value: food.Updated_at})

        upsert := true
        filter := bson.M{"food_id": foodId}

        opt := options.UpdateOptions{
            Upsert: &upsert,
        }

        result, err := foodCollection.UpdateOne(
            ctx,
            filter,
            bson.D{{Key: "$set", Value: updateObj}},
            &opt,
        )
        if err != nil {
            msg := fmt.Sprintf("failed to update food: %v", err.Error())
            c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
            return
        }

        c.JSON(http.StatusOK, result)
    }
}
