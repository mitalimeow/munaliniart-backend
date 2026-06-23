package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go_tutorials/internal/repository"
)

// GetCommissions returns all commission metadata — used by both public page and admin
func GetCommissions(repo *repository.CommissionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		commissions, err := repo.GetAll(r.Context())
		if err != nil {
			http.Error(w, `{"error":"Failed to fetch commissions."}`, http.StatusInternalServerError)
			return
		}
		if commissions == nil {
			commissions = []repository.Commission{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(commissions)
	}
}

// GetCommissionImage serves the raw image bytes for a single commission
func GetCommissionImage(repo *repository.CommissionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid commission ID", http.StatusBadRequest)
			return
		}

		data, mimeType, err := repo.GetImage(r.Context(), id)
		if err != nil {
			http.Error(w, "Commission not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Write(data)
	}
}

// CreateCommission handles multipart form upload for a new commission
func CreateCommission(repo *repository.CommissionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const maxMem = 10 * 1024 * 1024 // 10 MB for form parsing
		if err := r.ParseMultipartForm(maxMem); err != nil {
			http.Error(w, `{"error":"Failed to parse form."}`, http.StatusBadRequest)
			return
		}

		customerName := r.FormValue("customer_name")
		review := r.FormValue("review")
		priceStr := r.FormValue("price")

		if customerName == "" || review == "" || priceStr == "" {
			http.Error(w, `{"error":"customer_name, review, and price are required."}`, http.StatusBadRequest)
			return
		}

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil || price <= 0 {
			http.Error(w, `{"error":"Invalid price value."}`, http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			http.Error(w, `{"error":"No image file provided."}`, http.StatusBadRequest)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, `{"error":"Failed to read image."}`, http.StatusInternalServerError)
			return
		}

		mimeType := http.DetectContentType(data)
		allowed := map[string]bool{"image/png": true, "image/jpeg": true}
		if !allowed[mimeType] {
			http.Error(w, `{"error":"Only PNG and JPEG images are accepted."}`, http.StatusBadRequest)
			return
		}

		id, err := repo.Create(r.Context(), customerName, review, price, header.Filename, mimeType, data)
		if err != nil {
			http.Error(w, `{"error":"Database error."}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      id,
			"message": "Commission created successfully.",
		})
	}
}

// DeleteCommission removes a commission record
func DeleteCommission(repo *repository.CommissionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"error":"Invalid commission ID."}`, http.StatusBadRequest)
			return
		}

		if err := repo.Delete(r.Context(), id); err != nil {
			http.Error(w, `{"error":"Commission not found or already deleted."}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Commission deleted successfully."})
	}
}
