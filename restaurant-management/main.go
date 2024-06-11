package main

import (
	"os"

	"github.com/pirateunclejack/go-practice/restaurant-management/helpers"
	"github.com/pirateunclejack/go-practice/restaurant-management/middleware"
	"github.com/pirateunclejack/go-practice/restaurant-management/routes"

	"github.com/gin-gonic/gin"
)


func main() {
    helpers.InitHelper()
    port := os.Getenv("PORT")
    if port == "" {
        port = "8000"
    }

    router := gin.New()
    router.Use(gin.Logger())

    routes.UserRoutes(router)
    router.Use(middleware.Authentication())
    
    routes.FoodRoutes(router)
    routes.MenuRoutes(router)
    routes.TableRoutes(router)
    routes.OrderRoutes(router)
    routes.OrderItemRoutes(router)
    routes.InvoiceRoutes(router)

    router.Run(":" + port)
}
