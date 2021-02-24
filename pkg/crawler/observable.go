package crawler

import (
	"sync"
	"time"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type AddressObservable struct {
	AccountIndex int
	Address      string
	BlindingKey  []byte
}

func (a *AddressObservable) observe(
	explorerSvc explorer.Service,
	errChan chan error,
	eventChan chan Event,
) {
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

func (a *AddressObservable) key() string {
	return a.Address
}

type TransactionObservable struct {
	TxID string
}

func (t *TransactionObservable) observe(
	explorerSvc explorer.Service,
	errChan chan error,
	eventChan chan Event,
) {
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

func (t *TransactionObservable) key() string {
	return t.TxID
}

type observableHandler struct {
	observable  Observable
	explorerSvc explorer.Service
	wg          *sync.WaitGroup
	ticker      *time.Ticker
	eventChan   chan Event
	errChan     chan error
	stopChan    chan int
}

func newObservableHandler(
	observable Observable,
	explorerSvc explorer.Service,
	wg *sync.WaitGroup,
	interval int,
	eventChan chan Event,
	errChan chan error,
) *observableHandler {
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	stopChan := make(chan int, 1)
	return &observableHandler{
		observable,
		explorerSvc,
		wg,
		ticker,
		eventChan,
		errChan,
		stopChan,
	}
}

func (oh *observableHandler) start() {
	oh.wg.Add(1)
	for {
		select {
		case <-oh.ticker.C:
			oh.observable.observe(oh.explorerSvc, oh.errChan, oh.eventChan)
		case <-oh.stopChan:
			oh.ticker.Stop()
			close(oh.stopChan)
			return
		}
	}
}

func (oh *observableHandler) stop() {
	oh.stopChan <- 1
	oh.wg.Done()
}
