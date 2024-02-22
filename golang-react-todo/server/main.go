package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pirateunclejack/go-practice/golang-react-todo/router"
)

func main() {
	r := router.Router()
	fmt.Println("Starting the server on port 9000...")

	log.Fatal(http.ListenAndServe(":9000", r))
}