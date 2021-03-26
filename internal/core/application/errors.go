package application

import "errors"

var (
	// ErrFeeAccountNotFunded ...
	ErrFeeAccountNotFunded = errors.New(
		"fee account must be funded to perform the requested operation",
	)
	// ErrUnknownStrategy ...
	ErrUnknownStrategy = errors.New("strategy not supported")
	// ErrTxNotConfirmed ...
	ErrTxNotConfirmed = errors.New("transaction not confirmed")
	// ErrMarketNotExist ...
	ErrMarketNotExist = errors.New("market does not exists")
	// ErrMissingNonFundedMarkets ...
	ErrMissingNonFundedMarkets = errors.New("no non-funded markets found")
	// ErrInvalidOutpoint ...
	ErrInvalidOutpoint = errors.New("outpoint refers to inexistent tx output")
)
