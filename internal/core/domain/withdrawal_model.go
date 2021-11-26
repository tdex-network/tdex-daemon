package domain

// Withdrawal is used to follow funds withdrawal statistics
type Withdrawal struct {
	TxID            string
	AccountIndex    int
	BaseAmount      uint64
	QuoteAmount     uint64
	MillisatPerByte int64
	Address         string
	Timestamp       uint64
}
