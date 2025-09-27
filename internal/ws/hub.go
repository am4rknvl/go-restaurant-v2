package ws

import (
	"sync"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*Client]struct{})}
}

func (h *Hub) Subscribe(topic string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[topic] == nil {
		h.clients[topic] = make(map[*Client]struct{})
	}
	h.clients[topic][c] = struct{}{}
}

func (h *Hub) Unsubscribe(topic string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m, ok := h.clients[topic]; ok {
		delete(m, c)
		if len(m) == 0 {
			delete(h.clients, topic)
		}
	}
}

func (h *Hub) Broadcast(topic string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[topic] {
		select {
		case c.send <- data:
		default:
		}
	}
}
