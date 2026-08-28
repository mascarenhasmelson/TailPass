package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// RefreshTokenTTL controls how long a refresh session lasts before the user
// must log in again with their password.
const RefreshTokenTTL = 7 * 24 * time.Hour

// GenerateRefreshToken returns a cryptographically random opaque token. This
// is the value handed to the browser (in an httpOnly cookie) - only its
// SHA-256 hash is ever persisted, so a DB leak alone can't be replayed.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// StoreRefreshToken persists the hash of a newly issued refresh token.
func StoreRefreshToken(ctx context.Context, pool *pgxpool.Pool, userID int, token string) error {
	expiresAt := time.Now().Add(RefreshTokenTTL)
	_, err := pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, hashToken(token), expiresAt)
	return err
}

// ValidateAndRotateRefreshToken looks up the token, deletes it (single-use -
// rotation on every refresh limits the blast radius of a stolen cookie), and
// returns the owning user id if it was valid and unexpired.
func ValidateAndRotateRefreshToken(ctx context.Context, pool *pgxpool.Pool, token string) (int, error) {
	h := hashToken(token)

	var userID int
	var expiresAt time.Time
	err := pool.QueryRow(ctx,
		`DELETE FROM refresh_tokens WHERE token_hash = $1 RETURNING user_id, expires_at`,
		h).Scan(&userID, &expiresAt)
	if err != nil {
		return 0, errors.New("invalid refresh token")
	}
	if time.Now().After(expiresAt) {
		return 0, errors.New("refresh token expired")
	}
	return userID, nil
}

// RevokeRefreshToken deletes a single refresh token (used on logout).
func RevokeRefreshToken(ctx context.Context, pool *pgxpool.Pool, token string) error {
	_, err := pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, hashToken(token))
	return err
}

// PruneExpiredRefreshTokens removes stale rows so the table doesn't grow
// unbounded from abandoned sessions.
func PruneExpiredRefreshTokens(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	return err
}
