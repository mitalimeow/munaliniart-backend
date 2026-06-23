package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"go_tutorials/internal/repository"
)

const maxUploadSize = 100 * 1024 // 100 KB

var allowedMIMETypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
}

// UploadArtwork handles multipart/form-data image upload and stores it in DB as BYTEA
func UploadArtwork(repo *repository.ArtworkRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enforce max request body size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1024)

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			http.Error(w, `{"error":"Only upload artwork smaller than 100 KB."}`, http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			http.Error(w, `{"error":"No image file provided."}`, http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate file size
		if header.Size > maxUploadSize {
			http.Error(w, `{"error":"Only upload artwork smaller than 100 KB."}`, http.StatusBadRequest)
			return
		}

		// Read into memory
		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, `{"error":"Failed to read file."}`, http.StatusInternalServerError)
			return
		}

		// Validate MIME type
		mimeType := http.DetectContentType(data)
		if !allowedMIMETypes[mimeType] {
			http.Error(w, `{"error":"Only PNG and JPEG images are accepted."}`, http.StatusBadRequest)
			return
		}

		// Store in DB
		id, err := repo.CreateArtwork(r.Context(), header.Filename, mimeType, data)
		if err != nil {
			http.Error(w, `{"error":"Database error."}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      id,
			"message": "Artwork uploaded successfully.",
		})
	}
}

// GetAdminArtworks returns metadata for all artworks
func GetAdminArtworks(repo *repository.ArtworkRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		artworks, err := repo.GetArtworks(r.Context())
		if err != nil {
			http.Error(w, `{"error":"Failed to fetch artworks."}`, http.StatusInternalServerError)
			return
		}
		if artworks == nil {
			artworks = []repository.Artwork{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(artworks)
	}
}

// GetArtworkImage serves the raw image bytes from the DB
func GetArtworkImage(repo *repository.ArtworkRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid artwork ID", http.StatusBadRequest)
			return
		}

		data, mimeType, err := repo.GetArtworkImage(r.Context(), id)
		if err != nil {
			http.Error(w, "Artwork not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		w.Write(data)
	}
}

// DeleteAdminArtwork removes an artwork from DB
func DeleteAdminArtwork(repo *repository.ArtworkRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"error":"Invalid artwork ID."}`, http.StatusBadRequest)
			return
		}

		err = repo.DeleteArtwork(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"Artwork not found or already deleted."}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Artwork deleted successfully."})
	}
}
