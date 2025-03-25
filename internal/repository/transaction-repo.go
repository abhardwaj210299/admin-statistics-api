package repository

import (
	"context"
	"time"

	"admin-statistics-api/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TransactionRepository handles transaction data operations
type TransactionRepository struct {
	collection *mongo.Collection
}

// NewTransactionRepository creates a new TransactionRepository
func NewTransactionRepository(db *mongo.Database, collectionName string) *TransactionRepository {
	return &TransactionRepository{
		collection: db.Collection(collectionName),
	}
}

// InsertMany inserts multiple transactions
func (r *TransactionRepository) InsertMany(ctx context.Context, transactions []interface{}) error {
	_, err := r.collection.InsertMany(ctx, transactions)
	return err
}

// CalculateGGR calculates the Gross Gaming Revenue for a given time period
func (r *TransactionRepository) CalculateGGR(ctx context.Context, from, to time.Time) ([]bson.M, error) {
	pipeline := mongo.Pipeline{
		// Match transactions within the given time period
		{
			{"$match", bson.M{
				"createdAt": bson.M{
					"$gte": from,
					"$lte": to,
				},
			}},
		},
		// Group by currency and type
		{
			{"$group", bson.M{
				"_id": bson.M{
					"currency": "$currency",
					"type":     "$type",
				},
				"totalAmount":    bson.M{"$sum": "$amount"},
				"totalUSDAmount": bson.M{"$sum": "$usdAmount"},
			}},
		},
		// Reshape for wager and payout sums
		{
			{"$group", bson.M{
				"_id": "$_id.currency",
				"wager": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{
							bson.M{"$eq": bson.A{"$_id.type", model.TransactionTypeWager}},
							"$totalAmount",
							0,
						},
					},
				},
				"payout": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{
							bson.M{"$eq": bson.A{"$_id.type", model.TransactionTypePayout}},
							"$totalAmount",
							0,
						},
					},
				},
				"wagerUSD": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{
							bson.M{"$eq": bson.A{"$_id.type", model.TransactionTypeWager}},
							"$totalUSDAmount",
							0,
						},
					},
				},
				"payoutUSD": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{
							bson.M{"$eq": bson.A{"$_id.type", model.TransactionTypePayout}},
							"$totalUSDAmount",
							0,
						},
					},
				},
			}},
		},
		// Calculate GGR (wager - payout)
		{
			{"$project", bson.M{
				"currency": "$_id",
				"ggr":      bson.M{"$subtract": bson.A{"$wager", "$payout"}},
				"ggrUSD":   bson.M{"$subtract": bson.A{"$wagerUSD", "$payoutUSD"}},
				"_id":      0,
			}},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// CalculateDailyWagerVolume calculates daily wager volume
func (r *TransactionRepository) CalculateDailyWagerVolume(ctx context.Context, from, to time.Time) ([]bson.M, error) {
	pipeline := mongo.Pipeline{
		// Match wager transactions within the given time period
		{
			{"$match", bson.M{
				"createdAt": bson.M{
					"$gte": from,
					"$lte": to,
				},
				"type": model.TransactionTypeWager,
			}},
		},
		// Add a date field for grouping by day
		{
			{"$addFields", bson.M{
				"date": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d",
						"date":   "$createdAt",
					},
				},
			}},
		},
		// Group by date and currency
		{
			{"$group", bson.M{
				"_id": bson.M{
					"date":     "$date",
					"currency": "$currency",
				},
				"wagerAmount":    bson.M{"$sum": "$amount"},
				"wagerUSDAmount": bson.M{"$sum": "$usdAmount"},
			}},
		},
		// Reshape for better response format
		{
			{"$project", bson.M{
				"date":           "$_id.date",
				"currency":       "$_id.currency",
				"wagerAmount":    1,
				"wagerUSDAmount": 1,
				"_id":            0,
			}},
		},
		// Sort by date
		{
			{"$sort", bson.M{
				"date":     1,
				"currency": 1,
			}},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// CalculateUserWagerPercentile calculates user's percentile based on total wager amount
func (r *TransactionRepository) CalculateUserWagerPercentile(ctx context.Context, userID string, from, to time.Time) (float64, error) {
	// First, get the user's total wager
	userWagerPipeline := mongo.Pipeline{
		{
			{"$match", bson.M{
				"createdAt": bson.M{
					"$gte": from,
					"$lte": to,
				},
				"type":   model.TransactionTypeWager,
				"userId": userID,
			}},
		},
		{
			{"$group", bson.M{
				"_id":           "$userId",
				"totalWagerUSD": bson.M{"$sum": "$usdAmount"},
			}},
		},
	}

	userCursor, err := r.collection.Aggregate(ctx, userWagerPipeline)
	if err != nil {
		return 0, err
	}
	defer userCursor.Close(ctx)

	var userResults []bson.M
	if err = userCursor.All(ctx, &userResults); err != nil {
		return 0, err
	}

	if len(userResults) == 0 {
		return 0, nil // User has no wagers in this period
	}

	//userWagerUSD := userResults[0]["totalWagerUSD"]

	// Now calculate all users' wagers for ranking
	allUsersPipeline := mongo.Pipeline{
		{
			{"$match", bson.M{
				"createdAt": bson.M{
					"$gte": from,
					"$lte": to,
				},
				"type": model.TransactionTypeWager,
			}},
		},
		{
			{"$group", bson.M{
				"_id":           "$userId",
				"totalWagerUSD": bson.M{"$sum": "$usdAmount"},
			}},
		},
		{
			{"$sort", bson.M{
				"totalWagerUSD": -1, // Higher wagers first
			}},
		},
	}

	allUsersCursor, err := r.collection.Aggregate(ctx, allUsersPipeline)
	if err != nil {
		return 0, err
	}
	defer allUsersCursor.Close(ctx)

	var allUsersResults []bson.M
	if err = allUsersCursor.All(ctx, &allUsersResults); err != nil {
		return 0, err
	}

	totalUsers := len(allUsersResults)
	if totalUsers == 0 {
		return 0, nil
	}

	// Find user's position
	userRank := 0
	for i, result := range allUsersResults {
		id := result["_id"]
		if id.(string) == userID {
			userRank = i + 1
			break
		}
	}

	// Calculate percentile (higher rank = higher percentile)
	percentile := 100.0 - (float64(userRank-1) / float64(totalUsers) * 100.0)

	return percentile, nil
}

// Ensure TransactionRepository implements TransactionRepositoryInterface
var _ TransactionRepositoryInterface = (*TransactionRepository)(nil)