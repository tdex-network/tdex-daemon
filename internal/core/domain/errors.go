package domain

import (
	"errors"
)

// Market errors
var (
	// ErrMarketFeeTooLow ...
	ErrMarketFeeTooLow = errors.New("market fee too low, must be at least 1 bp (0.01%)")
	// ErrMarketFeeTooHigh ...
	ErrMarketFeeTooHigh = errors.New("market fee too high, must be at most 9999 bp (99,99%)")
	// ErrMarketMissingBaseAsset ...
	ErrMarketMissingBaseAsset = errors.New("base asset is missing")
	// ErrMarketMissingQuoteAsset ...
	ErrMarketMissingQuoteAsset = errors.New("quote asset is missing")
	// ErrMarketTooManyAssets ...
	ErrMarketTooManyAssets = errors.New(
		"It's not possible to determine the correct asset pair of the market " +
			"because more than 2 type of assets has been found in the outpoint list",
	)
	//ErrMarketNotFunded is thrown when a market requires being funded for a change
	ErrMarketNotFunded = errors.New("market must be funded")
	//ErrMarketIsClosed is thrown when a market requires being tradable for a change
	ErrMarketIsClosed = errors.New("market is closed")
	//ErrMarketMustBeClosed is thrown when a market requires being NOT tradable for a change
	ErrMarketMustBeClosed = errors.New("market must be closed")
	//ErrMarketNotPriced is thrown when the price is still 0 (ie. not initialized)
	ErrMarketNotPriced = errors.New("price must be inserted")
	//ErrMarketInvalidBasePrice is thrown when the amount for Base price is an invalid satoshis value.
	ErrMarketInvalidBasePrice = errors.New("the amount for base price is invalid")
	//ErrMarketInvalidQuotePrice is thrown when the amount for Quote price is an invalid satoshis value.
	ErrMarketInvalidQuotePrice = errors.New("the amount for base price is invalid")
	// ErrMarketInvalidBaseAsset is thrown when non valid base asset is given
	ErrMarketInvalidBaseAsset = errors.New("invalid base asset")
	// ErrMarketInvalidQuoteAsset is thrown when non valid quote asset is given
	ErrMarketInvalidQuoteAsset = errors.New("invalid quote asset")
)

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
	// ErrInvalidAccount ...
	ErrInvalidAccount = errors.New("account index must be a positive integer number")
)
