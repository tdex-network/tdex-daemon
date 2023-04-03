package domain

const (
	StrategyTypeUndefined = iota
	StrategyTypePluggable
	StrategyTypeBalanced
	StrategyTypeUnbalanced
)

const (
	TradeStatusCodeUndefined = iota
	TradeStatusCodeProposal
	TradeStatusCodeAccepted
	TradeStatusCodeCompleted
	TradeStatusCodeSettled
	TradeStatusCodeExpired
)

const (
	TradeBuy TradeType = iota
	TradeSell
)

const (
	MinPercentageFee = 0
	MaxPercentageFee = 9999

	FeeAccount              = "fee_account"
	FeeFragmenterAccount    = "fee_fragmenter_account"
	MarketFragmenterAccount = "market_fragmenter_account"
)
