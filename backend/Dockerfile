FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY . .
RUN go mod download
RUN go build -mod=mod -o TailPass ./cmd/main.go
RUN git clone https://github.com/mascarenhasmelson/tcp-tunnel-go.git /tcp-tunnel-go \
    && cd /tcp-tunnel-go \
    && go mod tidy \
    && go build -o /tcp
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/TailPass .
COPY --from=builder /tcp ./tcp
RUN chmod +x TailPass tcp
EXPOSE 8082
CMD ["./TailPass"]
