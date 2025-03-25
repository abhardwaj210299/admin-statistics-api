package config

import (
	"os"
	"time"
)

// Config stores application configuration
type Config struct {
	MongoDB      MongoDBConfig
	HTTP         HTTPConfig
	Auth         AuthConfig
	Redis        RedisConfig
	CacheTimeout time.Duration
}

// MongoDBConfig stores MongoDB configuration
type MongoDBConfig struct {
	URI        string
	Database   string
	Collection string
}

// HTTPConfig stores HTTP server configuration
type HTTPConfig struct {
	Port    string
	Timeout time.Duration
}

// AuthConfig stores authentication configuration
type AuthConfig struct {
	APIKey string
}

// RedisConfig stores Redis configuration
type RedisConfig struct {
	URL string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		MongoDB: MongoDBConfig{
			URI:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database:   getEnv("MONGODB_DATABASE", "casino"),
			Collection: getEnv("MONGODB_COLLECTION", "transactions"),
		},
		HTTP: HTTPConfig{
			Port:    getEnv("HTTP_PORT", "8080"),
			Timeout: 30 * time.Second,
		},
		Auth: AuthConfig{
			APIKey: getEnv("API_KEY", "test-api-key"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		CacheTimeout: 5 * time.Minute,
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}