package application

// trade type
const (
	TradeBuy = iota
	TradeSell
)

// restore status
const (
	Processing = iota
	Done
)

// webhook action types
const (
	TradeSettled      = iota
	AccountLowBalance = iota
	AccountWithdraw   = iota
	AllActions        = iota
)
