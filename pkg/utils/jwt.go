package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType is the value of the "type" claim.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims is the JWT payload used by this service.
type Claims struct {
	Type TokenType `json:"type"`
	jwt.RegisteredClaims
}

// SignToken issues a signed HS256 token with {sub, type, exp, iat}.
// ttl may be negative — useful for tests of expired tokens.
func SignToken(subject string, tokenType TokenType, secret string, ttl time.Duration) (string, error) {
	if subject == "" {
		return "", errors.New("subject must not be empty")
	}
	if secret == "" {
		return "", errors.New("secret must not be empty")
	}
	now := time.Now().UTC()
	claims := Claims{
		Type: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// VerifyToken parses and validates a signed token. The token must be HS256
// and have a non-expired exp claim.
func VerifyToken(tokenString, secret string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	tok, err := parser.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
