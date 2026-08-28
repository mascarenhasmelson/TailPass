package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mascarenhasmelson/TailPass/auth"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Router struct {
	ctx          context.Context
	pool         *pgxpool.Pool
	mux          *http.ServeMux
	jwtSecret    []byte
	cookieSecure bool
	// authLimiter throttles /auth/login and /auth/register per source IP to
	// slow down brute-force / credential-stuffing attempts.
	authLimiter *auth.IPRateLimiter
}

// NewRouter builds the HTTP mux. jwtSecret signs/verifies access tokens;
// cookieSecure controls whether the refresh-token cookie requires HTTPS
// (set COOKIE_SECURE=true once TailPass is served over TLS).
func NewRouter(ctx context.Context, pool *pgxpool.Pool, jwtSecret []byte, cookieSecure bool) *http.ServeMux {
	r := &Router{
		ctx:          ctx,
		pool:         pool,
		mux:          http.NewServeMux(),
		jwtSecret:    jwtSecret,
		cookieSecure: cookieSecure,
		authLimiter:  auth.NewIPRateLimiter(5, time.Minute),
	}

	r.routes()
	return r.mux
}

func (r *Router) routes() {
	// Public auth endpoints - not behind RequireAuth by definition.
	r.mux.HandleFunc("/auth/status", r.authStatusHandler)
	r.mux.HandleFunc("/auth/register", r.authRegisterHandler)
	r.mux.HandleFunc("/auth/login", r.authLoginHandler)
	r.mux.HandleFunc("/auth/refresh", r.authRefreshHandler)
	r.mux.HandleFunc("/auth/logout", r.authLogoutHandler)

	// Everything under /services exposes internal network reachability
	// info and can start/stop tunnels, so all of it requires a valid
	// access token.
	r.mux.HandleFunc("/services", auth.RequireAuth(r.jwtSecret, r.servicesHandler))
	r.mux.HandleFunc("/services/", auth.RequireAuth(r.jwtSecret, r.serviceHandler))
	r.mux.HandleFunc("/services/isp", auth.RequireAuth(r.jwtSecret, r.ispHandler))
	r.mux.HandleFunc("/services/tailscale-ip", auth.RequireAuth(r.jwtSecret, r.tailscaleIPHandler))
	r.mux.HandleFunc("/services/random-port", auth.RequireAuth(r.jwtSecret, r.randomPortHandler))
}

func (r *Router) servicesHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch req.Method {
	case http.MethodGet:
		HandleFetchServices(r.ctx, w, r.pool)
	case http.MethodPost:
		HandleAddService(r.ctx, w, r.pool, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) serviceHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	idStr := strings.TrimPrefix(req.URL.Path, "/services/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	if req.Method == http.MethodDelete {
		HandleDeleteService(r.ctx, w, r.pool, id)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) ispHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch req.Method {
	case http.MethodGet:
		HandleGetISPInfo(w, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) tailscaleIPHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch req.Method {
	case http.MethodGet:
		HandleGetTailscaleIP(w, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) randomPortHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch req.Method {
	case http.MethodGet:
		HandleGetRandomPort(w, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) authStatusHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	HandleAuthStatus(r.ctx, w, r.pool)
}

func (r *Router) authRegisterHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authLimiter.Allow(auth.ClientIP(req)) {
		http.Error(w, "Too many attempts, please wait a moment and try again", http.StatusTooManyRequests)
		return
	}
	HandleRegister(r.ctx, w, r.pool, req, r.jwtSecret, r.cookieSecure)
}

func (r *Router) authLoginHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authLimiter.Allow(auth.ClientIP(req)) {
		http.Error(w, "Too many attempts, please wait a moment and try again", http.StatusTooManyRequests)
		return
	}
	HandleLogin(r.ctx, w, r.pool, req, r.jwtSecret, r.cookieSecure)
}

func (r *Router) authRefreshHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	HandleRefresh(r.ctx, w, r.pool, req, r.jwtSecret, r.cookieSecure)
}

func (r *Router) authLogoutHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	HandleLogout(r.ctx, w, r.pool, req, r.cookieSecure)
}
