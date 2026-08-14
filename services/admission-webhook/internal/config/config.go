package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	MetricsPort string
	VerifierURL string
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "2112"
	}

	verifierURL := os.Getenv("VERIFIER_URL")
	if verifierURL == "" {
		return Config{}, fmt.Errorf("VERIFIER_URL is required")
	}

	return Config{
		Port:        port,
		MetricsPort: metricsPort,
		VerifierURL: verifierURL,
	}, nil
}