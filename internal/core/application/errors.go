package application

import "errors"

var (
	// ErrFeeAccountNotFunded ...
	ErrFeeAccountNotFunded = errors.New("fee account not funded")
	// ErrUnknownStrategy ...
	ErrUnknownStrategy = errors.New("strategy not supported")
	// ErrTxNotConfirmed ...
	ErrTxNotConfirmed = errors.New("transaction not confirmed")
	// ErrMarketNotExist ...
	ErrMarketNotExist = errors.New("market does not exists")
	// ErrMarketNotFunded ...
	ErrMarketNotFunded = errors.New("market account not funded")
	// ErrMissingNonFundedMarkets ...
	ErrMissingNonFundedMarkets = errors.New("no non-funded markets found")
	// ErrInvalidOutpoint ...
	ErrInvalidOutpoint = errors.New("outpoint refers to inexistent tx output")
	// ErrInvalidOutpoints ...
	ErrInvalidOutpoints = errors.New("all outpoints must be funded for the same account")
	// ErrServiceUnavailable is the error returned by the trade service in case of
	// internal errors
	ErrServiceUnavailable = errors.New("service is unavailable, try again later")
	// ErrWebhookManagerNotInitialized is returned when attempting to use
	// AddWebhook or RemoveWebhook RPCs without having initialized the manager.
	ErrWebhookManagerNotInitialized = errors.New("webhook manager is not initialized")
)
