package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pirateunclejack/go-practice/go-bookstore/pkg/routes"
	"github.com/rs/cors"
)

func main() {
	r := mux.NewRouter()
	routes.RegisterBookStoreRoutes(r)
	http.Handle("/", r)

	c := cors.New(cors.Options{
        AllowedOrigins: []string{"http://localhost:3000"},
        AllowCredentials: true,
    })

	handler := c.Handler(r)

	log.Fatal(http.ListenAndServe("localhost:9010", handler))
}
