package httpserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/Freedom-Club-Sec/Coldwire-server/internal/crypto"
)

type ctxKey string

var claimsKey = ctxKey("jwtClaims")

func (s *Server) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		token, claims, err := crypto.VerifyJWT(tokenString, s.Cfg.JWTSecret)
		if err != nil || !token.Valid {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Attach claims to context
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
