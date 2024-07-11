package domain

import "errors"

// ErrNoPriceMessage is returned when no price message is found in the database.
var ErrNoPriceMessage = errors.New("no price message found")

// ErrFailedToFetchPrice is returned when the price cannot be fetched from the data api.
var ErrFailedToFetchPrice = errors.New("failed to fetch price")
