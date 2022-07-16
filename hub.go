package main

import "log"
import "strings"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run(Logger *log.Logger) {
	for {
		select {
		case client := <-h.register:
			Logger.Println("Register")
			h.clients[client] = true
		case client := <-h.unregister:
			Logger.Println("Unregister")
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			event := string(message)
			arr := strings.SplitN(event, ";", 2)
			organization := arr[0]
			msg := arr[1]
			for client := range h.clients {
				if client.organization == organization {
					select {
					case client.send <- []byte(msg):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}
