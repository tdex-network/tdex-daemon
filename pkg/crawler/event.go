package crawler

import "github.com/tdex-network/tdex-daemon/pkg/explorer"

const (
	QuitSignal EventType = iota
	FeeAccountDeposit
	MarketAccountDeposit
	TransactionConfirmed
	TransactionUnConfirmed
)

type EventType int

func (et EventType) String() string {
	switch et {
	case QuitSignal:
		return "QuitSignal"
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

type QuitEvent struct{}

func (q QuitEvent) Type() EventType {
	return QuitSignal
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
