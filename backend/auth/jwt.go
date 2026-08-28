package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Claims are the JWT payload fields TailPass issues. Kept intentionally
// minimal: just enough to identify the user and enforce expiry.
type Claims struct {
	UserID    int    `json:"uid"`
	Username  string `json:"username"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// ErrInvalidToken covers every verification failure (bad signature,
// malformed structure, expired) without distinguishing which, so callers
// can't use error text to probe internals.
var ErrInvalidToken = errors.New("invalid or expired token")

func b64Encode(b []byte) string          { return base64.RawURLEncoding.EncodeToString(b) }
func b64Decode(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }

func sign(secret []byte, data string) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// GenerateAccessToken creates a signed HS256 JWT valid for ttl.
func GenerateAccessToken(secret []byte, userID int, username string, ttl time.Duration) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Username:  username,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(ttl).Unix(),
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	unsigned := b64Encode(headerJSON) + "." + b64Encode(claimsJSON)
	signature := sign(secret, unsigned)
	return unsigned + "." + b64Encode(signature), nil
}

// ParseAccessToken verifies the signature (constant-time comparison) and
// expiry of token, returning its claims if valid.
func ParseAccessToken(secret []byte, token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSig := sign(secret, unsigned)

	gotSig, err := b64Decode(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !hmac.Equal(expectedSig, gotSig) {
		return nil, ErrInvalidToken
	}

	claimsJSON, err := b64Decode(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}
