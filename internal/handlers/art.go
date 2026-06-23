package handlers

import (
	"encoding/json"
	"net/http"

	"go_tutorials/internal/repository"
)

type ArtUploadPayload struct {
	Title      string `json:"title"`
	Medium     string `json:"medium"`
	Dimensions string `json:"dimensions"`
	ImageURL   string `json:"image_url"`
}

func HandleArtGallery(repo *repository.ArtRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := repo.GetGallery(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(items)

		case http.MethodPost: // Protected via middleware block setup upstream
			var load ArtUploadPayload
			if err := json.NewDecoder(r.Body).Decode(&load); err != nil {
				http.Error(w, "Malformed fields structure mapping error", http.StatusBadRequest)
				return
			}

			if load.Title == "" || load.ImageURL == "" {
				http.Error(w, "Title and Image URL parameters are mandatory fields", http.StatusBadRequest)
				return
			}

			err := repo.AddArt(r.Context(), load.Title, load.Medium, load.Dimensions, load.ImageURL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"status": "Artwork catalog entry published successfully"})
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
