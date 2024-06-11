package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/pirateunclejack/go-practice/restaurant-management/controllers"
)


func FoodRoutes(incomingRoutes *gin.Engine)  {
    incomingRoutes.GET("/foods", controller.GetFoods())
    incomingRoutes.GET("/food/:food_id", controller.GetFood())
    incomingRoutes.POST("/food", controller.CreateFood())
    incomingRoutes.PATCH("/food/:food_id", controller.UpdateFood())
}
