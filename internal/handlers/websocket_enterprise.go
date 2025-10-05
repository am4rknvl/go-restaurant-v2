package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type OrderWSHandler struct {
	upgrader websocket.Upgrader
	hub      interface{ Broadcast(v interface{}) }
}

func NewOrderWSHandler(hub interface{ Broadcast(v interface{}) }) *OrderWSHandler {
	return &OrderWSHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		hub: hub,
	}
}

func (h *OrderWSHandler) HandleOrderUpdates(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
		
		// Broadcast order updates
		if msg.Type == "order.status_update" {
			h.hub.Broadcast(msg)
		}
	}
}

func (h *OrderWSHandler) BroadcastOrderUpdate(orderID, status string) {
	msg := WSMessage{
		Type: "order.status_update",
		Data: gin.H{
			"order_id": orderID,
			"status":   status,
		},
	}
	h.hub.Broadcast(msg)
}