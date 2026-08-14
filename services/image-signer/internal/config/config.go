package config

import (
	"fmt"
	"os"
)

type Config struct {
	KafkaBrokers string
	KafkaTopic   string
	KafkaGroupID string

	HarborRegistry string
	HarborUsername string
	HarborPassword string

	CosignBinary     string
	CosignPrivateKey string
	CosignPassword   string
	CosignAllowHTTP  bool
	CosignIgnoreTLog bool

	CosignAllowInsecure bool
}

func Load() (Config, error) {
	cfg := Config{
		KafkaBrokers: getenv("KAFKA_BROKERS", "kafka:9092"),
		KafkaTopic:   getenv("KAFKA_TOPIC", "harbor-events"),
		KafkaGroupID: getenv("KAFKA_GROUP_ID", "image-signer"),

		HarborRegistry: getenv("HARBOR_REGISTRY", ""),
		HarborUsername: getenv("HARBOR_USERNAME", ""),
		HarborPassword: getenv("HARBOR_PASSWORD", ""),

		CosignBinary:     getenv("COSIGN_BINARY", "/usr/local/bin/cosign"),
		CosignPrivateKey: getenv("COSIGN_PRIVATE_KEY", ""),
		CosignPassword:   getenv("COSIGN_PASSWORD", ""),

		CosignAllowHTTP:  os.Getenv("COSIGN_ALLOW_HTTP") == "true",
		CosignIgnoreTLog: os.Getenv("COSIGN_INSECURE_IGNORE_TLOG") == "true",

		CosignAllowInsecure: os.Getenv("COSIGN_ALLOW_INSECURE_REGISTRY") == "true",
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

	if cfg.CosignPrivateKey == "" {
		return Config{}, fmt.Errorf("COSIGN_PRIVATE_KEY is required")
	}

	if cfg.CosignPassword == "" {
		return Config{}, fmt.Errorf("COSIGN_PASSWORD is required")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
