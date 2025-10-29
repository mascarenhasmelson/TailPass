FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for cloning external repo
RUN apk add --no-cache git

# Copy backend source code
COPY . .

# Download Go dependencies
RUN go mod download

# Build backend binary
RUN go build -o TailPass main.go

# Clone and build tcp-tunnel binary from GitHub
RUN git clone https://github.com/mascarenhasmelson/tcp-tunnel-go.git /tcp-tunnel-go \
    && cd /tcp-tunnel-go \
    && go mod tidy \
    && go build -o /tcp

# ================================
# 🚀 Runtime Stage
# ================================
FROM alpine:latest

WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/TailPass .
COPY --from=builder /tcp ./tcp

# Make sure both are executable
RUN chmod +x TailPass tcp

# Expose backend API port
EXPOSE 8082

# Start backend
CMD ["./TailPass"]
