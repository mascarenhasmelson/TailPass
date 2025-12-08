package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"port/api"
	"port/servicetools"

	"github.com/jackc/pgx/v4/pgxpool"
)

const dbMaxConns = 5

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		fmt.Println("Kill...")
		cancel()
	}()
	connString := "postgres://admin:StrongPassword123@localhost:5432/tunnel_services"
	//  connString = os.Getenv("DATABASE_URL")

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Unable to parse DATABASE URL: %v", err)
	}
	config.MaxConns = dbMaxConns
	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("connected to PostgreSQL!")
	go servicetools.StartPortMonitor(ctx, pool)
	server := &http.Server{
		Addr:    ":8082",
		Handler: api.NewRouter(ctx, pool),
	}
	go func() {
		<-ctx.Done()
		fmt.Println("Stopping Backend server...")
		server.Shutdown(context.Background())
	}()
	fmt.Println("running on http://localhost:8082")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("HTTP server error:", err)
	}

}
