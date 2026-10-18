package utils

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const PasswordHashCost = 12

// HashPassword returns a bcrypt hash suitable for storage.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), PasswordHashCost)
	return string(hashed), err
}

// VerifyPassword reports whether password matches the stored bcrypt hash.
func VerifyPassword(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// HashCode returns a hex SHA-256 of an OTP code. OTPs are short and
// high-entropy-bounded, so a fast hash (not bcrypt) is appropriate; it keeps
// plaintext codes out of the database while allowing constant-time comparison
// via the stored hex digest.
func HashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}
