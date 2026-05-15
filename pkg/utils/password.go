package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// ErrEmptyPassword is returned when an empty password is provided.
var ErrEmptyPassword = errors.New("password must not be empty")

// HashPassword returns a bcrypt hash of the plaintext password using the
// default cost (10).
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", ErrEmptyPassword
	}
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
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
