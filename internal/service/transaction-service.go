package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"admin-statistics-api/internal/repository"
)

// TransactionService provides business logic for transactions
type TransactionService struct {
	repo  repository.TransactionRepositoryInterface
	cache repository.Cache
}

// NewTransactionService creates a new TransactionService
func NewTransactionService(repo repository.TransactionRepositoryInterface, cache repository.Cache) *TransactionService {
	return &TransactionService{
		repo:  repo,
		cache: cache,
	}
}

// CalculateGGR calculates the Gross Gaming Revenue
func (s *TransactionService) CalculateGGR(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("ggr:%s:%s", from.Format(time.RFC3339), to.Format(time.RFC3339))

	// Check cache
	if cachedData, found := s.cache.Get(cacheKey); found {
		// When using Redis, we need to handle the type conversion correctly
		switch data := cachedData.(type) {
		case []map[string]interface{}:
			return data, nil
		case []interface{}:
			// Convert from generic slice to the expected type
			result := make([]map[string]interface{}, len(data))
			for i, item := range data {
				if mapItem, ok := item.(map[string]interface{}); ok {
					result[i] = mapItem
				}
			}
			return result, nil
		default:
			// If we can't properly convert, just fetch from DB
			log.Printf("Cache type mismatch for key %s, fetching from DB", cacheKey)
		}
	}

	// Query the repository
	results, err := s.repo.CalculateGGR(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Convert to a more generic type
	response := make([]map[string]interface{}, len(results))
	for i, result := range results {
		response[i] = result
	}

	// Cache the results
	s.cache.Set(cacheKey, response, 5*time.Minute)

	return response, nil
}

// CalculateDailyWagerVolume calculates daily wager volume
func (s *TransactionService) CalculateDailyWagerVolume(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("daily_wager:%s:%s", from.Format(time.RFC3339), to.Format(time.RFC3339))

	// Check cache
	if cachedData, found := s.cache.Get(cacheKey); found {
		// When using Redis, we need to handle the type conversion correctly
		switch data := cachedData.(type) {
		case []map[string]interface{}:
			return data, nil
		case []interface{}:
			// Convert from generic slice to the expected type
			result := make([]map[string]interface{}, len(data))
			for i, item := range data {
				if mapItem, ok := item.(map[string]interface{}); ok {
					result[i] = mapItem
				}
			}
			return result, nil
		default:
			// If we can't properly convert, just fetch from DB
			log.Printf("Cache type mismatch for key %s, fetching from DB", cacheKey)
		}
	}

	// Query the repository
	results, err := s.repo.CalculateDailyWagerVolume(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Convert to a more generic type
	response := make([]map[string]interface{}, len(results))
	for i, result := range results {
		response[i] = result
	}

	// Cache the results
	s.cache.Set(cacheKey, response, 5*time.Minute)

	return response, nil
}

// CalculateUserWagerPercentile calculates user's wager percentile
func (s *TransactionService) CalculateUserWagerPercentile(ctx context.Context, userID string, from, to time.Time) (float64, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("percentile:%s:%s:%s", userID, from.Format(time.RFC3339), to.Format(time.RFC3339))

	// Check cache
	if cachedData, found := s.cache.Get(cacheKey); found {
		// When using Redis, we need to handle the type conversion correctly
		switch data := cachedData.(type) {
		case float64:
			return data, nil
		case int:
			return float64(data), nil
		case string:
			// Try to parse string to float64
			if val, err := strconv.ParseFloat(data, 64); err == nil {
				return val, nil
			}
		case map[string]interface{}:
			// Sometimes JSON unmarshals numbers into strings or floats
			if val, ok := data["value"].(float64); ok {
				return val, nil
			}
		default:
			// If we can't properly convert, just fetch from DB
			log.Printf("Cache type mismatch for key %s, fetching from DB", cacheKey)
		}
	}

	// Query the repository
	percentile, err := s.repo.CalculateUserWagerPercentile(ctx, userID, from, to)
	if err != nil {
		return 0, err
	}

	// Cache the result
	s.cache.Set(cacheKey, percentile, 5*time.Minute)

	return percentile, nil
}

// Ensure TransactionService implements TransactionServiceInterface
var _ TransactionServiceInterface = (*TransactionService)(nil)