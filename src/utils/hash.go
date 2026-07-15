package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// GeneratePassword generates a cryptographically secure alphanumeric password.
func GeneratePassword(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	password := make([]byte, length)
	for i := range password {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[n.Int64()]
	}
	return strings.ToLower(string(password))
}

// GenerateHash returns a hex string of `length` characters using crypto/rand.
func GenerateHash(length int) string {
	if length <= 0 {
		length = 8
	}
	b := make([]byte, (length+1)/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

// GenerateBase64Str returns a base64-encoded string from `n` random bytes.
func GenerateBase64Str(bytes int) string {
	if bytes <= 0 {
		bytes = 32
	}
	b := make([]byte, bytes)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// GenerateRandomDomain generates a sslip.io domain.
func GenerateRandomDomain(serverIP, projectName string) string {
	hash := GenerateHash(6)
	slugIP := strings.NewReplacer(".", "-", ":", "-").Replace(serverIP)
	// Domain labels have a max length of 63 characters
	// Reserve space for: hash (6) + separators (1-2) + ip section + dot + sslip.io (8)
	// Approx: 6 + 2 + (variable ip length) + 9 = ~19-30 chars for other parts
	const maxLen = 40
	if len(projectName) > maxLen {
		projectName = projectName[:maxLen]
	}
	if slugIP == "" {
		return fmt.Sprintf("%s-%s.sslip.io", projectName, hash)
	}
	return fmt.Sprintf("%s-%s-%s.sslip.io", projectName, hash, slugIP)
}

// GenerateJWT creates an HS256 JWT for template config embedding.
func GenerateJWT(secret string, payload map[string]any, length int) string {
	// jwt:<length> variant — just random hex, not a real JWT
	if length > 0 {
		b := make([]byte, length)
		_, _ = rand.Read(b)
		return hex.EncodeToString(b)
	}

	// Build claims
	claims := jwt.MapClaims{
		"iss": "goploy",
		"iat": time.Now().Unix(),
		// These are config tokens, not session tokens
		"exp": time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	// Merge caller's payload (overrides defaults except exp/iat)
	for k, v := range payload {
		claims[k] = v
	}

	// Random secret if none provided
	if secret == "" {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		secret = hex.EncodeToString(b)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		// Should never happen with HS256 + valid secret
		return fmt.Sprintf("jwt-error: %v", err)
	}
	return signed
}

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if a plain-text password matches a bcrypt hash.
func VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
