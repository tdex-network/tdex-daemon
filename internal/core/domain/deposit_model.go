package domain

// Deposit is used to follow deposit statistics made by operator
type Deposit struct {
	AccountIndex int
	TxID         string
	VOut         int
	Asset        string
	Value        uint64
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
