package repository

import (
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestRedisCache(t *testing.T) {
	// Skip real tests if INTEGRATION_TESTS environment variable is not set
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests")
	}

	// These tests connect to a real Redis server, so we need REDIS_URL env var
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	t.Run("connects to redis server", func(t *testing.T) {
		// Arrange & Act
		cache, err := NewRedisCache(redisURL)
		defer cache.Close()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, cache)
	})

	t.Run("set and get operations work", func(t *testing.T) {
		// Arrange
		cache, err := NewRedisCache(redisURL)
		assert.NoError(t, err)
		defer cache.Close()

		key := "test-key-" + time.Now().String()
		value := map[string]interface{}{
			"test": "value",
			"num":  123,
		}

		// Act
		cache.Set(key, value, 10*time.Second)
		retrievedValue, found := cache.Get(key)

		// Assert
		assert.True(t, found)
		assert.NotNil(t, retrievedValue)
		assert.Equal(t, "value", retrievedValue.(map[string]interface{})["test"])
		assert.Equal(t, float64(123), retrievedValue.(map[string]interface{})["num"]) // JSON serializes numbers to float64
	})

	t.Run("delete operation works", func(t *testing.T) {
		// Arrange
		cache, err := NewRedisCache(redisURL)
		assert.NoError(t, err)
		defer cache.Close()

		key := "test-delete-key"
		value := "test-value"
		cache.Set(key, value, 10*time.Second)

		// Verify it's there
		_, found := cache.Get(key)
		assert.True(t, found)

		// Act
		cache.Delete(key)

		// Assert
		_, found = cache.Get(key)
		assert.False(t, found)
	})

	t.Run("expiration works", func(t *testing.T) {
		// Arrange
		cache, err := NewRedisCache(redisURL)
		assert.NoError(t, err)
		defer cache.Close()

		key := "test-expiration-key"
		value := "test-value"

		// Act
		cache.Set(key, value, 1*time.Second) // Very short expiration
		_, foundBefore := cache.Get(key)
		
		// Wait for expiration
		time.Sleep(2 * time.Second)
		_, foundAfter := cache.Get(key)

		// Assert
		assert.True(t, foundBefore)
		assert.False(t, foundAfter)
	})
}

// Mock Redis tests using miniredis
func TestRedisCache_WithMiniRedis(t *testing.T) {
	// Start a miniredis server
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer s.Close()

	// Create a Redis URL pointing to the miniredis server
	redisURL := "redis://" + s.Addr()

	t.Run("set and get operations work with miniredis", func(t *testing.T) {
		// Arrange
		cache, err := NewRedisCache(redisURL)
		assert.NoError(t, err)
		defer cache.Close()

		key := "test-key"
		value := map[string]interface{}{
			"test": "value",
			"num":  123,
		}

		// Act
		cache.Set(key, value, 10*time.Second)
		retrievedValue, found := cache.Get(key)

		// Assert
		assert.True(t, found)
		assert.NotNil(t, retrievedValue)
		assert.Equal(t, "value", retrievedValue.(map[string]interface{})["test"])
		assert.Equal(t, float64(123), retrievedValue.(map[string]interface{})["num"])
	})

	t.Run("expiration works with miniredis", func(t *testing.T) {
		// Arrange
		cache, err := NewRedisCache(redisURL)
		assert.NoError(t, err)
		defer cache.Close()

		key := "test-expiration-key"
		value := "test-value"

		// Act
		cache.Set(key, value, 10*time.Second)
		_, foundBefore := cache.Get(key)
		
		// Fast-forward time in miniredis
		s.FastForward(15 * time.Second)
		
		_, foundAfter := cache.Get(key)

		// Assert
		assert.True(t, foundBefore)
		assert.False(t, foundAfter)
	})
}