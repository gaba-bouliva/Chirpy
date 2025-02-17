package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	if len(password) < 3 {
		return "", fmt.Errorf("invalid password provided")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userId string, tokenSecret string, expiresIn time.Duration) (string, error) {
	claim := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(tokenSecret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		if claims.Issuer != "chirpy" {
			return "", fmt.Errorf("invalid access token issuer")
		}
		if claims.ExpiresAt.Time.Before(time.Now()) {
			return "", fmt.Errorf("expired access token provided")
		}
		return claims.Subject, nil
	}

	return "", fmt.Errorf("invalid token provided")
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenString := headers.Get("Authorization")
	tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer "))
	if tokenString == "" {
		return "", fmt.Errorf("authentication token not found")
	}
	return tokenString, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	tokenStr := headers.Get("Authorization")
	tokenStr = strings.TrimSpace(strings.TrimPrefix(tokenStr, "ApiKey "))
	if tokenStr == "" {
		return "", fmt.Errorf("authentication token not found")
	}
	return tokenStr, nil
}

func MakeRefreshToken() (string, error) {
	randBytes := make([]byte, 32)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}
	encodedStr := hex.EncodeToString(randBytes)
	return encodedStr, nil
}
