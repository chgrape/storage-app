package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type key string

const UserIDKey key = "userId"

func Auth(jwks jwt.Keyfunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if tokenStr == "" {
			http.Error(w, "error: unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenStr, jwks)
		if err != nil || token == nil || !token.Valid {
			http.Error(w, fmt.Sprintf("error: unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
