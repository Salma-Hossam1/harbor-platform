package config

import "os"
import "log"
// Config contains the application's runtime configuration.
type Config struct {
	KafkaBrokers string
	KafkaTopic   string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	MetricsPort string
}

// Load reads configuration from environment variables.
func Load() Config {
	cfg := Config{
		KafkaBrokers: getenv("KAFKA_BROKERS", "kafka:9092"),
		KafkaTopic:   getenv("KAFKA_TOPIC", "harbor-events"),

		DBHost:     getenv("DB_HOST", "postgres"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBName:     getenv("DB_NAME", "metadata"),
		DBUser:     getenv("DB_USER", "metadata"),
		DBPassword: getenv("DB_PASSWORD", "metadatapass"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),
		MetricsPort: getenv("METRICS_PORT", "2112"),
	}


log.Printf("WORKER CONFIG: %+v", cfg)
	return cfg
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}