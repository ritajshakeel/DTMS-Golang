package routes

import (
	"dtms/websocket"

	"github.com/gin-gonic/gin"
)

func WebSocketRoutes(r *gin.Engine) {
	r.GET("/ws", websocket.HandleConnections)
}
