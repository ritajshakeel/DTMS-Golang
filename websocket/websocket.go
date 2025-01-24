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

type WebSocketManager struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

var manager *WebSocketManager

func InitWebSocketManager() {
	manager = &WebSocketManager{
		clients: make(map[*websocket.Conn]bool),
	}
}

func GetManager() *WebSocketManager {
	return manager
}

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

func (m *WebSocketManager) AddClient(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[conn] = true
}

func (m *WebSocketManager) RemoveClient(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, conn)
}

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

func (m *WebSocketManager) SendNotification(event string, data interface{}) {
	message := map[string]interface{}{
		"event": event,
		"data":  data,
	}
	msgBytes, _ := json.Marshal(message)
	m.Broadcast(msgBytes)
}
