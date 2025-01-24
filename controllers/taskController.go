package controllers

import (
	"dtms/config"
	"dtms/models"
	"dtms/websocket"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateTask(c *gin.Context) {
	var task models.Task

	// Bind JSON input to task struct
	if err := c.ShouldBindJSON(&task); err != nil {
		// Log the error details for better debugging
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Save the task to the database
	if err := config.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Send notification to WebSocket clients
	websocket.GetManager().SendNotification("task_created", task)

	// Respond with the created task
	c.JSON(http.StatusOK, gin.H{"message": "Task created successfully", "task": task})
}

func CreateTaskBulk(c *gin.Context) {

	// Using dummy data to test
	// file, openErr := os.Open("data.csv")
	file_ptr, getErr := c.FormFile("taskBulkUpload")

	if getErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get file",
		})
		return
	}

	file, openErr := file_ptr.Open()
	if openErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to open file",
		})
		return
	}

	defer file.Close()

	reader := csv.NewReader(file)

	var tasks []models.Task

	// Skip the header row
	reader.Read()

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error reading CSV",
			})
			return
		}

		// Parse the row data
		layout := "2006-01-02 15:04:05"
		plannedStartTime, err := time.Parse(layout, row[2]+" "+row[3])

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error parsing planned start time",
			})
			return
		}

		plannedEndTime, err := time.Parse(layout, row[4]+" "+row[5])

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error parsing planned end time",
			})
			return
		}

		// String to int64
		seconds, err := strconv.ParseInt(row[6], 10, 64)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error parsing seconds",
			})
			return
		}

		task := models.Task{
			Title:            row[0],
			Description:      row[1],
			PlannedStartTime: plannedStartTime,
			PlannedEndTime:   plannedEndTime,
			Seconds:          seconds,
		}

		tasks = append(tasks, task)
	}

	// Bulk insert the tasks into the database
	insertErr := config.DB.Create(&tasks).Error
	if insertErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error inserting tasks into the database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Successfully uploaded %d tasks", len(tasks)),
	})
}

func GetTasks(c *gin.Context) {
	var tasks []models.Task

	// Preload User only if AssignedTo is not null
	if err := config.DB.Preload("User").Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func UpdateTask(c *gin.Context) {
	task_id := c.Query("task_id")

	// Find the user by ID
	var task models.Task
	err := config.DB.First(&task, task_id).Error

	// Send notification to WebSocket clients
	websocket.GetManager().SendNotification("task_updated", task)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	// Get Title and Description
	var body struct {
		Title            string `json:"title"`
		Description      string `json:"description"`
		PlannedStartTime int64  `json:"planned_start_time"`
		PlannedEndTime   int64  `json:"planned_end_time"`
		ActualStartTime  int64  `json:"actual_start_time"`
		ActualEndTime    int64  `json:"actual_end_time"`
		Seconds          int64  `json:"seconds"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Fields are empty",
		})
		return
	}

	if body.Title != "" {
		task.Title = body.Title
	}

	if body.Description != "" {
		task.Description = body.Description
	}

	// Convert Unix timestamps to time.Time
	if body.PlannedStartTime != 0 {
		task.PlannedStartTime = time.Unix(body.PlannedStartTime, 0)
	}

	if body.PlannedEndTime != 0 {
		task.PlannedEndTime = time.Unix(body.PlannedEndTime, 0)
	}

	if body.ActualStartTime != 0 {
		task.ActualStartTime = time.Unix(body.ActualStartTime, 0)
	}

	if body.ActualEndTime != 0 {
		task.ActualEndTime = time.Unix(body.ActualEndTime, 0)
	}

	if body.Seconds != 0 {
		task.Seconds = body.Seconds
	}

	// Validate that the planned start time is before the planned end time
	if !task.PlannedStartTime.Before(task.PlannedEndTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"Invalid Start Time: cannot be before End Time. Planned Start Time: %s, Planned End Time: %s",
				task.PlannedStartTime.Format(time.RFC3339),
				task.PlannedEndTime.Format(time.RFC3339),
			),
		})
		return
	}

	// Validate that the actual start time is before the actual end time
	if !task.ActualStartTime.Before(task.ActualEndTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"Invalid Actual Start Time: cannot be before Actual End Time. Actual Start Time: %s, Actual End Time: %s",
				task.ActualStartTime.Format(time.RFC3339),
				task.ActualEndTime.Format(time.RFC3339),
			),
		})
		return
	}

	err = config.DB.Save(&task).Error

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error updating task",
		})
		return
	}

	// respond
	c.JSON(http.StatusOK, gin.H{
		"message": "Details added successfully",
		"task":    task,
	})
}

func DeleteTask(c *gin.Context) {
	task_id := c.Query("task_id")

	// Find the user by ID
	var task models.Task
	err := config.DB.First(&task, task_id).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	// Delete the task
	result := config.DB.Delete(&models.Task{}, task_id)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Error deleting task",
			"details": result.Error.Error(), // Include error details
		})
		return
	}

	// respond
	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}
func AssignTask(c *gin.Context) {
	var task models.Task
	var assignData struct {
		UserID uint `json:"user_id"`
		TaskID int  `json:"task_id"`
	}

	// Bind JSON input to assignData
	if err := c.ShouldBindJSON(&assignData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Find the task and preload the associated user
	if err := config.DB.Preload("User").First(&task, assignData.TaskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Assign the task to the user
	task.AssignedTo = &assignData.UserID

	// Save the updated task
	if err := config.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign task"})
		return
	}

	// Fetch the user details to include in the response
	var user models.User
	if err := config.DB.First(&user, assignData.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user details"})
		return
	}
	task.User = &user

	// Respond with the updated task
	c.JSON(http.StatusOK, gin.H{"message": "Task assigned successfully", "task": task})
}
