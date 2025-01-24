FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o main .

FROM golang:1.22

WORKDIR /app

COPY --from=builder /app /app

EXPOSE 8080


CMD ["/app/main"]
