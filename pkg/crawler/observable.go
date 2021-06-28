package crawler

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"

	log "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

const (
	New       Status = "NEW"
	Waiting   Status = "WAITING"
	Processed Status = "PROCESSED"
)

type AddressObservable struct {
	AccountIndex int
	Address      string
	BlindingKey  []byte
	cb           *gobreaker.CircuitBreaker
}

func NewAddressObservable(
	accountIndex int, address string, blindKey []byte,
) Observable {
	return &AddressObservable{
		AccountIndex: accountIndex,
		Address:      address,
		BlindingKey:  blindKey,
		cb:           newCircuitBreaker(),
	}
}

func (a *AddressObservable) Observe(
	explorerSvc explorer.Service,
	errChan chan error,
	eventChan chan Event,
	observableStatus *observableStatus,
	rateLimiter *rate.Limiter,
) {
	if a == nil {
		return
	}

	observableStatus.Set(Waiting)
	if err := rateLimiter.Wait(context.Background()); err != nil {
		errChan <- err
		return
	}

	iUnspents, err := a.cb.Execute(func() (interface{}, error) {
		return explorerSvc.GetUnspents(a.Address, [][]byte{a.BlindingKey})
	})
	if err != nil {
		errChan <- err
		return
	}
	unspents := iUnspents.([]explorer.Utxo)

	observableStatus.Set(Processed)

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

func (a *AddressObservable) Key() string {
	return a.Address
}

type TransactionObservable struct {
	TxID  string
	TxHex string
	cb    *gobreaker.CircuitBreaker
}

func NewTransactionObservable(txid string) Observable {
	return &TransactionObservable{
		TxID: txid,
		cb:   newCircuitBreaker(),
	}
}

func (t *TransactionObservable) Observe(
	explorerSvc explorer.Service,
	errChan chan error,
	eventChan chan Event,
	observableStatus *observableStatus,
	rateLimiter *rate.Limiter,
) {
	if t == nil {
		return
	}

	observableStatus.Set(Waiting)
	if err := rateLimiter.Wait(context.Background()); err != nil {
		errChan <- err
		return
	}

	iTxStatus, err := t.cb.Execute(func() (interface{}, error) {
		return explorerSvc.GetTransactionStatus(t.TxID)
	})
	if err != nil {
		errChan <- err
		return
	}
	txStatus := iTxStatus.(map[string]interface{})

	if len(t.TxHex) <= 0 {

		iTxHex, err := t.cb.Execute(func() (interface{}, error) {
			return explorerSvc.GetTransactionHex(t.TxID)
		})
		if err == nil {
			t.TxHex = iTxHex.(string)
		}
	}

	observableStatus.Set(Processed)

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
		TxHex:     t.TxHex,
		EventType: trxStatus,
		BlockHash: blockHash,
		BlockTime: blockTime,
	}

	eventChan <- event
}

func (t *TransactionObservable) Key() string {
	return t.TxID
}

type Status string

type observableStatus struct {
	sync.RWMutex
	status Status
}

func newObservableStatus() *observableStatus {
	return &observableStatus{
		status: New,
	}
}

func (o *observableStatus) Get() Status {
	o.RLock()
	defer o.RUnlock()
	return o.status
}

func (o *observableStatus) Set(status Status) {
	o.Lock()
	defer o.Unlock()
	o.status = status
}

type observableHandler struct {
	observable       Observable
	explorerSvc      explorer.Service
	wg               *sync.WaitGroup
	ticker           *time.Ticker
	eventChan        chan Event
	errChan          chan error
	stopChan         chan int
	observableStatus *observableStatus
	rateLimiter      *rate.Limiter
}

func newObservableHandler(
	observable Observable,
	explorerSvc explorer.Service,
	wg *sync.WaitGroup,
	interval time.Duration,
	eventChan chan Event,
	errChan chan error,
	rateLimiter *rate.Limiter,
) *observableHandler {
	ticker := time.NewTicker(interval)
	stopChan := make(chan int, 1)

	return &observableHandler{
		observable,
		explorerSvc,
		wg,
		ticker,
		eventChan,
		errChan,
		stopChan,
		newObservableStatus(),
		rateLimiter,
	}
}

func (oh *observableHandler) start() {
	oh.logAction("start")
	oh.wg.Add(1)
	for {
		select {
		case <-oh.ticker.C:
			if oh.observableStatus.Get() != Waiting {
				oh.observable.Observe(
					oh.explorerSvc,
					oh.errChan,
					oh.eventChan,
					oh.observableStatus,
					oh.rateLimiter,
				)
			}
		case <-oh.stopChan:
			oh.ticker.Stop()
			close(oh.stopChan)
			return
		}
	}
}

func (oh *observableHandler) stop() {
	oh.logAction("stop")
	oh.stopChan <- 1
	oh.wg.Done()
}

func (oh *observableHandler) logAction(action string) {
	obs := oh.observable
	switch obs.(type) {
	case *AddressObservable:
		log.Debugf("%s observing address: %v", action, obs.Key())
	case *TransactionObservable:
		log.Debugf("%s observing tx: %v", action, obs.Key())
	}
}

func newCircuitBreaker() *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "observable",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests > 20 && failureRatio >= 0.7
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			if to == gobreaker.StateOpen {
				log.Debug("cannot complete observation, will retry later")
			}
			if from == gobreaker.StateOpen && to == gobreaker.StateHalfOpen {
				log.Debug("check observation status")
			}
			if from == gobreaker.StateHalfOpen && to == gobreaker.StateClosed {
				log.Debug("restart observation")
			}
		},
	})
}
