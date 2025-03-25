package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"admin-statistics-api/internal/config"
)

// AuthMiddleware provides a middleware function for validating API keys.
// This middleware checks the incoming request's "Authorization" header against the
// expected API key configured in the application. If the key does not match,
// the middleware will abort the request with an Unauthorized status, ensuring
// that only authorized requests can access protected routes.
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the API key from the request header
		authHeader := c.GetHeader("Authorization")

		// If the API key is invalid or missing, respond with an error and stop processing
		if authHeader != cfg.Auth.APIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or missing API key",
			})
			return
		}

		// If the API key is valid, continue to the next middleware/handler
		c.Next()
	}
}
