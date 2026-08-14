package verifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type verifyRequest struct {
	Image string `json:"image"`
}

type verifyResponse struct {
	Verified bool   `json:"verified"`
	Image    string `json:"image"`
	Message  string `json:"message"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) Verify(ctx context.Context, image string) (bool, error) {
	payload := verifyRequest{
		Image: image,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshal verification request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/verify",
		bytes.NewReader(body),
	)
	if err != nil {
		return false, fmt.Errorf("create verification request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("call image verifier: %w", err)
	}
	defer resp.Body.Close()

	var result verifyResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode verifier response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return result.Verified, nil

	case http.StatusForbidden:
		return false, nil

	default:
		return false, fmt.Errorf(
			"image verifier returned HTTP %d",
			resp.StatusCode,
		)
	}
}