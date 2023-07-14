package v0domain

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

	StrategyTypePluggable  = 0
	StrategyTypeBalanced   = 1
	StrategyTypeUnbalanced = 2
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
