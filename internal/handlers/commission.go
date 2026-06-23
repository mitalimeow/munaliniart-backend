package handlers

import (
	"encoding/json"
	"net/http"

	"go_tutorials/internal/repository"
)

func HandleCommissions(repo *repository.CommissionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tiers, err := repo.GetTiers(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tiers)

		case http.MethodPut: // Protected via middleware block setup upstream
			var load repository.CommissionTier
			if err := json.NewDecoder(r.Body).Decode(&load); err != nil {
				http.Error(w, "Malformed fields parsing failure", http.StatusBadRequest)
				return
			}

			err := repo.UpdateTier(r.Context(), load.ID, load.TierName, load.Price, load.Description, load.IsAvailable)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "Commission Tier target configuration tracking modified successfully"})
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
