package middleware

import (
	"context"
	"net/http"

	"go_tutorials/internal/auth"
)

type contextKey string

const AdminUserKey contextKey = "admin_user"

func RequireAdminAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("admin_session_token")
			if err != nil {
				http.Error(w, "Unauthorized: No session active.", http.StatusUnauthorized)
				return
			}

			claims, err := auth.ValidateToken(cookie.Value, jwtSecret)
			if err != nil {
				http.Error(w, "Unauthorized: Session expired or invalid.", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AdminUserKey, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
