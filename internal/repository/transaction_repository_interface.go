package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// TransactionRepositoryInterface defines the interface for transaction repositories
type TransactionRepositoryInterface interface {
	CalculateGGR(ctx context.Context, from, to time.Time) ([]bson.M, error)
	CalculateDailyWagerVolume(ctx context.Context, from, to time.Time) ([]bson.M, error)
	CalculateUserWagerPercentile(ctx context.Context, userID string, from, to time.Time) (float64, error)
}