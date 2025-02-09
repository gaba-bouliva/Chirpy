package auth

import (
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	password := "password123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(hash) == 0 {
		t.Fatalf("expected hash to be non-empty")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	err = CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMakeJWT(t *testing.T) {
	userId := "user123"
	tokenSecret := "secret"
	expiresIn := time.Hour
	token, err := MakeJWT(userId, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(token) == 0 {
		t.Fatalf("expected token to be non-empty")
	}
}

func TestValidateJWT(t *testing.T) {
	userId := "user123"
	tokenSecret := "secret"
	expiresIn := time.Hour
	token, err := MakeJWT(userId, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	subject, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if subject != userId {
		t.Fatalf("expected subject to be %v, got %v", userId, subject)
	}
}
