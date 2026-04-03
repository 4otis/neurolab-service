package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string
	DBURL    string
	LogLevel string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Failed to load .env file, using default values")
	}

	return &Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		DBURL:    getDBURL(),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return defaultVal
}

func getDBURL() string {
	if dbURL := os.Getenv("PG_DB_URL"); dbURL != "" {
		return dbURL
	}

	host := getEnv("PG_DB_HOST", "localhost")
	port := getEnv("PG_DB_PORT", "5434")
	user := getEnv("PG_DB_USER", "postgres")
	password := getEnv("PG_DB_PASSWORD", "password")
	dbname := getEnv("PG_DB_NAME", "geonotify_db")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
}
