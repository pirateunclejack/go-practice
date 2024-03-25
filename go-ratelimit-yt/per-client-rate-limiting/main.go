package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

func perClientRateLimiter(
	next func(writer http.ResponseWriter, request *http.Request)) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	type ClientStore struct {
		mu      sync.Mutex
		clients map[string]*client
	}

	// var (
	// 	mu sync.Mutex
	// 	clients = make(map[string]*client)
	// )

	var client_store = ClientStore {
		clients: make(map[string]*client),
	}

	go func(client_store *ClientStore) {
		for {
			// time.Sleep(time.Minute)
			client_store.mu.Lock()
			for ip, client := range client_store.clients {
				if time.Since(client.lastSeen) > 5 * time.Second {
					delete(client_store.clients, ip)
				}
			}
			client_store.mu.Unlock()
		}
	}(&client_store)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		client_store.mu.Lock()
		defer client_store.mu.Unlock()
		if _, found := client_store.clients[ip]; !found {
			client_store.clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
		}
		client_store.clients[ip].lastSeen = time.Now()
		// fmt.Println(client_store.clients[ip].lastSeen.String())
		// fmt.Println(client_store.clients[ip].limiter.Allow())
		if !client_store.clients[ip].limiter.Allow() {
			message := Message{
				Status: "Request Failed",
				Body: "The API is at capacity. Please try again later.",
			}
			w.WriteHeader(http.StatusTooManyRequests)
			err := json.NewEncoder(w).Encode(&message)
			if err != nil {
				log.Printf("failed to encode message to response writer: %v", err)
				return
			}
			return
		}
		// execute the next function
		next(w, r)
	})
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
	http.Handle("/ping", perClientRateLimiter(endpointHandler))
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Println("failed to start http server: ", err)
	}
}
