package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/pirateunclejack/go-practice/restaurant-management/controllers"
)

func MenuRoutes(incomingRoutes *gin.Engine)  {
    incomingRoutes.GET("/menus", controller.GetMenus())
    incomingRoutes.GET("/menu/:menu_id", controller.GetMenu())
    incomingRoutes.POST("/menu", controller.CreateMenu())
    incomingRoutes.PATCH("/menu/:menu_id", controller.UpdateMenu())
}
