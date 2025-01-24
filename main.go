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

	// Initialize the database connection
	config.ConnectDatabase()

	// Create a new Gin router
	r := gin.Default()

	// Setup routes
	routes.SetupAuthRoutes(r) // Setup authentication-related routes
	routes.SetupTaskRoutes(r) // Setup task-related routes

	// Initialize WebSocket manager
	websocket.InitWebSocketManager()

	// Define additional routes
	r.POST("/tasks", controllers.CreateTask)
	r.GET("/ws", websocket.HandleConnections)

	// Start the server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}
