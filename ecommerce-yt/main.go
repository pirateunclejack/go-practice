package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pirateunclejack/go-practice/ecommerce-yt/controllers"
	"github.com/pirateunclejack/go-practice/ecommerce-yt/database"
	"github.com/pirateunclejack/go-practice/ecommerce-yt/middleware"
	"github.com/pirateunclejack/go-practice/ecommerce-yt/routes"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	app := controllers.NewApplication(
		database.ProductData(database.Client, "Product"),
		database.UserData(database.Client, "User"),
	)

	router := gin.New()
	router.Use(gin.Logger())

	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoveItem())
	router.GET("/listcart", controllers.GetItemFromCart())
	router.POST("/addaddress", controllers.AddAddress())
	router.GET("/address/list", controllers.ListAddress())
	router.PUT("/edithomeaddress", controllers.EditHomeAddress())
	router.PUT("/editworkaddress", controllers.EditWorkAddress())
	router.GET("/deleteaddresses", controllers.DeleteAddress())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())

	log.Fatal(router.Run(":" + port))
}
