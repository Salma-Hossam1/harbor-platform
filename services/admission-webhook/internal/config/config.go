package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	MetricsPort string
	VerifierURL string
	TLSCertFile string
	TLSKeyFile  string
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

	tlsCertFile := os.Getenv("TLS_CERT_FILE")
	if tlsCertFile == "" {
		return Config{}, fmt.Errorf("TLS_CERT_FILE is required")
	}

	tlsKeyFile := os.Getenv("TLS_KEY_FILE")
	if tlsKeyFile == "" {
		return Config{}, fmt.Errorf("TLS_KEY_FILE is required")
	}

	return Config{
		Port:        port,
		MetricsPort: metricsPort,
		VerifierURL: verifierURL,
		TLSCertFile: tlsCertFile,
		TLSKeyFile:  tlsKeyFile,
	}, nil
}