package main

import (
	"log"
	"net/http"
	"os"


	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go_tutorials/internal/config"
	"go_tutorials/internal/database"
	"go_tutorials/internal/handlers"
	customMiddleware "go_tutorials/internal/middleware"
	"go_tutorials/internal/repository"
)

// corsMiddleware adds CORS headers for the frontend dev server and production app
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigin := os.Getenv("FRONTEND_URL")
		
		// If frontend URL is set, only allow that one. Otherwise, allow localhost for dev.
		if allowedOrigin != "" {
			if origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		} else {
			// Fallback for local development
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	appConfig := config.LoadConfig()

	dbPool, err := database.ConnectDB(appConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	artRepo := repository.NewArtRepository(dbPool)
	commRepo := repository.NewCommissionRepository(dbPool)
	enqRepo := repository.NewEnquiryRepository(dbPool)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// ── Public routes ──────────────────────────────────────────────
	r.Get("/api/art", handlers.HandleArtGallery(artRepo))
	r.Get("/api/commissions", handlers.HandleCommissions(commRepo))

	// Contact / enquiry form (public)
	r.Post("/api/enquiry", handlers.HandleSubmitEnquiry(enqRepo))

	// Admin auth (rate-limited)
	// Pass appConfig directly to your handler
	r.With(customMiddleware.RateLimitAdmin).Post("/api/admin/login", handlers.AdminLogin(appConfig))
	r.Get("/api/admin/status", handlers.AdminStatus(appConfig.JWTSecret))

	// ── Protected admin routes ─────────────────────────────────────
	r.Group(func(p chi.Router) {
		p.Use(customMiddleware.RequireAdminAuth(appConfig.JWTSecret))

		p.Post("/api/admin/logout", handlers.AdminLogout())
		p.Get("/api/admin/enquiries", handlers.HandleGetEnquiries(enqRepo))
		p.Delete("/api/admin/enquiries/{id}", handlers.HandleDeleteEnquiry(enqRepo))

		// Protected art/commission management
		p.Post("/api/art", handlers.HandleArtGallery(artRepo))
		p.Put("/api/commissions", handlers.HandleCommissions(commRepo))
	})

	log.Printf("Server online on port %s...", appConfig.ServerPort)
	log.Fatal(http.ListenAndServe(":"+appConfig.ServerPort, r))
}
