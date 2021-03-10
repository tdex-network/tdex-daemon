package dbbadger

import "errors"

var (
	// ErrMarketInvalidRequest ...
	ErrMarketInvalidRequest = errors.New("requested market is null")
	// ErrMarketNotFound ...
	ErrMarketNotFound = errors.New("market not found")
)
