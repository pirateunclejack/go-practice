package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pirateunclejack/go-practice/go-sms-verify-yt/api"
)

func main() {
	router := gin.Default()

	app := api.Config{
		Router: router,
	}

	app.Routes()

	router.Run(":8001")
}
