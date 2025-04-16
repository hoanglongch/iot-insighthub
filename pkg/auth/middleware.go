package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/time/rate"
)

var (
	// jwtSecret is loaded from an environment variable. A default is used for development.
	jwtSecret = []byte(getJWTSecret())
	// Global rate limiter: allow 10 requests per second.
	limiter = rate.NewLimiter(rate.Every(time.Second), 10)
)

// getJWTSecret returns the JWT secret from the environment, or a default value.
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// update later
		secret = "mydefaultsecret"
	}
	return secret
}

// AuthMiddleware validates the Bearer token and enforces rate limiting.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Enforce rate limiting.
		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Retrieve the Authorization header.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// Expect header format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// Parse and validate the JWT token.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			log.Printf("Invalid token: %v", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Optionally, attach token claims to the request context.
		ctx := context.WithValue(r.Context(), "user", token.Claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
