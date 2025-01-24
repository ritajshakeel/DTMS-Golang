package main

import (
	"dtms/config"
	"dtms/controllers"
	"dtms/routes"
	"dtms/websocket"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	config.ConnectDatabase()
	r := gin.Default()

	routes.SetupAuthRoutes(r)
	routes.SetupTaskRoutes(r)

	websocket.InitWebSocketManager()

	r.POST("/tasks", controllers.CreateTask)
	r.GET("/ws", websocket.HandleConnections)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}
