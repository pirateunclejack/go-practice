package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/pirateunclejack/go-practice/golang-jwt-project/controllers"
	"github.com/pirateunclejack/go-practice/golang-jwt-project/middleware"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())
}