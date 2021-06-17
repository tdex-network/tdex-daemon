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
	// ErrPubSubServiceNotInitialized is returned when attempting to use
	// AddWebhook or RemoveWebhook RPCs without having initialized the pubsub
	// service at the start up.
	ErrPubSubServiceNotInitialized = errors.New("pubsub service is not initialized")
	// ErrInvalidActionType is returned if the attempting to register a webhook
	// for an invalid action type.
	ErrInvalidActionType = errors.New("action type is unknown")
	// ErrMarketBaseBalanceTooLow is returned when the balance of the base asset
	// of a market is below 0.
	ErrMarketBaseBalanceTooLow = errors.New("market base balance is too low")
	// ErrMarketQuoteBalanceTooLow is returned when the balance of the quote asset
	// of a market is below 0.
	ErrMarketQuoteBalanceTooLow = errors.New("market quote balance is too low")
	// ErrWithdrawBaseAmountTooBig is returned when attempting to withdraw more
	// than the total amount of base asset of a market.
	ErrWithdrawBaseAmountTooBig = errors.New(
		"base amount to withdraw exceeds the market base balance",
	)
	// ErrWithdrawQuoteAmountTooBig is returned when attempting to withdraw more
	// than the total amount of quote asset of a market.
	ErrWithdrawQuoteAmountTooBig = errors.New(
		"quote amount to withdraw is too big",
	)
)
