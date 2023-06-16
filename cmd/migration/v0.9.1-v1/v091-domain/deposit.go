package v091domain

type Deposit struct {
	AccountIndex int
	TxID         string
	VOut         int
	Asset        string
	Value        uint64
	Timestamp    uint64
}

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
