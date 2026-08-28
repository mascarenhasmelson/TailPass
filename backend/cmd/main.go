package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/mascarenhasmelson/TailPass/api"
	"github.com/mascarenhasmelson/TailPass/auth"
	"github.com/mascarenhasmelson/TailPass/servicetools"

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
	//connString := "postgres://admin:StrongPassword123@localhost:5432/tunnel_services"
	connString := os.Getenv("DATABASE_URL")

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

	if err := ensureAuthTables(ctx, pool); err != nil {
		log.Fatalf("Failed to prepare auth tables: %v", err)
	}

	jwtSecret := loadJWTSecret()
	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"

	restoreTunnels(ctx, pool)

	go servicetools.StartPortMonitor(ctx, pool)
	go pruneExpiredSessions(ctx, pool)

	server := &http.Server{
		Addr:    ":8082",
		Handler: api.NewRouter(ctx, pool, jwtSecret, cookieSecure),
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

// ensureAuthTables idempotently creates the users/refresh_tokens tables so
// existing TailPass deployments gain authentication on upgrade without
// requiring the Postgres volume to be wiped and re-initialized from
// db/schema.sql (which only runs once, against an empty volume).
func ensureAuthTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(64) NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

// loadJWTSecret reads JWT_SECRET from the environment. If it's unset, a
// random secret is generated for this process only - sessions won't survive
// a restart, but the server still starts securely rather than falling back
// to a hardcoded/predictable key.
func loadJWTSecret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	fmt.Println("WARNING: JWT_SECRET is not set. Generating a random, ephemeral secret for this run only - " +
		"all sessions will be invalidated on restart. Set JWT_SECRET in your environment for persistent sessions.")
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		log.Fatalf("Failed to generate JWT secret: %v", err)
	}
	return secret
}

// pruneExpiredSessions periodically clears out stale refresh tokens so the
// table doesn't grow unbounded from abandoned/expired sessions.
func pruneExpiredSessions(ctx context.Context, pool *pgxpool.Pool) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := auth.PruneExpiredRefreshTokens(ctx, pool); err != nil {
				fmt.Println("pruneExpiredSessions:", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// restoreTunnels re-starts the in-process TCP tunnel for every service
// already stored in the database. Since tunnels now live entirely inside
// this Go process (rather than as separate OS processes), a backend restart
// would otherwise leave every saved service un-forwarded until the user
// re-added it. This restores previous behavior automatically.
func restoreTunnels(ctx context.Context, pool *pgxpool.Pool) {
	rows, err := pool.Query(ctx, `
		SELECT id, host(local_ip), local_port, host(remote_ip), remote_port
		FROM services
	`)
	if err != nil {
		fmt.Println("restoreTunnels: query failed:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, localPort, remotePort int
		var localIP, remoteIP string
		if err := rows.Scan(&id, &localIP, &localPort, &remoteIP, &remotePort); err != nil {
			fmt.Println("restoreTunnels: scan failed:", err)
			continue
		}
		if err := servicetools.StartTunnel(ctx, id,
			localIP, strconv.Itoa(localPort),
			remoteIP, strconv.Itoa(remotePort),
		); err != nil {
			fmt.Printf("restoreTunnels: service %d: %v\n", id, err)
		}
	}
}
