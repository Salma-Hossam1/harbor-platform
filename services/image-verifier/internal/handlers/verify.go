package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"image-verifier/internal/verifier"
)

type VerifyRequest struct {
	Image string `json:"image"`
}

type VerifyResponse struct {
	Verified bool   `json:"verified"`
	Image    string `json:"image"`
	Message  string `json:"message"`
}

func Verify(v *verifier.Verifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request VerifyRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid JSON request", http.StatusBadRequest)
			return
		}

		if request.Image == "" {
			http.Error(w, "image is required", http.StatusBadRequest)
			return
		}

		err := v.Verify(r.Context(), request.Image)

		w.Header().Set("Content-Type", "application/json")

		if err != nil {

			slog.Error("image signature verification failed", "image", request.Image, "error", err)

			w.WriteHeader(http.StatusForbidden)

			_ = json.NewEncoder(w).Encode(VerifyResponse{
				Verified: false,
				Image:    request.Image,
				Message:  "image signature verification failed",
			})

			return
		}

		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(VerifyResponse{
			Verified: true,
			Image:    request.Image,
			Message:  "image signature verified",
		})
	}
}