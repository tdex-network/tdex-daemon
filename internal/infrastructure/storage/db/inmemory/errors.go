package inmemory

import "errors"

// Market errors
var (
	// ErrMarketNotExist ...
	ErrMarketNotExist = errors.New("market does not exists")
	// ErrMarketNotFound ...
	ErrMarketNotFound = errors.New("market not found")
	// ErrMarketInvalidRequest ...
	ErrMarketInvalidRequest = errors.New("requested market is null")
)

// Vault errors
var (
	// ErrVaultNotFound ...
	ErrVaultNotFound = errors.New("vault not found")
)

// Trade errors
var (
	// ErrTradeNotFound ...
	ErrTradeNotFound = errors.New("trade not found")
)
