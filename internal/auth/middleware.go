package auth

import (
	"net/http"
	"strings"
	"errors"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix       = "Bearer "
)

// GetBearerToken extracts the Bearer token from the Authorization header
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get(AuthorizationHeader)
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	return strings.TrimPrefix(authHeader, BearerPrefix), nil
}

// AuthMiddleware validates JWT tokens for protected routes
func (v *JWTValidator) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := GetBearerToken(r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if err := v.Validate(token); err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}