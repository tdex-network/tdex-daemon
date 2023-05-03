package domain

import (
	"errors"
	"fmt"
)

// Market errors
var (
	// ErrMarketFeeTooHigh ...
	ErrMarketInvalidPercentageFee = fmt.Errorf(
		"invalid market percentage fee, must be in range [%d, %d]",
		MinPercentageFee, MaxPercentageFee,
	)
	//ErrMarketIsClosed is thrown when a market requires being tradable for a change
	ErrMarketIsClosed = errors.New("the market is paused, please open it first")
	//ErrMarketMustBeClosed is thrown when a market requires being NOT tradable for a change
	ErrMarketIsOpen = errors.New("the market is active, please pause it first")
	//ErrMarketNotPriced is thrown when the price is still 0 (ie. not initialized)
	ErrMarketNotPriced = errors.New("the selected strategy mandates price to be updated manually")
	//ErrMarketInvalidBasePrice is thrown when the amount for Base price is an invalid satoshis value.
	ErrMarketInvalidBasePrice = errors.New("invalid base price amount, must be > 0")
	//ErrMarketInvalidQuotePrice is thrown when the amount for Quote price is an invalid satoshis value.
	ErrMarketInvalidQuotePrice = errors.New("invalid quote price amount, must be > 0")
	// ErrMarketInvalidBaseAsset is thrown when non valid base asset is given.
	ErrMarketInvalidBaseAsset = errors.New("invalid base asset")
	// ErrMarketInvalidQuoteAsset is thrown when non valid quote asset is given.
	ErrMarketInvalidQuoteAsset = errors.New("invalid quote asset")
	// ErrMarketInvalidBaseAssetPrecision is thrown when non valid quote asset precision is given.
	ErrMarketInvalidBaseAssetPrecision = errors.New("base asset precision must be in range [0, 8]")
	// ErrMarketInvalidQuoteAssetPrecision is thrown when non valid quote asset precision is given.
	ErrMarketInvalidQuoteAssetPrecision = errors.New("quote asset precision must be in range [0, 8]")
	// ErrInvalidFixedFee ...
	ErrMarketInvalidFixedFee = errors.New("invalid fixed fee amount")
	// ErrMarketPreviewAmountTooLow is returned when a preview fails because
	// the provided amount makes the previewed amount to be too low (lower than
	// the optional fixed fee).
	ErrMarketPreviewAmountTooLow = errors.New("provided amount is too low")
	// ErrMarketPreviewAmountTooBig is returned when a preview fails because
	// the provided amount makes the previewed amount to be too big (greater than
	// the overall balance).
	ErrMarketPreviewAmountTooBig = errors.New("provided amount is too big")
	// ErrMarketUnknownStrategy is thrown when an invalid strategy is given at
	// market creation.
	ErrMarketUnknownStrategy = errors.New("unknown market strategy")
)

// Trade errors
var (
	// ErrTradeUnknownType ...
	ErrTradeUnknownType = errors.New("unknown trade type")
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
	// ErrTradeInvalidExpiryTime ...
	ErrTradeInvalidExpiryTime = errors.New(
		"trade expiration date must be after proposal one",
	)
	// ErrTradeExpiryTimeNotReached ...
	ErrTradeExpiryTimeNotReached = errors.New(
		"trade must have reached the expiration date to be set expired",
	)
	// ErrTradeExpired ...
	ErrTradeExpired = errors.New("trade has expired")
	// ErrTradeNullExpiryTime ...
	ErrTradeNullExpiryTime = errors.New(
		"trade must have an expiration date set to be set expired",
	)
)
