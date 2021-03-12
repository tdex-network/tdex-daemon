package dbbadger

import "errors"

var (
	// ErrMarketInvalidRequest ...
	ErrMarketInvalidRequest = errors.New("requested market is null")
	// ErrMarketNotFound ...
	ErrMarketNotFound = errors.New("market not found")
)

var (
	// ErrVaultNotFound ...
	ErrVaultNotFound = errors.New("vault not found")
)

var (
	// ErrTradeNotFound ...
	ErrTradeNotFound = errors.New("trade not found")
)
