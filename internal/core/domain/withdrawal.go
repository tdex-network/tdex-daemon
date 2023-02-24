package domain

// Withdrawal holds info about txs with funds sent from a wallet account.
type Withdrawal struct {
	AccountName       string
	TxID              string
	TotAmountPerAsset map[string]uint64
	Timestamp         int64
}
