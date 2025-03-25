package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// MockTransactionRepository is a mock implementation of the transaction repository for testing
type MockTransactionRepository struct {
	CalculateGGRFn                 func(ctx context.Context, from, to time.Time) ([]bson.M, error)
	CalculateDailyWagerVolumeFn    func(ctx context.Context, from, to time.Time) ([]bson.M, error)
	CalculateUserWagerPercentileFn func(ctx context.Context, userID string, from, to time.Time) (float64, error)
	
	// Track function calls
	CalculateGGRCalls                []struct{From, To time.Time}
	CalculateDailyWagerVolumeCalls   []struct{From, To time.Time}
	CalculateUserWagerPercentileCalls []struct{UserID string; From, To time.Time}
}

// NewMockTransactionRepository creates a new MockTransactionRepository
func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		CalculateGGRCalls:                make([]struct{From, To time.Time}, 0),
		CalculateDailyWagerVolumeCalls:   make([]struct{From, To time.Time}, 0),
		CalculateUserWagerPercentileCalls: make([]struct{UserID string; From, To time.Time}, 0),
		
		// Default implementations return empty results
		CalculateGGRFn: func(ctx context.Context, from, to time.Time) ([]bson.M, error) {
			return []bson.M{}, nil
		},
		CalculateDailyWagerVolumeFn: func(ctx context.Context, from, to time.Time) ([]bson.M, error) {
			return []bson.M{}, nil
		},
		CalculateUserWagerPercentileFn: func(ctx context.Context, userID string, from, to time.Time) (float64, error) {
			return 0, nil
		},
	}
}

// CalculateGGR mocks the CalculateGGR method
func (r *MockTransactionRepository) CalculateGGR(ctx context.Context, from, to time.Time) ([]bson.M, error) {
	r.CalculateGGRCalls = append(r.CalculateGGRCalls, struct{From, To time.Time}{from, to})
	return r.CalculateGGRFn(ctx, from, to)
}

// CalculateDailyWagerVolume mocks the CalculateDailyWagerVolume method
func (r *MockTransactionRepository) CalculateDailyWagerVolume(ctx context.Context, from, to time.Time) ([]bson.M, error) {
	r.CalculateDailyWagerVolumeCalls = append(r.CalculateDailyWagerVolumeCalls, struct{From, To time.Time}{from, to})
	return r.CalculateDailyWagerVolumeFn(ctx, from, to)
}

// CalculateUserWagerPercentile mocks the CalculateUserWagerPercentile method
func (r *MockTransactionRepository) CalculateUserWagerPercentile(ctx context.Context, userID string, from, to time.Time) (float64, error) {
	r.CalculateUserWagerPercentileCalls = append(r.CalculateUserWagerPercentileCalls, struct{UserID string; From, To time.Time}{userID, from, to})
	return r.CalculateUserWagerPercentileFn(ctx, userID, from, to)
}

// Verify implementation of interface
var _ TransactionRepositoryInterface = (*MockTransactionRepository)(nil)