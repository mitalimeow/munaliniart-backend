package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"go_tutorials/internal/repository"
)

const (
	maxUploadSize  = 100 * 1024 // 100 KB
	uploadsDir     = "uploads/artworks"
)

var allowedMIMETypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
}

// UploadArtwork handles multipart/form-data image upload
func UploadArtwork(repo *repository.ArtworkRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enforce max request body size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1024)

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			http.Error(w, `{"error":"Only upload images smaller than 100 KB."}`, http.StatusBadRequest)
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
			http.Error(w, `{"error":"Only upload images smaller than 100 KB."}`, http.StatusBadRequest)
			return
		}

		// Validate MIME type by reading first 512 bytes
		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			http.Error(w, `{"error":"Failed to read file."}`, http.StatusInternalServerError)
			return
		}
		mimeType := http.DetectContentType(buf[:n])
		if !allowedMIMETypes[mimeType] {
			http.Error(w, `{"error":"Only PNG and JPEG images are accepted."}`, http.StatusBadRequest)
			return
		}

		// Seek back to start after reading for MIME detection
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		// Ensure upload directory exists
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			http.Error(w, `{"error":"Server storage error."}`, http.StatusInternalServerError)
			return
		}

		// Generate unique filename to prevent collisions
		ext := filepath.Ext(header.Filename)
		uniqueFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join(uploadsDir, uniqueFilename)

		out, err := os.Create(savePath)
		if err != nil {
			http.Error(w, `{"error":"Failed to save file."}`, http.StatusInternalServerError)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, `{"error":"Failed to write file."}`, http.StatusInternalServerError)
			return
		}

		// Store record in DB
		id, err := repo.CreateArtwork(r.Context(), header.Filename, "/"+savePath)
		if err != nil {
			os.Remove(savePath) // clean up file if DB insert fails
			http.Error(w, `{"error":"Database error."}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       id,
			"filename": header.Filename,
			"filepath": "/" + savePath,
			"message":  "Artwork uploaded successfully.",
		})
	}
}

// GetAdminArtworks returns all artworks from the database
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

// DeleteAdminArtwork removes an artwork from DB and deletes the file from disk
func DeleteAdminArtwork(repo *repository.ArtworkRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"error":"Invalid artwork ID."}`, http.StatusBadRequest)
			return
		}

		filePath, err := repo.DeleteArtwork(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"Artwork not found or already deleted."}`, http.StatusNotFound)
			return
		}

		// Delete the physical file (strip leading slash)
		diskPath := filePath
		if len(diskPath) > 0 && diskPath[0] == '/' {
			diskPath = diskPath[1:]
		}
		os.Remove(diskPath) // best-effort; don't fail if file is missing

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Artwork deleted successfully."})
	}
}
