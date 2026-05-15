package utils

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("hunter2")
	if err != nil {
		t.Fatalf("hash err: %v", err)
	}
	if hash == "hunter2" {
		t.Fatal("hash should differ from plaintext")
	}
	if !CheckPassword("hunter2", hash) {
		t.Fatal("CheckPassword should accept the right password")
	}
	if CheckPassword("wrong", hash) {
		t.Fatal("CheckPassword should reject wrong password")
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("expected error for empty password")
	}
}
