package v0domain

type Withdrawal struct {
	TxID            string
	AccountIndex    int
	BaseAmount      uint64
	QuoteAmount     uint64
	MillisatPerByte int64
	Address         string
	Timestamp       uint64
}
