package util

import (
	"chainlink-lite/internal/app/domain"
	"encoding/json"

	"github.com/google/uuid"
)

func MarshalMessage(msg *domain.PriceMessage) ([]byte, error) {
	return json.Marshal(msg)
}

func UnmarshalMessage(data []byte) (*domain.PriceMessage, error) {
	var msg domain.PriceMessage
	err := json.Unmarshal(data, &msg)
	return &msg, err
}

func GenerateUUID() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}
