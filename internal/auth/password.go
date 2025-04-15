package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// HashPassword creates a SHA-256 hash of a password for storage
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:]), nil
}

// CheckPassword compares a password with a stored hash
func CheckPassword(storedHash, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	if hash != storedHash {
		return errors.New("invalid password")
	}

	return nil
}