package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mascarenhasmelson/TailPass/auth"

	"github.com/jackc/pgx/v4/pgxpool"
)

// accessTokenTTL is intentionally short: a stolen/leaked access token (which
// only ever lives in browser JS memory, never localStorage) is useless
// after this window even without the user taking any action.
const accessTokenTTL = 15 * time.Minute

// refreshCookieName is the httpOnly cookie carrying the (single-use,
// rotated) refresh token. It's scoped to /auth so it's never sent to any
// other endpoint.
const refreshCookieName = "tailpass_refresh"

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleAuthStatus tells the frontend whether an admin account exists yet,
// so it knows whether to show the first-run setup screen or the login form.
func HandleAuthStatus(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool) {
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		http.Error(w, fmt.Sprintf("Failed to check setup status: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"setup_required": count == 0})
}

// HandleRegister creates the single admin account. It only ever succeeds
// once - as soon as a user exists, registration is permanently closed, so
// this endpoint can't be abused to create arbitrary accounts later.
func HandleRegister(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool, r *http.Request, secret []byte, cookieSecure bool) {
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		http.Error(w, "Failed to check existing accounts", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Setup already complete: registration is closed", http.StatusForbidden)
		return
	}

	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	creds.Username = strings.TrimSpace(creds.Username)

	if len(creds.Username) < 3 || len(creds.Username) > 64 {
		http.Error(w, "Username must be between 3 and 64 characters", http.StatusBadRequest)
		return
	}
	if err := auth.ValidatePasswordStrength(creds.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(creds.Password)
	if err != nil {
		http.Error(w, "Failed to secure password", http.StatusInternalServerError)
		return
	}

	var userID int
	err = pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`,
		creds.Username, hash,
	).Scan(&userID)
	if err != nil {
		http.Error(w, "Username already taken or database error", http.StatusConflict)
		return
	}

	issueSession(ctx, w, pool, userID, creds.Username, secret, cookieSecure)
}

// HandleLogin validates credentials and, on success, issues a fresh
// access/refresh token pair. It deliberately returns the same generic error
// for "unknown username" and "wrong password" so the response can't be used
// to enumerate valid usernames.
func HandleLogin(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool, r *http.Request, secret []byte, cookieSecure bool) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	creds.Username = strings.TrimSpace(creds.Username)

	var userID int
	var hash string
	err := pool.QueryRow(ctx,
		`SELECT id, password_hash FROM users WHERE username = $1`, creds.Username,
	).Scan(&userID, &hash)
	if err != nil || !auth.CheckPassword(hash, creds.Password) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	issueSession(ctx, w, pool, userID, creds.Username, secret, cookieSecure)
}

// HandleRefresh exchanges a valid refresh cookie for a new access token,
// rotating the refresh token in the process (old one is deleted, a new one
// issued) so a captured cookie can only ever be replayed once before the
// legitimate user's next refresh invalidates it.
func HandleRefresh(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool, r *http.Request, secret []byte, cookieSecure bool) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		http.Error(w, "No active session", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateAndRotateRefreshToken(ctx, pool, cookie.Value)
	if err != nil {
		clearRefreshCookie(w, cookieSecure)
		http.Error(w, "Session expired, please log in again", http.StatusUnauthorized)
		return
	}

	var username string
	if err := pool.QueryRow(ctx, `SELECT username FROM users WHERE id = $1`, userID).Scan(&username); err != nil {
		clearRefreshCookie(w, cookieSecure)
		http.Error(w, "Account no longer exists", http.StatusUnauthorized)
		return
	}

	issueSession(ctx, w, pool, userID, username, secret, cookieSecure)
}

// HandleLogout revokes the current refresh token (server-side) and clears
// the cookie (client-side).
func HandleLogout(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool, r *http.Request, cookieSecure bool) {
	if cookie, err := r.Cookie(refreshCookieName); err == nil {
		_ = auth.RevokeRefreshToken(ctx, pool, cookie.Value)
	}
	clearRefreshCookie(w, cookieSecure)
	w.WriteHeader(http.StatusOK)
}

func issueSession(ctx context.Context, w http.ResponseWriter, pool *pgxpool.Pool, userID int, username string, secret []byte, cookieSecure bool) {
	accessToken, err := auth.GenerateAccessToken(secret, userID, username, accessTokenTTL)
	if err != nil {
		http.Error(w, "Failed to issue access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Failed to issue refresh token", http.StatusInternalServerError)
		return
	}
	if err := auth.StoreRefreshToken(ctx, pool, userID, refreshToken); err != nil {
		http.Error(w, "Failed to persist session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   cookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(auth.RefreshTokenTTL),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": accessToken,
		"username":     username,
		"expires_in":   int(accessTokenTTL.Seconds()),
	})
}

func clearRefreshCookie(w http.ResponseWriter, cookieSecure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		Secure:   cookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}
