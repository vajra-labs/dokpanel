package rclone

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
)

// obscureKey is the static key used by rclone to obscure passwords in configuration files.
var obscureKey = []byte{
	0x9c, 0x93, 0x5b, 0x48, 0x73, 0x0a, 0x55, 0x4d,
	0x6b, 0xfd, 0x7c, 0x63, 0xc8, 0x86, 0xa9, 0x2b,
	0xd3, 0x90, 0x19, 0x8e, 0xb8, 0x12, 0x8a, 0xfb,
	0xf4, 0xde, 0x16, 0x2b, 0x8b, 0x95, 0xf6, 0x38,
}

// Obscure encrypts the plaintext using AES-CTR with the static rclone key and a random IV,
// returning the base64.RawURLEncoding representation (with IV prepended).
func Obscure(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	plaintextBytes := []byte(plaintext)

	block, err := aes.NewCipher(obscureKey)
	if err != nil {
		return "", err
	}

	// Generate a random IV of size 16 (AES block size)
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	// Encrypt using CTR mode
	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(plaintextBytes))
	stream.XORKeyStream(ciphertext, plaintextBytes)

	// Combine IV and ciphertext
	combined := append(iv, ciphertext...)

	return base64.RawURLEncoding.EncodeToString(combined), nil
}

// Reveal decrypts a password obscured by Obscure/rclone obscure.
func Reveal(obscured string) (string, error) {
	if obscured == "" {
		return "", nil
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(obscured)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", aes.KeySizeError(aes.BlockSize)
	}

	iv := ciphertext[:aes.BlockSize]
	data := ciphertext[aes.BlockSize:]

	block, err := aes.NewCipher(obscureKey)
	if err != nil {
		return "", err
	}

	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(data))
	stream.XORKeyStream(plaintext, data)

	return string(plaintext), nil
}

// obscurePassword obscures the password using our Go implementation.
// Falls back to the plaintext password if obscuring fails.
func obscurePassword(password string) string {
	obscured, err := Obscure(password)
	if err != nil {
		return password
	}
	return obscured
}
