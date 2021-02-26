package crawler

import "github.com/tdex-network/tdex-daemon/pkg/explorer"

const (
	CloseSignal EventType = iota
	FeeAccountDeposit
	MarketAccountDeposit
	TransactionConfirmed
	TransactionUnConfirmed
)

type EventType int

func (et EventType) String() string {
	switch et {
	case CloseSignal:
		return "CloseSignal"
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

type CloseEvent struct{}

func (q CloseEvent) Type() EventType {
	return CloseSignal
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
