package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"admission-webhook/internal/metrics"
	"admission-webhook/internal/verifier"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ValidateHandler struct {
	verifier *verifier.Client
	logger   *slog.Logger
}

func NewValidateHandler(
	verifierClient *verifier.Client,
	logger *slog.Logger,
) *ValidateHandler {
	return &ValidateHandler{
		verifier: verifierClient,
		logger:   logger,
	}
}

func (h *ValidateHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	metrics.ValidationRequestsTotal.Inc()

	w.Header().Set("Content-Type", "application/json")

	var review admissionv1.AdmissionReview

	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		metrics.ValidationErrorsTotal.Inc()

		writeAdmissionError(
			w,
			"",
			fmt.Sprintf("malformed AdmissionReview: %v", err),
		)

		return
	}

	if review.Request == nil {
		metrics.ValidationErrorsTotal.Inc()

		writeAdmissionError(
			w,
			"",
			"AdmissionReview request is missing",
		)

		return
	}

	request := review.Request

	response := &admissionv1.AdmissionResponse{
		UID: request.UID,
	}

	if request.Resource.Resource != "pods" {
		response.Allowed = true

		writeAdmissionResponse(w, review, response)
		return
	}

	var pod corev1.Pod

	if err := json.Unmarshal(request.Object.Raw, &pod); err != nil {
		metrics.ValidationErrorsTotal.Inc()

		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("failed to decode Pod: %v", err),
		}

		writeAdmissionResponse(w, review, response)
		return
	}

	images := collectImages(&pod)

	if len(images) == 0 {
		response.Allowed = true

		writeAdmissionResponse(w, review, response)
		return
	}

	for _, image := range images {
		if !isDigestImage(image) {
			metrics.ValidationRejectedTotal.Inc()

			response.Allowed = false
			response.Result = &metav1.Status{
				Message: fmt.Sprintf(
					"image %q is not immutable: image must use a digest",
					image,
				),
			}

			writeAdmissionResponse(w, review, response)
			return
		}

		verified, err := h.verifier.Verify(
			r.Context(),
			image,
		)

		if err != nil {
			metrics.ValidationErrorsTotal.Inc()

			h.logger.Error(
				"image verification failed",
				"error", err,
			)

			response.Allowed = false
			response.Result = &metav1.Status{
				Message: fmt.Sprintf(
					"image verification unavailable for %q",
					image,
				),
			}

			writeAdmissionResponse(w, review, response)
			return
		}

		if !verified {
			metrics.ValidationRejectedTotal.Inc()

			response.Allowed = false
			response.Result = &metav1.Status{
				Message: fmt.Sprintf(
					"image %q is not signed by a trusted key",
					image,
				),
			}

			writeAdmissionResponse(w, review, response)
			return
		}
	}

	response.Allowed = true

	writeAdmissionResponse(w, review, response)
}

func collectImages(pod *corev1.Pod) []string {
	var images []string

	for _, container := range pod.Spec.InitContainers {
		images = append(images, container.Image)
	}

	for _, container := range pod.Spec.Containers {
		images = append(images, container.Image)
	}

	return images
}

func isDigestImage(image string) bool {
	return strings.Contains(image, "@sha256:")
}

func writeAdmissionResponse(
	w http.ResponseWriter,
	review admissionv1.AdmissionReview,
	response *admissionv1.AdmissionResponse,
) {
	responseReview := admissionv1.AdmissionReview{
		TypeMeta: review.TypeMeta,
		Response: response,
	}

	if responseReview.APIVersion == "" {
		responseReview.APIVersion = "admission.k8s.io/v1"
	}

	if responseReview.Kind == "" {
		responseReview.Kind = "AdmissionReview"
	}

	_ = json.NewEncoder(w).Encode(responseReview)
}

func writeAdmissionError(
	w http.ResponseWriter,
	uid types.UID,
	message string,
) {
	response := &admissionv1.AdmissionResponse{
		UID:     uid,
		Allowed: false,
		Result: &metav1.Status{
			Message: message,
		},
	}

	review := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: response,
	}

	_ = json.NewEncoder(w).Encode(review)
}