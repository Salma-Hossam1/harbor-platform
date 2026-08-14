package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                  string
	MetricsPort           string
	HarborRegistry        string
	HarborUsername        string
	HarborPassword        string
	CosignPublicKey       string
	CosignBinary          string
	CosignIgnoreTLog      bool
}

func Load() (Config, error) {
	cfg := Config{
		Port:             getenv("PORT", "8080"),
		MetricsPort:      getenv("METRICS_PORT", "2112"),
		HarborRegistry:   os.Getenv("HARBOR_REGISTRY"),
		HarborUsername:   os.Getenv("HARBOR_USERNAME"),
		HarborPassword:   os.Getenv("HARBOR_PASSWORD"),
		CosignPublicKey:  os.Getenv("COSIGN_PUBLIC_KEY"),
		CosignBinary:     getenv("COSIGN_BINARY", "/usr/local/bin/cosign"),
		CosignIgnoreTLog: os.Getenv("COSIGN_INSECURE_IGNORE_TLOG") == "true",
	}

	if cfg.HarborRegistry == "" {
		return Config{}, fmt.Errorf("HARBOR_REGISTRY is required")
	}

	if cfg.HarborUsername == "" {
		return Config{}, fmt.Errorf("HARBOR_USERNAME is required")
	}

	if cfg.HarborPassword == "" {
		return Config{}, fmt.Errorf("HARBOR_PASSWORD is required")
	}

	if cfg.CosignPublicKey == "" {
		return Config{}, fmt.Errorf("COSIGN_PUBLIC_KEY is required")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}