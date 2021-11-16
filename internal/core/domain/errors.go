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
	// ErrInvalidFixedFee ...
	ErrInvalidFixedFee = errors.New("fixed fee must be a positive value")
	// ErrMarketPreviewAmountTooLow is returned when a preview fails because
	// the provided amount makes the previewed amount to be too low (lower than
	// the optional fixed fee).
	ErrMarketPreviewAmountTooLow = errors.New("provided amount is too low")
	// ErrMarketPreviewAmountTooBig is returned when a preview fails because
	// the provided amount makes the previewed amount to be too big (greater than
	// the overall balance).
	ErrMarketPreviewAmountTooBig = errors.New("provided amount is too big")
)

// Unspent errors
var (
	// ErrUnspentAlreadyLocked ...
	ErrUnspentAlreadyLocked = errors.New("cannot lock an already locked unspent")
)

// Account errors
var (
	// ErrInvalidAccount ...
	ErrInvalidAccount = errors.New("account index must be a positive integer number")
)

// Vault errors
var (
	// ErrVaultMustBeLocked is thrown when trying to change the passphrase with an unlocked wallet
	ErrVaultMustBeLocked = errors.New("wallet must be locked to perform this operation")
	// ErrVaultMustBeUnlocked is thrown when trying to make an operation that requires the wallet to be unlocked
	ErrVaultMustBeUnlocked = errors.New("wallet must be unlocked to perform this operation")
	// ErrVaultInvalidPassphrase ...
	ErrVaultInvalidPassphrase = errors.New("passphrase is not valid")
	// ErrVaultAlreadyInitialized ...
	ErrVaultAlreadyInitialized = errors.New("vault is already initialized")
	// ErrVaultNullMnemonicOrPassphrase ...
	ErrVaultNullMnemonicOrPassphrase = errors.New("mnemonic and/or passphrase must not be null")
	// ErrVaultNullNetwork ...
	ErrVaultNullNetwork = errors.New("network must not be null")
	// ErrVaultAccountNotFound ...
	ErrVaultAccountNotFound = errors.New("account not found")
)

// Trade errors
var (
	// ErrTradeMustBeEmpty ...
	ErrTradeMustBeEmpty = errors.New(
		"trade must be empty for parsing a proposal",
	)
	// ErrTradeMustBeProposal ...
	ErrTradeMustBeProposal = errors.New(
		"trade must be in proposal state for being accepted",
	)
	// ErrTradeMustBeAccepted ...
	ErrTradeMustBeAccepted = errors.New(
		"trade must be in accepted state for being completed",
	)
	// ErrTradeMustBeCompleted ...
	ErrTradeMustBeCompletedOrAccepted = errors.New(
		"trade must be in completed or accepted to be settled",
	)
	// ErrTradeExpirationDateNotReached ...
	ErrTradeExpirationDateNotReached = errors.New(
		"trade must have reached the expiration date to be set expired",
	)
	// ErrTradeExpired ...
	ErrTradeExpired = errors.New("trade has expired")
	// ErrTradeNullExpirationDate ...
	ErrTradeNullExpirationDate = errors.New(
		"trade must have an expiration date set to be set expired",
	)
)
