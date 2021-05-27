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

// Topics to be published
const (
	TradeSettled = iota
	AccountLowBalance
	AccountWithdraw
)
