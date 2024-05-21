package main

import (
	"log"

	"github.com/pirateunclejack/go-practice/golang-csrf-project/db"
	"github.com/pirateunclejack/go-practice/golang-csrf-project/server"
	"github.com/pirateunclejack/go-practice/golang-csrf-project/server/middleware/myJwt"
)

var host = "localhost"
var port = "9000"

func main() {
    db.InitDB()
    jwtErr := myJwt.InitJWT()
    if jwtErr != nil {
        log.Println("JWT initialization failed", jwtErr)
    }

    serverErr := server.StartServer(host, port)
    if serverErr != nil {
        log.Println("Server initialization failed", serverErr)
    }
}
