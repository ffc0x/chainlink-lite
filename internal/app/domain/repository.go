package domain

import (
	"time"

	"golang.org/x/net/context"
)

type EthPriceTicker interface {
	// Fetch the price of Ethereum in USD
	FetchPrice() (string, error)
}

type PriceMessageRepository interface {
	// Store the priceMsg if at least minInterval has passed since the last write
	// Timestamp is used to check the last write time
	// Returns true if the message was stored, false if it was skipped
	StorePriceIfAllowed(ctx context.Context, priceMsg *PriceMessage, minInterval time.Duration) (bool, error)
}
