package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/pirateunclejack/go-practice/restaurant-management/controllers"
)

func TableRoutes(incomingRoutes *gin.Engine)  {
    incomingRoutes.GET("/tables", controller.GetTables())
    incomingRoutes.GET("/table/:table_id", controller.GetTable())
    incomingRoutes.POST("/table", controller.CreateTable())
    incomingRoutes.PATCH("/table/:table_id", controller.UpdateTable())
}
