package main

import (
	"encoding/json"
	"log"
	"net/http"
)


type Message struct {
	Status string `json:"status"`
	Body string `json:"body"`
}

func endpointHandler(writer http.ResponseWriter, request *http.Request)  {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	message := Message{
		Status: "success",
		Body: "Hello, world!",
	}
	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		log.Printf("failed to encode message to response writer: %v", err)
		return
	}
}

func main() {
	http.Handle("/ping", rateLimter(endpointHandler))
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Println("failed to start http server: ", err)
	}
}
