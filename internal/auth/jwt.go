package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
	secret string
}

func NewJWTValidator(secret string) *JWTValidator {
	return &JWTValidator{secret: secret}
}

// Validate checks if a JWT token is properly formatted and valid
func (v *JWTValidator) Validate(tokenString string) error {
	if tokenString == "" {
		return errors.New("empty token")
	}

	// Basic format check
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return errors.New("invalid JWT format")
	}

	// Parse with claims but don't verify yet (ClickHouse will verify)
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, &jwt.RegisteredClaims{})
	if err != nil {
		return fmt.Errorf("invalid JWT: %v", err)
	}

	return nil
}

// GenerateToken creates a JWT token for ClickHouse authentication
func (v *JWTValidator) GenerateToken(user string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   user,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(v.secret))
}

// ExtractUserFromToken gets the user/subject from a JWT token without verifying
func (v *JWTValidator) ExtractUserFromToken(tokenString string) (string, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &jwt.RegisteredClaims{})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		return claims.Subject, nil
	}

	return "", errors.New("invalid token claims")
}