package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/mascarenhasmelson/TailPass/servicetools"
	"github.com/mascarenhasmelson/TailPass/utils"

	"github.com/jackc/pgx/v4/pgxpool"
)

var mu sync.Mutex

func HandleFetchServices(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool) {
	rows, err := pool.Query(ctx, `
		SELECT id, 
		service_name::text,
	       host(local_ip) AS local_ip,
	       local_port,
	       host(remote_ip) AS remote_ip,
	       remote_port,
		   online,
		last_seen
	FROM services 
	ORDER BY id ASC
	`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []utils.Service
	for rows.Next() {
		var s utils.Service
		if err := rows.Scan(
			&s.ID,
			&s.Service_name,
			&s.LocalIP,
			&s.LocalPort,
			&s.RemoteIP,
			&s.RemotePort,
			&s.Online,
			&s.Lastseen,
		); err != nil {
			http.Error(w, fmt.Sprintf("Row scan failed: %v", err), http.StatusInternalServerError)
			return
		}
		// Reflect the real in-process tunnel state rather than a stale DB flag.
		s.Running = servicetools.IsRunning(s.ID)
		services = append(services, s)
	}
	if rows.Err() != nil {
		http.Error(w, fmt.Sprintf("Rows iteration error: %v", rows.Err()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// delete service one at a time
func HandleDeleteService(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool, id int) {
	mu.Lock()
	defer mu.Unlock()

	cmd, err := pool.Exec(ctx, `DELETE FROM services WHERE id = $1`, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete service: %v", err), http.StatusInternalServerError)
		return
	}
	if cmd.RowsAffected() == 0 {
		http.Error(w, "No record found with that ID", http.StatusNotFound)
		return
	}

	// Tear down the in-process tunnel. Not fatal if it's already gone (e.g.
	// after a backend restart it may not have been restored yet).
	if err := servicetools.StopTunnel(id); err != nil {
		fmt.Printf("Service %d deleted, tunnel stop note: %v\n", id, err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Service with ID %d deleted successfully", id)))
}

// add portforward
func HandleAddService(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var s utils.Service
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(s.Service_name) == "" {
		http.Error(w, "service_name is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(s.RemoteIP) == "" || s.RemotePort <= 0 || s.RemotePort > 65535 {
		http.Error(w, "remote_ip and a valid remote_port are required", http.StatusBadRequest)
		return
	}

	// TailPass always binds to the Tailscale interface. If the caller didn't
	// pin a specific local IP, auto-detect this host's Tailscale IP instead
	// of asking the user to know/type it.
	if strings.TrimSpace(s.LocalIP) == "" {
		tsIP, err := servicetools.GetTailscaleIP()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to auto-detect Tailscale IP: %v", err), http.StatusInternalServerError)
			return
		}
		s.LocalIP = tsIP
	}

	// If no local port was supplied, generate a free random one on that IP
	// instead of requiring the user to pick one manually.
	if s.LocalPort <= 0 {
		port, err := servicetools.GetFreePort(s.LocalIP)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to allocate a local port: %v", err), http.StatusInternalServerError)
			return
		}
		s.LocalPort = port
	}
	if s.LocalPort > 65535 {
		http.Error(w, "local_port must be between 1 and 65535", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO services (service_name, local_ip, local_port, remote_ip, remote_port, pid)
		VALUES ($1, $2, $3, $4, $5, 0)
		RETURNING id;
	`
	if err := pool.QueryRow(ctx, query, s.Service_name, s.LocalIP, s.LocalPort, s.RemoteIP, s.RemotePort).Scan(&s.ID); err != nil {
		http.Error(w, fmt.Sprintf("Database insert failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Start the in-process TCP tunnel (no external binary involved).
	if err := servicetools.StartTunnel(ctx, s.ID, s.LocalIP, strconv.Itoa(s.LocalPort), s.RemoteIP, strconv.Itoa(s.RemotePort)); err != nil {
		_, _ = pool.Exec(ctx, `DELETE FROM services WHERE id = $1`, s.ID)
		http.Error(w, fmt.Sprintf("Failed to start tunnel: %v", err), http.StatusInternalServerError)
		return
	}

	s.Running = true
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

// HandleGetTailscaleIP reports this host's auto-detected Tailscale IP so the
// frontend never has to ask the user to type it in.
func HandleGetTailscaleIP(w http.ResponseWriter, r *http.Request) {
	ip, err := servicetools.GetTailscaleIP()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.Error{Message: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"ip": ip})
}

// HandleGetRandomPort returns a free, currently-unused local port on the
// given (or Tailscale, by default) IP, for the frontend's "randomize" action.
func HandleGetRandomPort(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimSpace(r.URL.Query().Get("local_ip"))
	if ip == "" {
		if tsIP, err := servicetools.GetTailscaleIP(); err == nil {
			ip = tsIP
		}
	}

	port, err := servicetools.GetFreePort(ip)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.Error{Message: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]int{"port": port})
}

// cors
//
// Reflects the request's Origin (rather than "*") and enables credentials so
// the browser will send/receive the httpOnly refresh-token cookie used by
// /auth/refresh. TailPass is a personal/self-hosted dashboard typically
// reached at a Tailscale IP that can vary, so a fixed allow-list origin
// isn't practical here; the SameSite=Lax + httpOnly refresh cookie, short
// access-token lifetime, and token rotation are the primary protections
// against token theft/replay.
func EnableCORS(w *http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin != "" {
		(*w).Header().Set("Access-Control-Allow-Origin", origin)
		(*w).Header().Set("Access-Control-Allow-Credentials", "true")
		(*w).Header().Set("Vary", "Origin")
	} else {
		(*w).Header().Set("Access-Control-Allow-Origin", "*")
	}
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
	if r.Method == http.MethodOptions {
		(*w).WriteHeader(http.StatusOK)
		return true
	}
	return false
}
