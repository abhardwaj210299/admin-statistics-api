package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"admin-statistics-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCalculateGGR(t *testing.T) {
	// Setup
	mockRepo := repository.NewMockTransactionRepository()
	mockCache := repository.NewMockCache()
	service := NewTransactionService(mockRepo, mockCache)

	// Test data
	ctx := context.Background()
	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
	cacheKey := "ggr:2023-01-01T00:00:00Z:2023-01-31T00:00:00Z"

	// Test cases
	t.Run("returns cached data when available", func(t *testing.T) {
		// Arrange
		cachedResult := []map[string]interface{}{
			{
				"currency": "BTC",
				"ggr":      "10.50",
				"ggrUSD":   "525000.00",
			},
		}
		mockCache.Set(cacheKey, cachedResult, time.Minute)

		// Act
		result, err := service.CalculateGGR(ctx, from, to)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, cachedResult, result)
		assert.Len(t, mockRepo.CalculateGGRCalls, 0, "Repository should not be called when cache hit")
		assert.Contains(t, mockCache.GetCalls, cacheKey, "Cache should be queried")
	})

	t.Run("fetches and caches data when not in cache", func(t *testing.T) {
		// Arrange - reset mocks
		mockRepo = repository.NewMockTransactionRepository()
		mockCache = repository.NewMockCache()
		service = NewTransactionService(mockRepo, mockCache)

		// Setup expected repository response
		repoResult := []bson.M{
			{
				"currency": "BTC",
				"ggr":      "10.50",
				"ggrUSD":   "525000.00",
			},
		}
		mockRepo.CalculateGGRFn = func(ctx context.Context, from, to time.Time) ([]bson.M, error) {
			return repoResult, nil
		}

		// Act
		result, err := service.CalculateGGR(ctx, from, to)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "BTC", result[0]["currency"])
		assert.Len(t, mockRepo.CalculateGGRCalls, 1, "Repository should be called when cache miss")
		assert.Contains(t, mockCache.GetCalls, cacheKey, "Cache should be queried")
		assert.Contains(t, mockCache.SetCalls, cacheKey, "Result should be cached")
	})

	t.Run("handles repository errors", func(t *testing.T) {
		// Arrange - reset mocks
		mockRepo = repository.NewMockTransactionRepository()
		mockCache = repository.NewMockCache()
		service = NewTransactionService(mockRepo, mockCache)

		// Setup expected repository error
		expectedError := errors.New("database error")
		mockRepo.CalculateGGRFn = func(ctx context.Context, from, to time.Time) ([]bson.M, error) {
			return nil, expectedError
		}

		// Act
		result, err := service.CalculateGGR(ctx, from, to)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		assert.Len(t, mockRepo.CalculateGGRCalls, 1, "Repository should be called when cache miss")
	})
}

func TestCalculateDailyWagerVolume(t *testing.T) {
	// Setup
	mockRepo := repository.NewMockTransactionRepository()
	mockCache := repository.NewMockCache()
	service := NewTransactionService(mockRepo, mockCache)

	// Test data
	ctx := context.Background()
	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
	cacheKey := "daily_wager:2023-01-01T00:00:00Z:2023-01-31T00:00:00Z"

	// Test cases
	t.Run("returns cached data when available", func(t *testing.T) {
		// Arrange
		cachedResult := []map[string]interface{}{
			{
				"date":           "2023-01-01",
				"currency":       "ETH",
				"wagerAmount":    "150.75",
				"wagerUSDAmount": "301500.00",
			},
		}
		mockCache.Set(cacheKey, cachedResult, time.Minute)

		// Act
		result, err := service.CalculateDailyWagerVolume(ctx, from, to)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, cachedResult, result)
		assert.Len(t, mockRepo.CalculateDailyWagerVolumeCalls, 0, "Repository should not be called when cache hit")
	})

	t.Run("fetches and caches data when not in cache", func(t *testing.T) {
		// Arrange - reset mocks
		mockRepo = repository.NewMockTransactionRepository()
		mockCache = repository.NewMockCache()
		service = NewTransactionService(mockRepo, mockCache)

		// Setup expected repository response
		repoResult := []bson.M{
			{
				"date":           "2023-01-01",
				"currency":       "ETH",
				"wagerAmount":    "150.75",
				"wagerUSDAmount": "301500.00",
			},
		}
		mockRepo.CalculateDailyWagerVolumeFn = func(ctx context.Context, from, to time.Time) ([]bson.M, error) {
			return repoResult, nil
		}

		// Act
		result, err := service.CalculateDailyWagerVolume(ctx, from, to)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "2023-01-01", result[0]["date"])
		assert.Len(t, mockRepo.CalculateDailyWagerVolumeCalls, 1, "Repository should be called when cache miss")
	})
}

func TestCalculateUserWagerPercentile(t *testing.T) {
	// Setup
	mockRepo := repository.NewMockTransactionRepository()
	mockCache := repository.NewMockCache()
	service := NewTransactionService(mockRepo, mockCache)

	// Test data
	ctx := context.Background()
	userID := "01HRMD5HGTZB3TW3PGYXRD07CQT" // ULID string instead of ObjectID
	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
	cacheKey := "percentile:" + userID + ":2023-01-01T00:00:00Z:2023-01-31T00:00:00Z"

	// Test cases
	t.Run("returns cached data when available", func(t *testing.T) {
		// Arrange
		cachedResult := 95.5
		mockCache.Set(cacheKey, cachedResult, time.Minute)

		// Act
		result, err := service.CalculateUserWagerPercentile(ctx, userID, from, to)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, cachedResult, result)
		assert.Len(t, mockRepo.CalculateUserWagerPercentileCalls, 0, "Repository should not be called when cache hit")
	})

	t.Run("fetches and caches data when not in cache", func(t *testing.T) {
		// Arrange - reset mocks
		mockRepo = repository.NewMockTransactionRepository()
		mockCache = repository.NewMockCache()
		service = NewTransactionService(mockRepo, mockCache)

		// Setup expected repository response
		expectedPercentile := 95.5
		mockRepo.CalculateUserWagerPercentileFn = func(ctx context.Context, userID string, from, to time.Time) (float64, error) {
			return expectedPercentile, nil
		}

		// Act
		result, err := service.CalculateUserWagerPercentile(ctx, userID, from, to)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedPercentile, result)
		assert.Len(t, mockRepo.CalculateUserWagerPercentileCalls, 1, "Repository should be called when cache miss")
	})

	t.Run("handles error from repository", func(t *testing.T) {
		// Arrange - reset mocks
		mockRepo = repository.NewMockTransactionRepository()
		mockCache = repository.NewMockCache()
		service = NewTransactionService(mockRepo, mockCache)

		// Setup expected repository error
		expectedError := errors.New("database error")
		mockRepo.CalculateUserWagerPercentileFn = func(ctx context.Context, userID string, from, to time.Time) (float64, error) {
			return 0, expectedError
		}

		// Act
		result, err := service.CalculateUserWagerPercentile(ctx, userID, from, to)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Equal(t, float64(0), result)
		assert.Len(t, mockRepo.CalculateUserWagerPercentileCalls, 1, "Repository should be called when cache miss")
	})
}