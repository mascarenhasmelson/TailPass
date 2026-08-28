FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY . .
RUN go mod download
RUN go build -mod=mod -o TailPass ./cmd/main.go

FROM alpine:latest
WORKDIR /app
# The 'tailscale' CLI is optional: if present on PATH, TailPass uses it to
# resolve this host's Tailscale IP; otherwise it falls back to scanning
# network interfaces for a 100.64.0.0/10 address itself.
COPY --from=builder /app/TailPass .
RUN chmod +x TailPass
EXPOSE 8082
CMD ["./TailPass"]
