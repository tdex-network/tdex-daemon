package domain

import "errors"

var (
	// ErrMustBeLocked is thrown when trying to change the passphrase with an unlocked wallet
	ErrMustBeLocked = errors.New("wallet must be locked to perform this operation")
	// ErrMustBeUnlocked is thrown when trying to make an operation that requires the wallet to be unlocked
	ErrMustBeUnlocked = errors.New("wallet must be unlocked to perform this operation")
	// ErrInvalidPassphrase ...
	ErrInvalidPassphrase = errors.New("passphrase is not valid")
	// ErrVaultAlreadyInitialized ...
	ErrVaultAlreadyInitialized = errors.New("vault is already initialized")
	// ErrNullMnemonicOrPassphrase ...
	ErrNullMnemonicOrPassphrase = errors.New("mnemonic and/or passphrase must not be null")
	//ErrNotFunded is thrown when a market requires being funded for a change
	ErrNotFunded = errors.New("market must be funded")
	//ErrMarketMustBeOpen is thrown when a market requires being tradable for a change
	ErrMarketMustBeOpen = errors.New("market must be opened")
	//ErrMarketMustBeClose is thrown when a market requires being NOT tradable for a change
	ErrMarketMustBeClose = errors.New("market must be closed")
	//ErrPriceExists is thrown when a price for that given timestamp already exists
	ErrPriceExists = errors.New("price has been inserted already")
	//ErrNotPriced is thrown when the price is still 0 (ie. not initialized)
	ErrNotPriced = errors.New("price must be inserted")
	// ErrMustBeEmpty ...
	ErrMustBeEmpty = errors.New(
		"trade must be empty for parsing a proposal",
	)
	// ErrMustBeProposal ...
	ErrMustBeProposal = errors.New(
		"trade must be in proposal state for being accepted",
	)
	// ErrMustBeAccepted ...
	ErrMustBeAccepted = errors.New(
		"trade must be in accepted state for being completed",
	)
	// ErrMustBeCompleted ...
	ErrMustBeCompleted = errors.New(
		"trade must be in completed state to add txid",
	)
	// ErrExpirationDateNotReached ...
	ErrExpirationDateNotReached = errors.New(
		"trade did not reached expiration date yet and cannot be set expired",
	)
	// ErrMarketNotExist ...
	ErrMarketNotExist = errors.New("market does not exists")
)
