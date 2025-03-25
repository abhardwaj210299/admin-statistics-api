package model

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Transaction represents a user transaction in the system
type Transaction struct {
	ID        string               `bson:"_id"`        // ULID string
	CreatedAt time.Time            `bson:"createdAt"`

	UserID    string               `bson:"userId"`     // ULID string
	RoundID   string               `bson:"roundId"`

	Type      string               `bson:"type"`       // Either "Wager" or "Payout"
	Amount    primitive.Decimal128 `bson:"amount"`     // Should always be >= 0
	Currency  string               `bson:"currency"`   // Either "ETH", "BTC", or "USDT"
	USDAmount primitive.Decimal128 `bson:"usdAmount"`  // The USD value of the `amount` and `currency`
}

// Transaction types
const (
	TransactionTypeWager  = "Wager"
	TransactionTypePayout = "Payout"
)

// Currency types
const (
	CurrencyETH  = "ETH"
	CurrencyBTC  = "BTC"
	CurrencyUSDT = "USDT"
)

// GenerateULID generates a new ULID string
func GenerateULID() string {
	// Create entropy source for ULID
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	
	// Generate ULID with current timestamp
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	
	return id.String()
}

// ParseULID parses a ULID string
func ParseULID(s string) (ulid.ULID, error) {
	return ulid.Parse(s)
}