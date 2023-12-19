package main

import (
	"github.com/labstack/echo/v4"
	"github.com/pirateunclejack/go-practice/go-echo-yt/cmd/api/handlers"
)

func main() {
	e := echo.New()
	e.GET("/health-check", handlers.HealthCheckHandler)
	e.GET("posts", handlers.PostIndexHandler)
	e.GET("post/:id", handlers.PostSingleHandler)

	e.Logger.Fatal(e.Start(":1323"))
}
