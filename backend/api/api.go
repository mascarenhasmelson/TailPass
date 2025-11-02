package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"port/utils"
	"sync"
	"os"
  "port/portforward"
   "strconv"
	"github.com/jackc/pgx/v4/pgxpool"
)
var mu sync.Mutex
// var services []utils.Service
func HandleFetchServices(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool) {
	rows, err := pool.Query(ctx, `
		SELECT id, service_name::text,
	       host(local_ip) AS local_ip,
	       local_port,
	       host(remote_ip) AS remote_ip,
	       remote_port
	FROM services
	`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []utils.Service
	for rows.Next() {
		var s utils.Service
		if err := rows.Scan(&s.ID, &s.Service_name, &s.LocalIP, &s.LocalPort, &s.RemoteIP, &s.RemotePort); err != nil {
			http.Error(w, fmt.Sprintf("Row scan failed: %v", err), http.StatusInternalServerError)
			return
		}
		services = append(services, s)
	}
	if rows.Err() != nil {
		http.Error(w, fmt.Sprintf("Rows iteration error: %v", rows.Err()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}
//delete service one at a time
func HandleDeleteService(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool, id int) {
	mu.Lock()
	defer mu.Unlock()
		var pid int
		err := pool.QueryRow(ctx, `SELECT pid FROM services WHERE id = $1`, id).Scan(&pid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch PID for service %d: %v", id, err), http.StatusInternalServerError)
		return
	}
	cmd, err := pool.Exec(ctx, `DELETE FROM services WHERE id = $1`, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete service: %v", err), http.StatusInternalServerError)
		return
	}

	if cmd.RowsAffected() == 0 {
		http.Error(w, "No record found with that ID", http.StatusNotFound)
		return
	}
	proc, err := os.FindProcess(pid)
		if err != nil {
			fmt.Fprintf(w, "Service deleted, but failed to find process (PID %d): %v\n", pid, err)
			return
		}
			err = proc.Kill()
			if err != nil {
				fmt.Fprintf(w, "Service deleted, but failed to kill process (PID %d): %v\n", pid, err)
				return
			}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Service with ID %d deleted successfully", id)))
}
//add portforward
func HandleAddService(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	var s utils.Service
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	pid, err := portforward.PortForward(ctx,s.LocalIP, strconv.Itoa(s.LocalPort), s.RemoteIP, strconv.Itoa(s.RemotePort))
if err != nil {
	http.Error(w, fmt.Sprintf("Failed to start port forward: %v", err), http.StatusInternalServerError)
	return
}
s.PID = pid
	query := `
		INSERT INTO services (service_name, local_ip, local_port, remote_ip, remote_port,pid)
		VALUES ($1, $2, $3, $4, $5,$6)
		RETURNING id;
	`
	err = pool.QueryRow(ctx, query, s.Service_name, s.LocalIP, s.LocalPort, s.RemoteIP, s.RemotePort, s.PID).Scan(&s.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database insert failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}
//cors
func EnableCORS(w *http.ResponseWriter, r *http.Request) bool {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
	if r.Method == http.MethodOptions {
		(*w).WriteHeader(http.StatusOK)
		   // (*w).WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}
