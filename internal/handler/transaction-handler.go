package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"admin-statistics-api/internal/service"
)

// TransactionHandler handles HTTP requests for transactions
type TransactionHandler struct {
	service  service.TransactionServiceInterface
	validate *validator.Validate
}

// NewTransactionHandler creates a new TransactionHandler
func NewTransactionHandler(service service.TransactionServiceInterface) *TransactionHandler {
	return &TransactionHandler{
		service:  service,
		validate: validator.New(),
	}
}

// TimeframeParams represents query parameters for date range
type TimeframeParams struct {
	From time.Time `form:"from" validate:"required"`
	To   time.Time `form:"to" validate:"required,gtefield=From"`
}

// GetGrossGamingRevenue handles the GGR endpoint
func (h *TransactionHandler) GetGrossGamingRevenue(c *gin.Context) {
	var params TimeframeParams

	// Parse query parameters
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use ISO 8601 (YYYY-MM-DDThh:mm:ssZ)"})
		return
	}

	// Validate parameters
	if err := h.validate.Struct(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error: " + err.Error()})
		return
	}

	// Call service to get GGR
	results, err := h.service.CalculateGGR(c, params.From, params.To)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate GGR: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"timeframe": gin.H{"from": params.From, "to": params.To},
		"data":      results,
	})
}

// GetDailyWagerVolume handles the daily wager volume endpoint
func (h *TransactionHandler) GetDailyWagerVolume(c *gin.Context) {
	var params TimeframeParams

	// Parse query parameters
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use ISO 8601 (YYYY-MM-DDThh:mm:ssZ)"})
		return
	}

	// Validate parameters
	if err := h.validate.Struct(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error: " + err.Error()})
		return
	}

	// Call service to get daily wager volume
	results, err := h.service.CalculateDailyWagerVolume(c, params.From, params.To)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate daily wager volume: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"timeframe": gin.H{"from": params.From, "to": params.To},
		"data":      results,
	})
}

// GetUserWagerPercentile handles the user wager percentile endpoint
func (h *TransactionHandler) GetUserWagerPercentile(c *gin.Context) {
	var params TimeframeParams

	// Get user ID from path
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Parse query parameters
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use ISO 8601 (YYYY-MM-DDThh:mm:ssZ)"})
		return
	}

	// Validate parameters
	if err := h.validate.Struct(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error: " + err.Error()})
		return
	}

	// Call service to get user wager percentile
	percentile, err := h.service.CalculateUserWagerPercentile(c, userID, params.From, params.To)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate user wager percentile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID":     userID,
		"percentile": percentile,
		"timeframe":  gin.H{"from": params.From, "to": params.To},
	})
}