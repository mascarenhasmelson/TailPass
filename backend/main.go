package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"port/api"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	dbMaxConns = 5
	connString = "postgres://admin:StrongPassword123@192.168.30.249:5432/tunnel_services"
)

func main() {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse DATABASE URL: %v\n", err)
		os.Exit(1)
	}
	config.MaxConns = dbMaxConns
	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	fmt.Println("connected to PostgreSQL!")
	http.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		if api.EnableCORS(&w, r) {
    return
}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// w.WriteHeader(http.StatusOK)
		// w.Write([]byte("Request processed successfully"))
		switch r.Method {
		case http.MethodGet:
			api.HandleFetchServices(ctx, w, pool)
		case http.MethodPost:
			api.HandleAddService(ctx, w, pool, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/services/", func(w http.ResponseWriter, r *http.Request) {
		if api.EnableCORS(&w, r) {
			return
		}
		///services/{id}
		idStr := strings.TrimPrefix(r.URL.Path, "/services/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid service ID", http.StatusBadRequest)
			return
		}
		if r.Method == http.MethodDelete {
			api.HandleDeleteService(ctx, w, pool, id)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	fmt.Println("running on http://localhost:8082")
	http.ListenAndServe(":8082", nil)
}
