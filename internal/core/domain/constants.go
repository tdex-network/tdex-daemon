package domain

const (
	FeeAccount = iota
	WalletAccount
	FeeFragmenterAccount
	MarketFragmenterAccount
	UnusedAccount3
	MarketAccountStart

	ExternalChain = 0
	InternalChain = 1

	MinMilliSatPerByte = 100

	StrategyTypePluggable  StrategyType = 0
	StrategyTypeBalanced   StrategyType = 1
	StrategyTypeUnbalanced StrategyType = 2
)

const (
	Empty = iota - 1
	Undefined
	Proposal
	Accepted
	Completed
	Settled
	Expired
)
