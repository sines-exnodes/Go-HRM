package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

const testSecret = "test-secret-key-do-not-use-in-prod"

func TestSignAndVerifyAccessToken(t *testing.T) {
	uid := uuid.New()
	tok, err := SignToken(uid.String(), TokenTypeAccess, testSecret, time.Minute)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	claims, err := VerifyToken(tok, testSecret)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.Subject != uid.String() {
		t.Errorf("sub mismatch: got %s", claims.Subject)
	}
	if claims.Type != TokenTypeAccess {
		t.Errorf("type mismatch: got %s", claims.Type)
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("iat/exp must be set")
	}
}

func TestVerifyToken_ExpiredFails(t *testing.T) {
	tok, _ := SignToken("subj", TokenTypeAccess, testSecret, -time.Minute)
	if _, err := VerifyToken(tok, testSecret); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerifyToken_BadSignatureFails(t *testing.T) {
	tok, _ := SignToken("subj", TokenTypeAccess, testSecret, time.Minute)
	if _, err := VerifyToken(tok, "wrong-secret"); err == nil {
		t.Fatal("expected error for bad signature")
	}
}

func TestVerifyToken_TamperedPayloadFails(t *testing.T) {
	tok, _ := SignToken("subj", TokenTypeAccess, testSecret, time.Minute)
	tampered := tok[:len(tok)-4] + "xxxx"
	if _, err := VerifyToken(tampered, testSecret); err == nil {
		t.Fatal("expected error for tampered token")
	}
}
