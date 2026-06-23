package handlers

import (
	"encoding/json"
	"go_tutorials/internal/auth"
	"go_tutorials/internal/config" // Import config package
	"net/http"
	"time"
)

type LoginRequest struct {
	PasswordOne string `json:"password_one"`
	PasswordTwo string `json:"password_two"`
}

// Pass the entire AppConfig profile into your initialization wrapper handler
func AdminLogin(appCfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Malformed JSON input stream", http.StatusBadRequest)
			return
		}

		// Secure Comparison using variables safely extracted from your config instance!
		if req.PasswordOne != appCfg.PasswordOne || req.PasswordTwo != appCfg.PasswordTwo {
			http.Error(w, "Invalid credentials matching failure.", http.StatusUnauthorized)
			return
		}

		token, err := auth.GenerateToken("admin-mrunalini", appCfg.JWTSecret)
		if err != nil {
			http.Error(w, "Internal token initialization error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session_token",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(1 * time.Hour),
			HttpOnly: true,
			Secure:   false, // Set to true when deploying live over HTTPS
			SameSite: http.SameSiteLaxMode,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Authenticated successfully"})
	}
}

// AdminStatus checks if the request contains a valid session cookie
func AdminStatus(jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_session_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		_, err = auth.ValidateToken(cookie.Value, jwtSecret)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "authenticated"})
	}
}

// AdminLogout clears the session cookie
func AdminLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session_token",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
	}
}
