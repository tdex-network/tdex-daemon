package application

// Passphrase msg type
const (
	InitWallet = iota
	UnlockWallet
	ChangePassphrase
)

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

var (
	esploraUrlByNetwork = map[string]string{
		"liquid":  "https://blockstream.info/liquid",
		"testnet": "https://blockstream.info/liquidtestnet",
		"regtest": "http://localhost:5001",
	}
)
