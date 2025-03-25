package service

import (
	"context"
	"time"
)

// TransactionServiceInterface defines the interface for transaction services
type TransactionServiceInterface interface {
	CalculateGGR(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error)
	CalculateDailyWagerVolume(ctx context.Context, from, to time.Time) ([]map[string]interface{}, error)
	CalculateUserWagerPercentile(ctx context.Context, userID string, from, to time.Time) (float64, error)
}