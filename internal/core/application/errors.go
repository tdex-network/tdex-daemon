package application

import "errors"

var (
	// ErrFeeAccountNotFunded ...
	ErrFeeAccountNotFunded = errors.New(
		"fee account must be funded to perform the requested operation",
	)
	// ErrUnknownStrategy ...
	ErrUnknownStrategy = errors.New("strategy not supported")
	// ErrMarketNotExist ...
	ErrMarketNotExist = errors.New("market does not exists")
)
