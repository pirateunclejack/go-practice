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

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

func GetMenus() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        result, err := menuCollection.Find(context.TODO(), bson.M{})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "failed to get menus: " + err.Error(),
            })
        }

        var allMenus []bson.M
        if err = result.All(ctx, &allMenus); err != nil {
            log.Fatal("failed to save all menus to allMenus: ", err)
        }

        c.JSON(http.StatusOK, allMenus)
    }
}

func GetMenu() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        menuId := c.Param("menu_id")
        var menu models.Menu

        err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "failed to get menu: " + err.Error(),
            })
        }

        c.JSON(http.StatusOK, menu)
    }
}

func CreateMenu() gin.HandlerFunc {
    return func(c *gin.Context) {
        var menu models.Menu
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        if err := c.BindJSON(&menu); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "failed to parse menu: " + err.Error(),
            })
            return
        }

        validationErr := validate.Struct(menu)
        if validationErr != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "failed to validate menu: " + validationErr.Error(),
            })
            return
        }

        menu.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
        menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
        menu.ID = primitive.NewObjectID()
        menu.Menu_id = menu.ID.Hex()

        result, insertErr := menuCollection.InsertOne(ctx, menu)
        if insertErr != nil {
            msg := fmt.Sprintf("failed to insert menu: %v", insertErr)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": msg,
            })
            return
        }

        c.JSON(http.StatusOK, result)
    }
}

func inTimeSpan(start, end, now time.Time) bool {
	return start.After(now) && end.After(start)
}

func UpdateMenu() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        var menu models.Menu

        if err := c.BindJSON(&menu); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "failed to parse menu: " + err.Error(),
            })
            return
        }

        menuId := c.Param("menu_id")
        filter := bson.M{"menu_id": menuId}

        var updateObj primitive.D

        if menu.Start_Date != nil && menu.End_Date != nil {
            if !inTimeSpan(*menu.Start_Date, *menu.End_Date, time.Now() ) {
                msg := "menu time wrong"
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": msg,
                })
                return
            }

            updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.Start_Date})
            updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.End_Date})

            if menu.Name != "" {
                updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
            }
            if menu.Category != "" {
                updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
            }
            menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
            updateObj = append(updateObj, bson.E{Key: "updated_at", Value: menu.Updated_at})

            upsert := true
            opt := options.UpdateOptions{
                Upsert: &upsert,
            }

            result, err := menuCollection.UpdateOne(
                ctx,
                filter,
                bson.D{
                    {Key: "$set", Value: updateObj},
                },
                &opt,
            )
            if err != nil {
                msg := "failed to update menu: " + err.Error()
                c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
                return
            }

            c.JSON(http.StatusOK, result)
        }
    }
}
