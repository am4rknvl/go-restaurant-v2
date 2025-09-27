package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn *websocket.Conn
	mu   sync.Mutex
	role string
}

type Hub struct {
	clients   map[*Client]bool
	mu        sync.RWMutex
	broadcast chan interface{}
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*Client]bool),
		broadcast: make(chan interface{}, 256),
	}
}

func (h *Hub) Run() {
	for msg := range h.broadcast {
		h.mu.RLock()
		for c := range h.clients {
			c.send(msg)
		}
		h.mu.RUnlock()
	}
}

func (c *Client) send(v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, _ := json.Marshal(v)
	_ = c.conn.WriteMessage(websocket.TextMessage, data)
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(v interface{}) {
	select {
	case h.broadcast <- v:
	default:
	}
}

// BroadcastToKitchen sends a message only to clients with role kitchen
func (h *Hub) BroadcastToKitchen(v interface{}) {
	h.mu.RLock()
	for c := range h.clients {
		if c.role == "kitchen" {
			c.send(v)
		}
	}
	h.mu.RUnlock()
}

// HandleWebSocket upgrades the connection and registers the client
func HandleWebSocket(h *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	role := r.URL.Query().Get("role")
	client := &Client{conn: conn, role: role}

	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	// Simple read pump to keep connection open; close on error
	go func() {
		defer func() {
			conn.Close()
			h.mu.Lock()
			delete(h.clients, client)
			h.mu.Unlock()
		}()
		for {
			if _, _, err := conn.NextReader(); err != nil {
				return
			}
		}
	}()
}
