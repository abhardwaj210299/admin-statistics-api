package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"admin-statistics-api/internal/service"
)

// MockTransactionService implements service.TransactionServiceInterface for testing
type MockTransactionService struct {
	GGRFn               func(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error)
	DailyWagerVolumeFn  func(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error)
	UserPercentileFn    func(ctx context.Context, userID string, from, to time.Time) (float64, error)
}

// Make sure MockTransactionService implements the interface
var _ service.TransactionServiceInterface = (*MockTransactionService)(nil)

// CalculateGGR implements service.TransactionServiceInterface
func (m *MockTransactionService) CalculateGGR(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error) {
	if m.GGRFn != nil {
		return m.GGRFn(ctx, from, to)
	}
	return nil, errors.New("not implemented")
}

// CalculateDailyWagerVolume implements service.TransactionServiceInterface
func (m *MockTransactionService) CalculateDailyWagerVolume(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error) {
	if m.DailyWagerVolumeFn != nil {
		return m.DailyWagerVolumeFn(ctx, from, to)
	}
	return nil, errors.New("not implemented")
}

// CalculateUserWagerPercentile implements service.TransactionServiceInterface
func (m *MockTransactionService) CalculateUserWagerPercentile(ctx context.Context, userID string, from, to time.Time) (float64, error) {
	if m.UserPercentileFn != nil {
		return m.UserPercentileFn(ctx, userID, from, to)
	}
	return 0, errors.New("not implemented")
}

// Setup the test router
func setupTestRouter(mockService service.TransactionServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &TransactionHandler{
		service:  mockService,
		validate: validator.New(), // Initialize validator properly
	}

	router.GET("/gross_gaming_rev", handler.GetGrossGamingRevenue)
	router.GET("/daily_wager_volume", handler.GetDailyWagerVolume)
	router.GET("/user/:user_id/wager_percentile", handler.GetUserWagerPercentile)

	return router
}

func TestGetGrossGamingRevenue(t *testing.T) {
	// Test cases
	t.Run("returns 200 with valid data", func(t *testing.T) {
		// Arrange
		mockService := &MockTransactionService{
			GGRFn: func(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error) {
				return []map[string]interface{}{
					{
						"currency": "BTC",
						"ggr":      "10.50",
						"ggrUSD":   "525000.00",
					},
				}, nil
			},
		}
		router := setupTestRouter(mockService)

		// Setup request
		fromDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
		req, _ := http.NewRequest("GET", "/gross_gaming_rev?from="+fromDate.Format(time.RFC3339)+"&to="+toDate.Format(time.RFC3339), nil)
		w := httptest.NewRecorder()

		// Act
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, 200, w.Code)
		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "data")
		assert.Contains(t, response, "timeframe")
		data := response["data"].([]interface{})
		assert.Len(t, data, 1)
		firstItem := data[0].(map[string]interface{})
		assert.Equal(t, "BTC", firstItem["currency"])
	})

	t.Run("returns 400 with invalid date format", func(t *testing.T) {
		// Arrange
		mockService := &MockTransactionService{}
		router := setupTestRouter(mockService)

		// Setup request with invalid date
		req, _ := http.NewRequest("GET", "/gross_gaming_rev?from=invalid-date&to=2023-01-31T00:00:00Z", nil)
		w := httptest.NewRecorder()

		// Act
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, 400, w.Code)
		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"].(string), "Invalid date format")
	})

	t.Run("returns 500 when service returns error", func(t *testing.T) {
		// Arrange
		mockService := &MockTransactionService{
			GGRFn: func(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error) {
				return nil, errors.New("service error")
			},
		}
		router := setupTestRouter(mockService)

		// Setup request
		fromDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
		req, _ := http.NewRequest("GET", "/gross_gaming_rev?from="+fromDate.Format(time.RFC3339)+"&to="+toDate.Format(time.RFC3339), nil)
		w := httptest.NewRecorder()

		// Act
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, 500, w.Code)
		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"].(string), "Failed to calculate GGR")
	})
}

func TestGetDailyWagerVolume(t *testing.T) {
	// Test cases
	t.Run("returns 200 with valid data", func(t *testing.T) {
		// Arrange
		mockService := &MockTransactionService{
			DailyWagerVolumeFn: func(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error) {
				return []map[string]interface{}{
					{
						"date":           "2023-01-01",
						"currency":       "ETH",
						"wagerAmount":    "150.75",
						"wagerUSDAmount": "301500.00",
					},
				}, nil
			},
		}
		router := setupTestRouter(mockService)

		// Setup request
		fromDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
		req, _ := http.NewRequest("GET", "/daily_wager_volume?from="+fromDate.Format(time.RFC3339)+"&to="+toDate.Format(time.RFC3339), nil)
		w := httptest.NewRecorder()

		// Act
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, 200, w.Code)
		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "data")
		data := response["data"].([]interface{})
		assert.Len(t, data, 1)
		firstItem := data[0].(map[string]interface{})
		assert.Equal(t, "2023-01-01", firstItem["date"])
	})
}

func TestGetUserWagerPercentile(t *testing.T) {
	// Test cases
	t.Run("returns 200 with valid data", func(t *testing.T) {
		// Arrange
		mockService := &MockTransactionService{
			UserPercentileFn: func(ctx context.Context, userID string, from, to time.Time) (float64, error) {
				return 95.5, nil
			},
		}
		router := setupTestRouter(mockService)

		// Setup request
		userID := "65f7d1a8a2c40e1234567890"
		fromDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
		req, _ := http.NewRequest("GET", "/user/"+userID+"/wager_percentile?from="+fromDate.Format(time.RFC3339)+"&to="+toDate.Format(time.RFC3339), nil)
		w := httptest.NewRecorder()

		// Act
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, 200, w.Code)
		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "percentile")
		assert.Equal(t, 95.5, response["percentile"])
		assert.Equal(t, userID, response["userID"])
	})

	t.Run("returns 400 with missing user ID", func(t *testing.T) {
		// Arrange
		mockService := &MockTransactionService{}
		router := setupTestRouter(mockService)

		// Setup request without user ID
		fromDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
		req, _ := http.NewRequest("GET", "/user//wager_percentile?from="+fromDate.Format(time.RFC3339)+"&to="+toDate.Format(time.RFC3339), nil)
		w := httptest.NewRecorder()

		// Act
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, 400, w.Code) // Handler returns 400 Bad Request when user ID is empty
		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"].(string), "User ID is required")
	})

	t.Run("returns 500 when service returns error", func(t *testing.T) {
		// Arrange
		mockService := &MockTransactionService{
			UserPercentileFn: func(ctx context.Context, userID string, from, to time.Time) (float64, error) {
				return 0, errors.New("service error")
			},
		}
		router := setupTestRouter(mockService)

		// Setup request
		userID := "65f7d1a8a2c40e1234567890"
		fromDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
		req, _ := http.NewRequest("GET", "/user/"+userID+"/wager_percentile?from="+fromDate.Format(time.RFC3339)+"&to="+toDate.Format(time.RFC3339), nil)
		w := httptest.NewRecorder()

		// Act
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, 500, w.Code)
		var response map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"].(string), "Failed to calculate user wager percentile")
	})
}