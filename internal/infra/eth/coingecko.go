package eth

import (
	"encoding/json"
	"net/http"

	"chainlink-lite/internal/app/domain"

	log "github.com/sirupsen/logrus"
)

type CoingeckoEthPriceTicker struct {
	url string
}

type Response struct {
	Ethereum struct {
		Usd json.Number `json:"usd"` // Use json.Number to avoid floating point precision issues
	} `json:"ethereum"`
}

var _ domain.EthPriceTicker = (*CoingeckoEthPriceTicker)(nil)

func NewEthPriceTicker(url string) *CoingeckoEthPriceTicker {
	return &CoingeckoEthPriceTicker{url: url}
}

// Fetch the price of Ethereum in USD
func (e *CoingeckoEthPriceTicker) FetchPrice() (string, error) {
	resp, err := http.Get(e.url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if the response status code is not 200
	if resp.StatusCode != http.StatusOK {
		log.Debug("Status code is not 200: ", "status ", resp.StatusCode)
		return "", domain.ErrFailedToFetchPrice
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Debug("Failed to decode response ", err)
		return "", domain.ErrFailedToFetchPrice
	}

	return string(result.Ethereum.Usd), nil
}
