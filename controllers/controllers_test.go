package controllers

import (
	"bytes"
	"dtms/config"
	"dtms/models"
	"dtms/websocket"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setup() {
	gin.SetMode(gin.TestMode)

	websocket.InitWebSocketManager()

	config.ConnectDatabase()

	if err := config.DB.AutoMigrate(&models.User{}, &models.Task{}); err != nil {
		panic(fmt.Sprintf("Failed to migrate the test database: %v", err))
	}
	config.DB.Exec("DELETE FROM users")
	config.DB.Exec("DELETE FROM tasks")
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.POST("/register", Register)
		auth.POST("/login", Login)
	}

	tasks := r.Group("/task")
	{
		tasks.POST("/create", CreateTask)
		tasks.POST("/bulkupload", CreateTaskBulk)
		tasks.GET("/", GetTasks)
		tasks.PUT("/update", UpdateTask)
		tasks.PUT("/assign", AssignTask)
		tasks.DELETE("/delete", DeleteTask)
	}
	return r
}

func RegisterUserForTest() models.User {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("Failed to hash password: %v", err))
	}

	user := models.User{
		Email:    "testuser@example.com",
		Password: string(hashedPassword),
		Username: "testuser",
	}

	result := config.DB.Create(&user)
	if result.Error != nil {
		panic(fmt.Sprintf("Error creating user for test: %v", result.Error))
	}

	return user
}

func TestRegister(t *testing.T) {
	setup()
	r := setupRouter()

	t.Run("Valid Registration", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":            "newuser@example.com",
			"password":         "Password123",
			"confirm_password": "Password123",
			"username":         "newuser",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP 200 for valid registration")
		assert.Contains(t, w.Body.String(), "User registered successfully!")
	})
}

func TestLogin(t *testing.T) {
	setup()
	router := setupRouter()

	t.Run("Valid Login", func(t *testing.T) {

		_ = RegisterUserForTest()

		payload := map[string]interface{}{
			"email":    "testuser@example.com",
			"password": "Password123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP 200 for valid login")
		assert.Contains(t, w.Body.String(), "Login successful!", "Expected success message in response")
	})
}

func TestCreateTask(t *testing.T) {

	setup()
	router := setupRouter()

	t.Run("Valid Task Creation", func(t *testing.T) {

		payload := map[string]interface{}{
			"title":              "New Task",
			"description":        "This is a new task.",
			"planned_start_time": "2025-01-24T09:00:00Z",
			"planned_end_time":   "2025-01-24T10:00:00Z",
			"seconds":            3600,
		}

		jsonData, err := json.Marshal(payload)
		assert.NoError(t, err, "Failed to marshal payload")

		req, err := http.NewRequest("POST", "/task/create", bytes.NewBuffer(jsonData))
		assert.NoError(t, err, "Failed to create HTTP request")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")
		assert.Contains(t, w.Body.String(), "Task created successfully", "Expected success message")
	})

	t.Run("Invalid Task Data", func(t *testing.T) {

		payload := map[string]interface{}{
			"title":              "Invalid Task",
			"description":        "",
			"planned_start_time": "",
			"planned_end_time":   "",
			"seconds":            0,
		}

		jsonData, err := json.Marshal(payload)
		assert.NoError(t, err, "Failed to marshal payload")

		req, err := http.NewRequest("POST", "/task/create", bytes.NewBuffer(jsonData))
		assert.NoError(t, err, "Failed to create HTTP request")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")
		assert.Contains(t, w.Body.String(), "Invalid input", "Expected error message")
	})
}

func TestUpdateTask(t *testing.T) {
	setup()
	router := setupRouter()

	t.Run("Valid Task Update", func(t *testing.T) {

		task := CreateTestTask()

		updatedStartTime := task.PlannedStartTime.Add(time.Hour)
		updatedEndTime := updatedStartTime.Add(time.Hour * 2)
		updatedActualStartTime := updatedStartTime.Add(time.Minute * 30)
		updatedActualEndTime := updatedEndTime.Add(time.Minute * 30)

		payload := map[string]interface{}{
			"title":              "Updated Task Title",
			"planned_start_time": updatedStartTime.Unix(),
			"planned_end_time":   updatedEndTime.Unix(),
			"actual_start_time":  updatedActualStartTime.Unix(),
			"actual_end_time":    updatedActualEndTime.Unix(),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/task/update?task_id=%d", task.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Details added successfully")
	})

	t.Run("Task Not Found", func(t *testing.T) {

		payload := map[string]interface{}{
			"title": "Non-Existent Task",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/task/update?task_id=999", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Task not found")
	})
}

func TestDeleteTask(t *testing.T) {
	setup()
	router := setupRouter()

	t.Run("Valid Task Deletion", func(t *testing.T) {
		task := CreateTestTask()

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/task/delete?task_id=%d", task.ID), nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Task deleted successfully")
	})

	t.Run("Task Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/task/delete?task_id=999", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Task not found")
	})
}

func TestAssignTask(t *testing.T) {
	setup()
	router := setupRouter()

	t.Run("Valid Task Assignment", func(t *testing.T) {
		task := CreateTestTask()
		user := CreateTestUser()

		payload := map[string]interface{}{
			"user_id": user.ID,
			"task_id": task.ID,
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/task/assign", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Task assigned successfully")
	})

	t.Run("Task Not Found", func(t *testing.T) {
		payload := map[string]interface{}{
			"user_id": 1,
			"task_id": 999,
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/task/assign", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Task not found")
	})
}

func CreateTestTask() *models.Task {

	task := &models.Task{
		Title:            "Test Task",
		Description:      "Test Description",
		PlannedStartTime: time.Now(),
		PlannedEndTime:   time.Now().Add(time.Hour),
		Seconds:          3600,
	}

	result := config.DB.Create(task)

	if task.ID == 0 {
		panic(fmt.Sprintf("Failed to create task: ID is not set"))
	}

	if result.Error != nil {
		panic(fmt.Sprintf("Error creating test task: %v", result.Error))
	}

	return task
}

func CreateTestUser() models.User {
	user := models.User{
		Username: "testuser",
		Password: "testpassword",
		Email:    "testuser@example.com",
	}
	result := config.DB.Create(&user)
	if result.Error != nil {
		panic(fmt.Sprintf("Error creating test user: %v", result.Error))
	}
	return user
}
