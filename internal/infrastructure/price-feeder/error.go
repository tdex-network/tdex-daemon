package pricefeederinfra

import "errors"

var (
	ErrPriceFeedNotFound              = errors.New("price feed not found")
	ErrPriceFeedMarketCannotBeChanged = errors.New("price feed market cannot be changed")
	ErrPriceFeedAlreadyExists         = errors.New("price feed already exists")
)
