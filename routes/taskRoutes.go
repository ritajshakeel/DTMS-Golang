package routes

import (
	"dtms/controllers"
	"dtms/middleware"

	"github.com/gin-gonic/gin"
)

func SetupTaskRoutes(r *gin.Engine) {
	tasks := r.Group("/task", middleware.AuthMiddleware())
	{
		tasks.POST("/create", controllers.CreateTask)
		tasks.POST("/bulkupload", controllers.CreateTaskBulk)
		tasks.GET("/", controllers.GetTasks)
		tasks.PUT("/update", controllers.UpdateTask)
		tasks.PUT("/assign", controllers.AssignTask)
		tasks.DELETE("/delete", controllers.DeleteTask)
	}
}
