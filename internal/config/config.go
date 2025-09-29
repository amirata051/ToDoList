package config

import (
	"log"
	"os"
)

// Config holds the application configuration.
type Config struct {
	MongoURI       string
	Port           string
	DbName         string
	CollectionName string
}

// Load loads configuration from environment variables or uses defaults.
func Load() *Config {
	return &Config{
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		Port:           getEnv("PORT", ":9000"),
		DbName:         getEnv("DB_NAME", "todo_mongo"),
		CollectionName: getEnv("COLLECTION_NAME", "todo"),
	}
}

// getEnv retrieves an environment variable or returns a fallback value.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Using default value for %s", key)
	return fallback
}