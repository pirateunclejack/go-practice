package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/didip/tollbooth/v7"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

func endpointHandler(writer http.ResponseWriter, request *http.Request) {
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
	message := Message{
		Status: "Request Failed",
		Body: "The API is at capacity. Please try again later.",
	}

    jsonMessage, _ := json.Marshal(message)
    tlbthLimiter := tollbooth.NewLimiter(1, nil)
    tlbthLimiter.SetMessageContentType("application/json")
    tlbthLimiter.SetMessage(string(jsonMessage))
    http.Handle("/ping", tollbooth.LimitFuncHandler(tlbthLimiter, endpointHandler))
    err := http.ListenAndServe(":8888", nil)
    if err != nil {
        log.Println("failed to start server on 8888")
    }
}
