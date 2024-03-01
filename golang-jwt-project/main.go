package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pirateunclejack/go-practice/golang-jwt-project/routes"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}
	port := os.Getenv("PORT")
	if port=="" {
		port="8001"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1", func(c *gin.Context){
		c.JSON(200, gin.H{"success":"Access granted for api-1"})
	})

	router.GET("/api-2", func(c *gin.Context){
		c.JSON(200, gin.H{"success":"Access granted for api-2"})
	})

	router.Run(":" + port)
}