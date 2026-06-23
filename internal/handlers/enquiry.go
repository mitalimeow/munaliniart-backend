package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go_tutorials/internal/repository"
)

type FormSubmission struct {
	Name    string `json:"sender_name"`
	Email   string `json:"sender_email"`
	Message string `json:"message"`
}

// HandleSubmitEnquiry handles POST /api/enquiry (public)
func HandleSubmitEnquiry(repo *repository.EnquiryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var form FormSubmission
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if form.Name == "" || form.Email == "" || form.Message == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		if err := repo.CreateEnquiry(r.Context(), form.Name, form.Email, form.Message); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Enquiry received"})
	}
}

// HandleGetEnquiries handles GET /api/admin/enquiries (protected)
func HandleGetEnquiries(repo *repository.EnquiryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logs, err := repo.GetAll(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if logs == nil {
			logs = []repository.Enquiry{}
		}
		json.NewEncoder(w).Encode(logs)
	}
}

// HandleDeleteEnquiry handles DELETE /api/admin/enquiries/{id} (protected)
func HandleDeleteEnquiry(repo *repository.EnquiryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid enquiry ID", http.StatusBadRequest)
			return
		}

		if err := repo.DeleteEnquiry(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Enquiry deleted"})
	}
}

// HandleEnquiries is kept for backward compatibility
func HandleEnquiries(repo *repository.EnquiryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			HandleSubmitEnquiry(repo)(w, r)
		case http.MethodGet:
			HandleGetEnquiries(repo)(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
