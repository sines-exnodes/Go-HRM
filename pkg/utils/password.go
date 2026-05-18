package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// ErrEmptyPassword is returned when an empty password is provided.
var ErrEmptyPassword = errors.New("password must not be empty")

// bcryptCost is the bcrypt work factor. OWASP recommends a cost of at
// least 12 for password hashing.
const bcryptCost = 12

// HashPassword returns a bcrypt hash of the plaintext password using a
// cost factor of 12 (OWASP recommendation).
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", ErrEmptyPassword
	}
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// CheckPassword returns true when plain matches hashed.
func CheckPassword(plain, hashed string) bool {
	if plain == "" || hashed == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
