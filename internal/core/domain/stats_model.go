package domain

// Withdrawal is used to follow funds withdrawal statistics
type Withdrawal struct {
	ID              uint `badgerhold:"key"`
	AccountIndex    int
	BaseAmount      uint64
	QuoteAmount     uint64
	MillisatPerByte int64
	Address         string
}

// Deposit is used to follow deposit statistics made by operator
type Deposit struct {
	AccountIndex int
	TxID         string
	VOut         int
}

// DepositKey represent the ID of an Deposit, composed by its txid and vout.
type DepositKey struct {
	TxID string
	VOut int
}

func (d Deposit) Key() DepositKey {
	return DepositKey{
		TxID: d.TxID,
		VOut: d.VOut,
	}
}
