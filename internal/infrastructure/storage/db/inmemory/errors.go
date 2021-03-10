package inmemory

import "errors"

// Market errors
var (
	// ErrMarketNotExist is thrown when a market is not found
	ErrMarketNotExist = errors.New("market does not exists")
	// ErrMarketNotFound is thrown when there is no market associated to a given quote asset
	ErrMarketNotFound = errors.New("market not found")
	// ErrMarketInvalidRequest ...
	ErrMarketInvalidRequest = errors.New("requested market is null")
)

var (
	// ErrTradesNotFound is thrown when there is no trades associated to a given trade or swap ID
	ErrTradesNotFound = errors.New("no trades found for the given tradeID/SwapID")
	// ErrAlreadyLocked is thrown when trying to lock an already locked wallet
	ErrAlreadyLocked = errors.New("wallet is already locked")
	// ErrAlreadyUnlocked is thrown when trying to lunock an already unlocked wallet
	ErrAlreadyUnlocked = errors.New("wallet is already unlocked")
	// ErrWalletNotExist is thrown when mnemonic is not found
	ErrWalletNotExist = errors.New("wallet does not exist")
	// ErrWalletAlreadyExist is thrown when trying to create a new mnemonic if another one already exists
	ErrWalletAlreadyExist = errors.New("wallet already initialized with mnemonic")
	// ErrMustBeLocked is thrown when trying to change the passphrase with an unlocked wallet
	ErrMustBeLocked = errors.New("wallet must be locked to perform this operation")
	// ErrMustBeUnlocked is thrown when trying to make an operation that requires the wallet to be unlocked
	ErrMustBeUnlocked = errors.New("wallet must be unlocked to perform this operation")
	// ErrAccountNotExist is thrown when account is not found
	ErrAccountNotExist = errors.New("account does not exist")
)
