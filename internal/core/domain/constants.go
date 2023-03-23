package domain

const (
	StrategyTypeUndefined = iota
	StrategyTypePluggable
	StrategyTypeBalanced
	StrategyTypeUnbalanced

	TradeStatusCodeUndefined = iota
	TradeStatusCodeProposal
	TradeStatusCodeAccepted
	TradeStatusCodeCompleted
	TradeStatusCodeSettled
	TradeStatusCodeExpired

	MinPercentageFee = 0
	MaxPercentageFee = 9999

	FeeAccount              = "fee_account"
	FeeFragmenterAccount    = "fee_fragmenter_account"
	MarketFragmenterAccount = "market_fragmenter_account"

	TradeBuy TradeType = iota
	TradeSell
)
