package crawler

import "github.com/tdex-network/tdex-daemon/pkg/explorer"

const (
	FeeAccountDeposit EventType = iota
	MarketAccountDeposit
	TransactionConfirmed
	TransactionUnConfirmed
)

type EventType int

func (et EventType) String() string {
	switch et {
	case FeeAccountDeposit:
		return "FeeAccountDeposit"
	case MarketAccountDeposit:
		return "MarketAccountDeposit"
	case TransactionConfirmed:
		return "TransactionConfirmed"
	case TransactionUnConfirmed:
		return "TransactionUnConfirmed"
	default:
		return "Unknown"
	}
}

type AddressEvent struct {
	EventType    EventType
	AccountIndex int
	Address      string
	Utxos        []explorer.Utxo
}

func (a AddressEvent) Type() EventType {
	return a.EventType
}

type TransactionEvent struct {
	TxID      string
	EventType EventType
	BlockHash string
	BlockTime float64
}

func (t TransactionEvent) Type() EventType {
	return t.EventType
}
