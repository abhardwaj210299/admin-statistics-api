package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"admin-statistics-api/internal/config"
	"admin-statistics-api/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// Number of game rounds to generate
	numRounds = 2_000_000

	// Number of unique user IDs
	numUsers = 500

	// Exchange rates to USD (simplified)
	ethToUSD  = 2000.0
	btcToUSD  = 50000.0
	usdtToUSD = 1.0

	// Batch size for MongoDB insertions
	batchSize = 1000
)

func main() {
	log.Println("Starting data seeding process...")

	// Load configuration
	cfg := config.DefaultConfig()

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Check connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully")

	// Get collection
	db := client.Database(cfg.MongoDB.Database)
	collection := db.Collection(cfg.MongoDB.Collection)

	// Drop existing collection
	if err := collection.Drop(ctx); err != nil {
		log.Printf("Warning: Failed to drop collection: %v", err)
	}

	// Generate user IDs
	userIDs := generateUserIDs(numUsers)

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Variables for progress tracking
	startTime := time.Now()
	lastProgressTime := startTime
	var transactions []interface{}

	log.Printf("Generating %d game rounds for %d users...", numRounds, numUsers)
	for i := 0; i < numRounds; i++ {
		// Generate a random round ID
		roundID := fmt.Sprintf("round-%d", i+1)

		// Choose a random user
		userID := userIDs[rand.Intn(len(userIDs))]

		// Choose a random currency
		currency := randomCurrency()

		// Generate a random time within the past year
		createdAt := randomTimeInPastYear()

		// Generate wager transaction
		wagerAmount := randomAmount()
		wagerUSDAmount := convertToUSD(wagerAmount, currency)
		wager := model.Transaction{
			ID:        model.GenerateULID(),
			CreatedAt: createdAt,
			UserID:    userID,
			RoundID:   roundID,
			Type:      model.TransactionTypeWager,
			Amount:    wagerAmount,
			Currency:  currency,
			USDAmount: wagerUSDAmount,
		}
		transactions = append(transactions, wager)

		// Generate payout transaction (later than wager)
		payoutCreatedAt := createdAt.Add(time.Duration(rand.Intn(300)) * time.Second)
		payoutAmount := randomAmount()
		payoutUSDAmount := convertToUSD(payoutAmount, currency)
		payout := model.Transaction{
			ID:        model.GenerateULID(),
			CreatedAt: payoutCreatedAt,
			UserID:    userID,
			RoundID:   roundID,
			Type:      model.TransactionTypePayout,
			Amount:    payoutAmount,
			Currency:  currency,
			USDAmount: payoutUSDAmount,
		}
		transactions = append(transactions, payout)

		// Insert in batches
		if len(transactions) >= batchSize {
			if err := insertBatch(ctx, collection, transactions); err != nil {
				log.Fatalf("Failed to insert batch: %v", err)
			}
			transactions = nil

			// Show progress
			now := time.Now()
			if now.Sub(lastProgressTime) > 5*time.Second {
				progress := float64(i+1) / float64(numRounds) * 100
				elapsed := now.Sub(startTime).Seconds()
				remaining := (elapsed / float64(i+1)) * float64(numRounds-i-1)
				log.Printf("Progress: %.2f%% (%.0f rounds/sec, %.0f seconds remaining)",
					progress, float64(i+1)/elapsed, remaining)
				lastProgressTime = now
			}
		}
	}

	// Insert any remaining transactions
	if len(transactions) > 0 {
		if err := insertBatch(ctx, collection, transactions); err != nil {
			log.Fatalf("Failed to insert final batch: %v", err)
		}
	}

	// Create indexes for better query performance
	log.Println("Creating indexes for optimization...")
	indexModels := []mongo.IndexModel{
		{
			Keys: primitive.D{{"createdAt", 1}},
		},
		{
			Keys: primitive.D{{"userId", 1}},
		},
		{
			Keys: primitive.D{{"roundId", 1}},
		},
		{
			Keys: primitive.D{{"type", 1}},
		},
		{
			Keys: primitive.D{{"currency", 1}},
		},
		// Compound indexes for common query patterns
		{
			Keys: primitive.D{{"createdAt", 1}, {"type", 1}},
		},
		{
			Keys: primitive.D{{"userId", 1}, {"createdAt", 1}},
		},
	}

	for _, indexModel := range indexModels {
		_, err := collection.Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	duration := time.Since(startTime)
	log.Printf("Seeding complete! Generated %d transactions (%d game rounds) in %s",
		numRounds*2, numRounds, duration)
}

// generateUserIDs generates unique user IDs
func generateUserIDs(count int) []string {
	userIDs := make([]string, count)
	for i := 0; i < count; i++ {
		userIDs[i] = model.GenerateULID()
	}
	return userIDs
}

// randomCurrency returns a random currency
func randomCurrency() string {
	currencies := []string{model.CurrencyETH, model.CurrencyBTC, model.CurrencyUSDT}
	return currencies[rand.Intn(len(currencies))]
}

// randomTimeInPastYear returns a random time in the past year
func randomTimeInPastYear() time.Time {
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)

	diff := now.Sub(oneYearAgo)
	randomDuration := time.Duration(rand.Int63n(int64(diff)))

	return oneYearAgo.Add(randomDuration)
}

// randomAmount generates a random amount between 0.01 and 100
func randomAmount() primitive.Decimal128 {
	// Generate a random amount between 0.01 and 100
	amount := 0.01 + rand.Float64()*99.99

	// Convert to Decimal128
	decimal, _ := primitive.ParseDecimal128(fmt.Sprintf("%.2f", amount))
	return decimal
}

// convertToUSD converts an amount in a given currency to USD
func convertToUSD(amount primitive.Decimal128, currency string) primitive.Decimal128 {
	// Convert Decimal128 to float64
	amountStr := amount.String()
	var amountFloat float64
	fmt.Sscanf(amountStr, "%f", &amountFloat)

	// Apply conversion rate
	var usdAmount float64
	switch currency {
	case model.CurrencyETH:
		usdAmount = amountFloat * ethToUSD
	case model.CurrencyBTC:
		usdAmount = amountFloat * btcToUSD
	case model.CurrencyUSDT:
		usdAmount = amountFloat * usdtToUSD
	}

	// Convert back to Decimal128
	usdDecimal, _ := primitive.ParseDecimal128(fmt.Sprintf("%.2f", usdAmount))
	return usdDecimal
}

// insertBatch inserts a batch of transactions into MongoDB
func insertBatch(ctx context.Context, collection *mongo.Collection, transactions []interface{}) error {
	_, err := collection.InsertMany(ctx, transactions)
	return err
}