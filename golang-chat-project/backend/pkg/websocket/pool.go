package websocket

import "fmt"

type Pool struct {
	Register	chan *Client
	Unregister  chan *Client
	Clients		map[*Client]bool
	Broadcast	chan Message
}

func NewPool() *Pool {
	return &Pool{
		Register: make(chan *Client),
		Unregister: make(chan *Client),
		Clients: make(map[*Client]bool),
		Broadcast: make(chan Message),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Printf("size of connection pool: %v", len(pool.Clients))
			for client := range pool.Clients {
				fmt.Printf("client: %v", client)
				client.Conn.WriteJSON(Message{Type:1, Body: "New User Joined..."})
			}
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Printf("size of connection pool: %v", len(pool.Clients))
			for client := range pool.Clients {
				fmt.Printf("client: %v", client)
				client.Conn.WriteJSON(Message{Type:1, Body: "User Disconnected..."})
			}
			break
		case message := <- pool.Broadcast:
			fmt.Println("sending message to all clients in the pool")
			for client := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Printf("failed to broadcast json message: %v", err)
					return
				}
			}
		}
		
	}
}
