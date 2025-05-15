# Multi-stage Dockerfile for Go backend
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY ./backend/go.mod ./backend/go.sum ./
RUN go mod download
COPY ./backend .
RUN go build -o url-shortener main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/url-shortener .
EXPOSE 8080
CMD ["./url-shortener"]