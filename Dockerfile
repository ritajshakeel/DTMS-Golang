# Stage 1: Build the Go application
FROM golang:1.22 AS builder

# Set the working directory
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

# Stage 2: The final image for running the app and tests
FROM golang:1.22

# Set the working directory
WORKDIR /app

# Copy the Go binary and dependencies from the builder stage
COPY --from=builder /app /app

# Expose the port for the app
EXPOSE 8080

# Default command to run the app
CMD ["/app/main"]
