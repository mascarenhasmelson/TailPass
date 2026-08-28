package auth

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type contextKey string

const userContextKey contextKey = "tailpass.authedUser"

// AuthedUser is the identity extracted from a validated access token.
type AuthedUser struct {
	ID       int
	Username string
}

// RequireAuth wraps a handler so it only runs for requests carrying a valid
// "Authorization: Bearer <token>" header. CORS preflight (OPTIONS) requests
// are passed through untouched since browsers never attach credentials to
// them - the wrapped handler is expected to answer preflights itself (as all
// TailPass handlers do via EnableCORS).
func RequireAuth(secret []byte, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}

		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := ParseAccessToken(secret, token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, AuthedUser{ID: claims.UserID, Username: claims.Username})
		next(w, r.WithContext(ctx))
	}
}

// UserFromContext retrieves the authenticated user set by RequireAuth.
func UserFromContext(ctx context.Context) (AuthedUser, bool) {
	u, ok := ctx.Value(userContextKey).(AuthedUser)
	return u, ok
}

// IPRateLimiter is a simple fixed-window limiter used to slow down
// credential-stuffing / brute-force attempts against /auth/login and
// /auth/register. It's intentionally in-memory: TailPass runs as a single
// instance, so no shared store is needed.
type IPRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*limiterEntry
	max     int
	window  time.Duration
}

type limiterEntry struct {
	count   int
	resetAt time.Time
}

func NewIPRateLimiter(max int, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{entries: make(map[string]*limiterEntry), max: max, window: window}
}

// Allow reports whether another attempt from ip is permitted right now,
// incrementing its counter if so.
func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.entries[ip]
	if !ok || now.After(e.resetAt) {
		l.entries[ip] = &limiterEntry{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if e.count >= l.max {
		return false
	}
	e.count++
	return true
}

// ClientIP extracts the request's remote IP, stripping the port.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
