package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"admission-webhook/internal/verifier"
)

func testLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(
			httptest.NewRecorder(),
			nil,
		),
	)
}

func TestValidateSignedImage(t *testing.T) {
	verifierServer := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"verified": true,
				"message": "image signature verified"
			}`))
		}),
	)
	defer verifierServer.Close()

	client := verifier.NewClient(verifierServer.URL)

	handler := NewValidateHandler(
		client,
		testLogger(),
	)

	requestBody := `{
		"apiVersion":"admission.k8s.io/v1",
		"kind":"AdmissionReview",
		"request":{
			"uid":"test-1",
			"resource":{
				"resource":"pods"
			},
			"object":{
				"apiVersion":"v1",
				"kind":"Pod",
				"spec":{
					"containers":[
						{
							"name":"app",
							"image":"localhost:8088/demo/app@sha256:abc123"
						}
					]
				}
			}
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/validate",
		strings.NewReader(requestBody),
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"allowed":true`) {
		t.Fatalf("expected request to be allowed: %s", rec.Body.String())
	}
}