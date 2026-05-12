package middleware

import (
	"net/http"
	"strings"

	"queryforge/backend/internal/services"
)

func Auth(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeMiddlewareError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			userID, err := authService.ParseAccessToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				writeMiddlewareError(w, http.StatusUnauthorized, "invalid bearer token")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
		})
	}
}
