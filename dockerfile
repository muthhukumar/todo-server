# Builder stage
FROM golang:1.21-alpine AS builder

ENV CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o todo-server ./cmd/todo-server/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/todo-server .

EXPOSE 8080

CMD ["./todo-server"]
