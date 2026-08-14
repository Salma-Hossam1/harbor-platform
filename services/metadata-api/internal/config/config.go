package config

import (
	"log"
	"os"
)

type Config struct {
	Port string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string
}


func Load() Config {
	cfg := Config{
		Port: getenv("PORT", "8081"),

		DBHost:     getenv("DB_HOST", "postgres"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBName:     getenv("DB_NAME", "metadata"),
		DBUser:     getenv("DB_USER", "metadata"),
		DBPassword: getenv("DB_PASSWORD", "metadatapass"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),
	}

	log.Printf("CONFIG: %+v", cfg)

	return cfg
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}