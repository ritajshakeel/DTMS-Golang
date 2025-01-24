# Distributed Task Management System (DTMS)

## Overview

DTMS is a system designed to manage distributed tasks with features like user registration, task assignment, real-time notifications, and task status updates. The backend is built using Go with the Gin framework, GORM for ORM, SQLite for data storage, WebSockets for real-time communication, and middleware authentication.

## Project Structure

The project is organized as follows:

```
DTMS/
│-- config/
│   └── database.go
│-- controllers/
│   ├── authController.go
│   ├── controllers_test.go
│   └── taskController.go
│-- middleware/
│   └── authMiddleware.go
│-- models/
│   ├── task.go
│   └── users.go
│-- routes/
│   ├── authRoutes.go
│   ├── taskRoutes.go
│   └── WebSocketsRoutes.go
│-- websocket/
│   └── websocket.go
│-- docker-compose.yml
│-- Dockerfile
│-- go.mod
│-- go.sum
│-- main.go
```

## Prerequisites

Before running the project, ensure you have the following installed:

- [Docker](https://www.docker.com/get-started)
- [Postman](https://www.postman.com/downloads/) (for API testing)

## Instructions to Compile and Run the Project

### Step 1: Open the Project Directory

```sh
cd dtms
```

### Step 2: Build and Run the Project using Docker Compose

```sh
docker compose build
```

```sh
docker compose up -d
```

### Step 3: API Testing

Use Postman to test the available API endpoints. Import the provided Postman collection or manually test the endpoints.

(I have attached file (task-management.postman_collection.json))

## Instructions to Deploy the Project using Docker

### Step 1: Build and Run the Project using Docker Compose

```sh
docker compose build
```

```sh
docker compose up -d
```

### Step 2: Verify the Deployment

Open Postman and test the API by sending requests to:

```
http://localhost:8080/
```

### Step 3: Stop and Remove the Containers

To stop the running containers:

```sh
docker compose down
```

## Unit Testing with Docker

To run unit tests using Docker Compose, execute:

```sh
docker compose run dtms-app go test ./...
```

## Database Configuration

The project supports SQLite. Configure the database in the appropriate Go files within the `config/` directory.

SQLite:

```go
DB, err := gorm.Open(sqlite.Open("./dtms.db"), &gorm.Config{})
```


## Real-Time Communication

The project supports real-time updates via WebSockets. Ensure clients are set up to connect accordingly.

## Authentication

Middleware authentication is implemented to secure endpoints. Ensure that valid tokens are used when accessing protected routes.

