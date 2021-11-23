package application

import (
	"errors"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

var (
	ErrMarketInvalidBaseAsset  = domain.ErrMarketInvalidBaseAsset
	ErrMarketInvalidQuoteAsset = domain.ErrMarketInvalidQuoteAsset
	// ErrFeeAccountNotFunded ...
	ErrFeeAccountNotFunded = errors.New("fee account not funded")
	// ErrUnknownStrategy ...
	ErrUnknownStrategy = errors.New("strategy not supported")
	// ErrMarketAlreadyExist ...
	ErrMarketAlreadyExist = errors.New("market already exists")
	// ErrMarketNotExist ...
	ErrMarketNotExist = errors.New("market does not exist")
	// ErrMarketNotFunded ...
	ErrMarketNotFunded = errors.New("market account not funded")
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
	// ErrMarketBalanceTooLow is returned when the balance of the base or quote
	// asset of a market is below its fixed fee.
	ErrMarketBalanceTooLow = errors.New("market base or quote balance too low")
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
	// ErrWithdrawAmountTooBig is returned when attempting to withdraw more
	// than the total amount of LBTC asset of the fee account.
	ErrWithdrawAmountTooBig = errors.New(
		"amount to withdraw is too big",
	)
	// ErrInvalidAccountIndex returned if provided account index is invalid
	ErrInvalidAccountIndex = errors.New("invalid account index")
	// ErrMissingMarketBalanceToWithdraw is returned in case none of the base or
	// quote amounts to withdraw for a market are defined.
	ErrMissingMarketBalanceToWithdraw = errors.New(
		"either base or quote amount to withdraw must be specified",
	)
	// ErrMissingWithdrawAddress is returned if the receiving address to is not
	// set for a withdrawal.
	ErrMissingWithdrawAddress = errors.New("address must not be null")
	// ErrMarketNonZeroBalance is returned when trying to drop a market that still
	// owns some base or quote asset funds.
	ErrMarketNonZeroBalance = errors.New("market must not own funds to be dropped")
)
