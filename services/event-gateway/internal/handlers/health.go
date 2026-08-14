package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthResponse is the JSON returned by the health endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// Health handles GET /health requests.
func Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
