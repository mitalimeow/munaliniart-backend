package main

import (
	"context"
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

		// If FRONTEND_URL is set, only allow that origin. Otherwise, allow localhost for dev.
		if allowedOrigin != "" {
			if origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		} else {
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

	// ── Auto-migrate Database Tables ─────────────────────────────────────────
	_, err = dbPool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS enquiries (
			id SERIAL PRIMARY KEY,
			sender_name VARCHAR(100) NOT NULL,
			sender_email VARCHAR(255) NOT NULL,
			message TEXT NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			submitted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS artworks (
			id SERIAL PRIMARY KEY,
			filename TEXT NOT NULL,
			filepath TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Printf("Warning: Failed to auto-migrate tables: %v", err)
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	artRepo := repository.NewArtRepository(dbPool)
	artworkRepo := repository.NewArtworkRepository(dbPool)
	commRepo := repository.NewCommissionRepository(dbPool)
	enqRepo := repository.NewEnquiryRepository(dbPool)

	// ── Router ────────────────────────────────────────────────────────────────
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// Serve uploaded artwork files statically
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// ── Public routes ─────────────────────────────────────────────────────────
	r.Get("/api/art", handlers.HandleArtGallery(artRepo))
	r.Get("/api/commissions", handlers.HandleCommissions(commRepo))
	r.Post("/api/enquiry", handlers.HandleSubmitEnquiry(enqRepo))

	// ── Admin auth (rate-limited, public) ────────────────────────────────────
	r.With(customMiddleware.RateLimitAdmin).Post("/api/admin/login", handlers.AdminLogin(appConfig))
	r.Get("/api/admin/status", handlers.AdminStatus(appConfig.JWTSecret))

	// ── Protected admin routes ────────────────────────────────────────────────
	r.Group(func(p chi.Router) {
		p.Use(customMiddleware.RequireAdminAuth(appConfig.JWTSecret))

		p.Post("/api/admin/logout", handlers.AdminLogout())

		// Enquiries
		p.Get("/api/admin/enquiries", handlers.HandleGetEnquiries(enqRepo))
		p.Delete("/api/admin/enquiries/{id}", handlers.HandleDeleteEnquiry(enqRepo))

		// Artworks (upload management)
		p.Get("/api/admin/art", handlers.GetAdminArtworks(artworkRepo))
		p.Post("/api/admin/art", handlers.UploadArtwork(artworkRepo))
		p.Delete("/api/admin/art/{id}", handlers.DeleteAdminArtwork(artworkRepo))

		// Commissions (read-only admin view)
		p.Get("/api/admin/commissions", handlers.HandleCommissions(commRepo))
	})

	log.Printf("Server online on port %s...", appConfig.ServerPort)
	log.Fatal(http.ListenAndServe(":"+appConfig.ServerPort, r))
}
