package crawler

import "github.com/tdex-network/tdex-daemon/pkg/explorer"

const (
	CloseSignal EventType = iota
	AddressUnspents
	TransactionConfirmed
	TransactionUnconfirmed
	OutpointsUnspent
	OutpointsSpentAndUnconfirmed
	OutpointsSpentAndConfirmed
)

type EventType int

func (et EventType) String() string {
	switch et {
	case CloseSignal:
		return "CloseSignal"
	case AddressUnspents:
		return "AddressUnspents"
	case TransactionConfirmed:
		return "TransactionConfirmed"
	case TransactionUnconfirmed:
		return "TransactionUnconfirmed"
	case OutpointsUnspent:
		return "OutpointsUnspent"
	case OutpointsSpentAndUnconfirmed:
		return "OutpointsSpentAndUnconfirmed"
	case OutpointsSpentAndConfirmed:
		return "OutpointsSpentAndConfirmed"
	default:
		return "Unknown"
	}
}

type CloseEvent struct{}

func (q CloseEvent) Type() EventType {
	return CloseSignal
}

type AddressEvent struct {
	EventType EventType
	Address   string
	ExtraData interface{}
	Utxos     []explorer.Utxo
}

func (a AddressEvent) Type() EventType {
	return a.EventType
}

type TransactionEvent struct {
	TxID      string
	TxHex     string
	EventType EventType
	BlockHash string
	BlockTime int
	ExtraData interface{}
}

func (t TransactionEvent) Type() EventType {
	return t.EventType
}

type OutpointsEvent struct {
	EventType EventType
	Outpoints []Outpoint
	ExtraData interface{}
	TxID      string
	TxHex     string
	BlockHash string
	BlockTime int
}

func (o OutpointsEvent) Type() EventType {
	return o.EventType
}
