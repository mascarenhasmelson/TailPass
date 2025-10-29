# Build Stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o TailPass main.go

# Run Stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/TailPass .
EXPOSE 8082
CMD ["./TailPass"]
