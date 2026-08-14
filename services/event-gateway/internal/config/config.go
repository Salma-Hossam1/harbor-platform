package config

import "os"

// Config contains all runtime configuration.
type Config struct {
	Port          string
	KafkaBrokers string
}

// Load reads configuration from environment variables.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka:9092"
	}

	return Config{
		Port:          port,
		KafkaBrokers: brokers,
	}
}