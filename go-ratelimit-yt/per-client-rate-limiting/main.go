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
		limiter *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu *sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 1 * time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(2,4)}
		}
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			message := Message{
				Status: "success",
				Body: "Hello, world!",
			}
			err := json.NewEncoder(w).Encode(&message)
			if err != nil {
				log.Printf("failed to encode message to response writer: %v", err)
				return
			}
		}
		mu.Unlock()
		// execute the next function
		next(w, r)
	})
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
	http.Handle("/ping", perClientRateLimiter(endpointHandler))
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Println("failed to start http server: ", err)
	}
}
