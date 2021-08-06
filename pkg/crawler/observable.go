package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"

	log "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"github.com/tdex-network/tdex-daemon/pkg/circuitbreaker"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

const (
	New       Status = "NEW"
	Waiting   Status = "WAITING"
	Processed Status = "PROCESSED"
)

type AddressObservable struct {
	Address     string
	BlindingKey []byte
	ExtraData   interface{}
	cb          *gobreaker.CircuitBreaker
}

func NewAddressObservable(
	address string, blindKey []byte, extraData interface{},
) Observable {
	return &AddressObservable{
		Address:     address,
		BlindingKey: blindKey,
		ExtraData:   extraData,
		cb:          circuitbreaker.NewCircuitBreaker(),
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
	defer observableStatus.Set(Processed)
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

	event := AddressEvent{
		EventType: AddressUnspents,
		ExtraData: a.ExtraData,
		Address:   a.Address,
		Utxos:     unspents,
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
		cb:   circuitbreaker.NewCircuitBreaker(),
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
	defer observableStatus.Set(Processed)
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
	txStatus := iTxStatus.(explorer.TransactionStatus)

	if len(t.TxHex) <= 0 {
		iTxHex, err := t.cb.Execute(func() (interface{}, error) {
			return explorerSvc.GetTransactionHex(t.TxID)
		})
		if err == nil {
			t.TxHex = iTxHex.(string)
		}
	}

	trxStatus := TransactionUnconfirmed
	if txStatus.Confirmed() {
		trxStatus = TransactionConfirmed
	}

	event := TransactionEvent{
		TxID:      t.TxID,
		TxHex:     t.TxHex,
		EventType: trxStatus,
		BlockHash: txStatus.BlockHash(),
		BlockTime: txStatus.BlockTime(),
	}

	eventChan <- event
}

func (t *TransactionObservable) Key() string {
	return t.TxID
}

type Outpoint interface {
	Hash() string
	Index() uint32
}
type OutpointsObservable struct {
	Outpoints []Outpoint
	ExtraData interface{}
	cb        *gobreaker.CircuitBreaker
}

func NewOutpointsObservable(
	outpoints []Outpoint, extraData interface{},
) Observable {
	return &OutpointsObservable{
		Outpoints: outpoints,
		ExtraData: extraData,
		cb:        circuitbreaker.NewCircuitBreaker(),
	}
}

func (o *OutpointsObservable) Observe(
	explorerSvc explorer.Service,
	errChan chan error,
	eventChan chan Event,
	observableStatus *observableStatus,
	rateLimiter *rate.Limiter,
) {
	if o == nil {
		return
	}

	observableStatus.Set(Waiting)
	defer observableStatus.Set(Processed)
	if err := rateLimiter.Wait(context.Background()); err != nil {
		errChan <- err
		return
	}

	numOuts := len(o.Outpoints)
	statuses := make([]bool, 0)
	var txHash string

	for _, outpoint := range o.Outpoints {
		iRes, err := o.cb.Execute(func() (interface{}, error) {
			return explorerSvc.GetUnspentStatus(
				outpoint.Hash(), outpoint.Index(),
			)
		})
		if err != nil {
			errChan <- err
			return
		}
		unspentStatus := iRes.(explorer.UtxoStatus)

		if unspentStatus.Spent() {
			statuses = append(statuses, true)
			if txHash == "" {
				txHash = unspentStatus.Hash()
			}
		}
	}
	if len(statuses) != numOuts {
		eventChan <- OutpointsEvent{
			EventType: OutpointsUnspent,
			Outpoints: o.Outpoints,
			ExtraData: o.ExtraData,
		}
		return
	}

	iRes, err := o.cb.Execute(func() (interface{}, error) {
		return explorerSvc.GetTransactionHex(txHash)
	})
	if err != nil {
		errChan <- err
		return
	}
	txHex := iRes.(string)

	iTxStatus, err := o.cb.Execute(func() (interface{}, error) {
		return explorerSvc.GetTransactionStatus(txHash)
	})
	if err != nil {
		errChan <- err
		return
	}
	txStatus := iTxStatus.(explorer.TransactionStatus)
	eventType := OutpointsSpentAndUnconfirmed
	if txStatus.Confirmed() {
		eventType = OutpointsSpentAndConfirmed
	}

	eventChan <- OutpointsEvent{
		EventType: eventType,
		Outpoints: o.Outpoints,
		ExtraData: o.ExtraData,
		TxID:      txHash,
		TxHex:     txHex,
		BlockHash: txStatus.BlockHash(),
		BlockTime: txStatus.BlockTime(),
	}
}

func (o *OutpointsObservable) Key() string {
	str := ""
	for _, out := range o.Outpoints {
		str += fmt.Sprintf("%s:%d", out.Hash(), out.Index())
	}
	buf := sha256.Sum256([]byte(str))

	return hex.EncodeToString(buf[:])
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

type ObservableHandler struct {
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

func NewObservableHandler(
	observable Observable,
	explorerSvc explorer.Service,
	wg *sync.WaitGroup,
	interval time.Duration,
	eventChan chan Event,
	errChan chan error,
	rateLimiter *rate.Limiter,
) *ObservableHandler {
	ticker := time.NewTicker(interval)
	stopChan := make(chan int, 1)

	return &ObservableHandler{
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

func (oh *ObservableHandler) Start() {
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

func (oh ObservableHandler) Stop() {
	oh.logAction("stop")
	oh.stopChan <- 1
	oh.wg.Done()
}

func (oh *ObservableHandler) logAction(action string) {
	obs := oh.observable
	switch obs.(type) {
	case *AddressObservable:
		log.Debugf("%s observing address: %v", action, obs.Key())
	case *TransactionObservable:
		log.Debugf("%s observing tx: %v", action, obs.Key())
	}
}
