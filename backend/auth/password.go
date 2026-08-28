package auth

import (
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost 12 is a solid default in 2026: expensive enough to slow
// offline brute-force attempts against a stolen password_hash column,
// cheap enough not to bottleneck a single-admin login flow.
const bcryptCost = 12

// HashPassword returns a bcrypt hash of password, safe to store in the DB.
// Passwords are NEVER stored or logged in plaintext anywhere in TailPass.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword reports whether password matches the given bcrypt hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ValidatePasswordStrength enforces a minimum bar for the admin account
// password created during first-run setup.
func ValidatePasswordStrength(password string) error {
	if len(password) < 10 {
		return fmt.Errorf("password must be at least 10 characters long")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return fmt.Errorf("password must include upper and lower case letters, a number, and a symbol")
	}
	return nil
}
