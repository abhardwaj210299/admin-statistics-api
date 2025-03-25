package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"admin-statistics-api/internal/config"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create a test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			APIKey: "test-api-key",
		},
	}

	// Create a test router
	router := gin.New()
	router.Use(AuthMiddleware(cfg))
	
	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Test cases
	t.Run("allows request with valid API key", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "test-api-key")
		w := httptest.NewRecorder()
		
		// Act
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("blocks request with invalid API key", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "invalid-key")
		w := httptest.NewRecorder()
		
		// Act
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("blocks request with missing API key", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		
		// Act
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}