package utils

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

var (
	lowerCharSet   = "abcdedfghijklmnopqrst"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "!@#$%&*"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
)

// Generate random 16 char password
func GeneratePassword() (string, string, error) {
	const length = 16
	randomPassword := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(allCharSet))))
		if err != nil {
			return "", "", err
		}
		randomPassword[i] = allCharSet[num.Int64()]
	}
	// bcrypt cost 10 (same as saltRounds = 10)
	hashedPassword, err := bcrypt.GenerateFromPassword(randomPassword, 10)
	if err != nil {
		return "", "", err
	}
	return string(randomPassword), string(hashedPassword), nil
}
