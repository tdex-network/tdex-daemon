package inmemory

import "errors"

var (
	// ErrMarketNotExist is thrown when a market is not found
	ErrMarketNotExist = errors.New("market does not exists")
)
