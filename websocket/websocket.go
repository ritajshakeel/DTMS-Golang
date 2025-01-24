package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins; restrict this in production
	},
}

// WebSocketManager manages WebSocket connections and broadcasts messages to clients
type WebSocketManager struct {
	clients map[*websocket.Conn]bool // Holds the WebSocket connections
	mu      sync.Mutex               // Protects concurrent access to the clients map
}

// Global WebSocket manager instance
var manager *WebSocketManager

// InitWebSocketManager initializes the WebSocket manager
func InitWebSocketManager() {
	manager = &WebSocketManager{
		clients: make(map[*websocket.Conn]bool),
	}
}

// GetManager returns the global WebSocket manager instance
func GetManager() *WebSocketManager {
	return manager
}

// HandleConnections upgrades the HTTP request to a WebSocket connection
func HandleConnections(c *gin.Context) {
	w := c.Writer
	r := c.Request

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	manager.AddClient(conn)
	defer manager.RemoveClient(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// AddClient adds a WebSocket client to the manager
func (m *WebSocketManager) AddClient(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[conn] = true
}

// RemoveClient removes a WebSocket client from the manager
func (m *WebSocketManager) RemoveClient(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, conn)
}

// Broadcast sends a message to all connected WebSocket clients
func (m *WebSocketManager) Broadcast(message []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for client := range m.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			client.Close()
			delete(m.clients, client)
		}
	}
}

// SendNotification sends a structured notification to all clients
func (m *WebSocketManager) SendNotification(event string, data interface{}) {
	message := map[string]interface{}{
		"event": event,
		"data":  data,
	}
	msgBytes, _ := json.Marshal(message) // Serialize to JSON
	m.Broadcast(msgBytes)
}
