package inmemory

import "errors"

var (
	// ErrMarketNotExist is thrown when a market is not found
	ErrMarketNotExist = errors.New("market does not exists")
	// ErrMarketsNotFound is thrown when there is no market associated to a given quote asset
	ErrMarketsNotFound = errors.New("no markets found for the given address")
	// ErrTradesNotFound is thrown when there is no trades associated to a given trade or swap ID
	ErrTradesNotFound = errors.New("no trades found for the given tradeID/SwapID")
)
