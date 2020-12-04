package crawler

import (
	"bytes"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type AddressObservable struct {
	AccountIndex int
	Address      string
	BlindingKey  []byte
}

func (a *AddressObservable) observe(
	w *sync.WaitGroup,
	explorerSvc explorer.Service,
	errChan chan error,
	eventChan chan Event,
) {
	defer w.Done()

	if a == nil {
		return
	}

	unspents, err := explorerSvc.GetUnspents(a.Address, [][]byte{a.BlindingKey})
	if err != nil {
		errChan <- err
		return
	}
	var eventType EventType
	switch a.AccountIndex {
	case domain.FeeAccount:
		eventType = FeeAccountDeposit
	default:
		eventType = MarketAccountDeposit
	}
	event := AddressEvent{
		EventType:    eventType,
		AccountIndex: a.AccountIndex,
		Address:      a.Address,
		Utxos:        unspents,
	}
	eventChan <- event
}

func (a *AddressObservable) isEqual(observable Observable) bool {
	switch observable.(type) {
	case *AddressObservable:
		return a.equalsTo(observable.(*AddressObservable))
	default:
		return false
	}
}

func (a *AddressObservable) equalsTo(o *AddressObservable) bool {
	return a.AccountIndex == o.AccountIndex &&
		a.Address == o.Address &&
		bytes.Equal(a.BlindingKey, o.BlindingKey)
}

type TransactionObservable struct {
	TxID string
}

func (t *TransactionObservable) observe(
	w *sync.WaitGroup,
	explorerSvc explorer.Service,
	errChan chan error,
	eventChan chan Event,
) {
	defer w.Done()

	if t == nil {
		return
	}

	txStatus, err := explorerSvc.GetTransactionStatus(t.TxID)
	if err != nil {
		errChan <- err
		return
	}

	var confirmed bool
	var blockHash string
	var blockTime float64

	for k, v := range txStatus {
		switch value := v.(type) {
		case bool:
			if k == "confirmed" {
				confirmed = value
			}
		case string:
			if k == "block_hash" {
				blockHash = value
			}
		case float64:
			if k == "block_time" {
				blockTime = value
			}
		}

	}

	trxStatus := TransactionUnConfirmed
	if confirmed {
		trxStatus = TransactionConfirmed
	}

	event := TransactionEvent{
		TxID:      t.TxID,
		EventType: trxStatus,
		BlockHash: blockHash,
		BlockTime: blockTime,
	}

	eventChan <- event
}

func (t *TransactionObservable) isEqual(observable Observable) bool {
	switch observable.(type) {
	case *TransactionObservable:
		return t.equalsTo(observable.(*TransactionObservable))
	default:
		return false
	}
}

func (t *TransactionObservable) equalsTo(o *TransactionObservable) bool {
	return t.TxID == o.TxID
}
