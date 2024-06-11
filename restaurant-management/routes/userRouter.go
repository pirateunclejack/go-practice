package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pirateunclejack/go-practice/restaurant-management/controllers"
)


func UserRoutes(incomingRoutes *gin.Engine)  {
    incomingRoutes.GET("/users", controllers.GetUsers())
    incomingRoutes.GET("/user/:user_id", controllers.GetUser())
    incomingRoutes.POST("/user/signup", controllers.SignUp())
    incomingRoutes.POST("/user/login", controllers.Login())
}
