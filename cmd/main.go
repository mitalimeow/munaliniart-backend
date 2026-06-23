package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

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

func seedDatabase(dbPool *pgxpool.Pool) {
	ctx := context.Background()

	// Seed 16 initial artworks from seed_data folder if they don't exist
	for i := 1; i <= 16; i++ {
		filename := fmt.Sprintf("p%d.png", i)
		filePath := filepath.Join("seed_data", filename)

		// Check if file exists on disk
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue // Skip if not found
		}

		// Check if already in DB to prevent duplicates
		var exists bool
		err := dbPool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM art WHERE filename = $1)`, filename).Scan(&exists)
		if err != nil || exists {
			continue // Skip if error or already exists
		}

		// Read file bytes
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read seed file %s: %v", filename, err)
			continue
		}

		// Insert into DB
		mimeType := http.DetectContentType(data)
		_, err = dbPool.Exec(ctx, `INSERT INTO art (filename, mime_type, image_data) VALUES ($1, $2, $3)`, filename, mimeType, data)
		if err != nil {
			log.Printf("Failed to insert seed file %s: %v", filename, err)
		} else {
			log.Printf("Seeded initial artwork: %s", filename)
		}
	}
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

		CREATE TABLE IF NOT EXISTS art (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL,
			mime_type VARCHAR(100) NOT NULL,
			image_data BYTEA NOT NULL,
			uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Printf("Warning: Failed to auto-migrate tables: %v", err)
	}

	// Run Database Seeder
	seedDatabase(dbPool)

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

	// ── Public routes ─────────────────────────────────────────────────────────
	r.Get("/api/art", handlers.HandleArtGallery(artRepo)) // old route
	r.Get("/api/art/image/{id}", handlers.GetArtworkImage(artworkRepo)) // serves image bytes
	r.Get("/api/admin/art", handlers.GetAdminArtworks(artworkRepo)) // fetching metadata (used by public gallery too)
	
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
		p.Post("/api/admin/art", handlers.UploadArtwork(artworkRepo))
		p.Delete("/api/admin/art/{id}", handlers.DeleteAdminArtwork(artworkRepo))

		// Commissions (read-only admin view)
		p.Get("/api/admin/commissions", handlers.HandleCommissions(commRepo))
	})

	log.Printf("Server online on port %s...", appConfig.ServerPort)
	log.Fatal(http.ListenAndServe(":"+appConfig.ServerPort, r))
}
