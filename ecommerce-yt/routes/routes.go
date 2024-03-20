package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pirateunclejack/go-practice/ecommerce-yt/controllers"
)

func UserRoutes(incommingRoutes *gin.Engine)  {
	incommingRoutes.POST("/users/signup", controllers.Signup())
	incommingRoutes.POST("/users/login", controllers.Login())
	incommingRoutes.POST("/admin/addproduct", controllers.ProductViewerAdmin())
	incommingRoutes.GET("/users/productview", controllers.SearchProduct())
	incommingRoutes.GET("/users/search", controllers.SearchProductByQuery())
}