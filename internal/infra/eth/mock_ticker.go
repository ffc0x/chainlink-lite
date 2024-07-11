package eth

import (
	"chainlink-lite/internal/app/domain"
	"math/rand"
	"strconv"
)

type MockEthPriceTicker struct {
}

var _ domain.EthPriceTicker = (*MockEthPriceTicker)(nil)

func NewMockTicker() *MockEthPriceTicker {
	return &MockEthPriceTicker{}
}

// Return a random price
func (e *MockEthPriceTicker) FetchPrice() (string, error) {
	// Returns a random price
	price := rand.Float64() * 4000
	return strconv.FormatFloat(price, 'f', 2, 64), nil
}
