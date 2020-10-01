package crawler

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type Service interface {
	Start()
	Stop()
	AddObservable(observable Observable)
	RemoveObservable(observable Observable)
	GetEventChannel() chan Event
}

const (
	FeeAccountDeposit = iota
	MarketAccountDeposit
	TransactionConfirmed
	TransactionUnConfirmed
)

type Event interface {
	Type() int
}

type AddressEvent struct {
	EventType    int
	AccountIndex int
	Address      string
	Utxos        []explorer.Utxo
}

func (a AddressEvent) Type() int {
	return a.EventType
}

type TransactionEvent struct {
	TxID      string
	EventType int
	BlockHash string
	BlockTime float64
}

func (t TransactionEvent) Type() int {
	return t.EventType
}

type Observable interface {
	observe(
		w *sync.WaitGroup,
		explorerSvc explorer.Service,
		errChan chan error,
		eventChan chan Event,
	)
}

type AddressObservable struct {
	AccountIndex int
	AssetHash    string
	Address      string
	BlindingKey  []byte
}

type TransactionObservable struct {
	TxID string
}

type utxoCrawler struct {
	explorerApiUrl string
	interval       *time.Ticker
	explorerSvc    explorer.Service
	errChan        chan error
	quitChan       chan int
	eventChan      chan Event
	observables    []Observable
	errorHandler   func(err error)
	mutex          *sync.RWMutex
}

func NewService(
	explorerSvc explorer.Service,
	observables []Observable,
	errorHandler func(err error),
) Service {

	intervalInSeconds := config.GetInt(config.CrawlIntervalKey)
	interval := time.NewTicker(time.Duration(intervalInSeconds) * time.Second)

	return &utxoCrawler{
		explorerApiUrl: config.GetString(config.ExplorerEndpointKey),
		interval:       interval,
		explorerSvc:    explorerSvc,
		errChan:        make(chan error),
		quitChan:       make(chan int),
		eventChan:      make(chan Event),
		observables:    observables,
		errorHandler:   errorHandler,
		mutex:          &sync.RWMutex{},
	}
}

//Start starts crawler which periodically "scans" blockchain for specific
//events/Observable object
func (u *utxoCrawler) Start() {
	var wg sync.WaitGroup
	log.Debug("start observe")
	for {
		select {
		case <-u.interval.C:
			log.Debug("observe interval")
			u.observeAll(&wg)
		case err := <-u.errChan:
			u.errorHandler(err)
		case <-u.quitChan:
			log.Debug("stop observe")
			u.interval.Stop()
			wg.Wait()
			close(u.eventChan)
			return
		}
	}
}

//Stop stops crawler
func (u *utxoCrawler) Stop() {
	u.quitChan <- 1
}

func (u *utxoCrawler) getObservable() []Observable {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.observables
}

//AddObservable adds new Observable to the list of Observables to be "watched
//over"
func (u *utxoCrawler) AddObservable(observable Observable) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.observables = append(u.observables, observable)
}

//RemoveObservable stops "watching" given Observable
func (u *utxoCrawler) RemoveObservable(observable Observable) {
	observables := u.getObservable()

	u.mutex.Lock()

	switch obs := observable.(type) {
	case *AddressObservable:
		u.removeAddressObservable(*obs, observables)
	case *TransactionObservable:
		u.removeTransactionObservable(*obs, observables)
	}
	u.mutex.Unlock()
}

func (u *utxoCrawler) removeAddressObservable(
	observable AddressObservable,
	observables []Observable) {
	newObservableList := make([]Observable, 0)
	for _, obs := range observables {
		if o, ok := obs.(*AddressObservable); ok {
			if o.Address != observable.Address {
				newObservableList = append(newObservableList, o)
			}
		} else {
			newObservableList = append(newObservableList, o)
		}
	}
	u.observables = newObservableList
}

func (u *utxoCrawler) removeTransactionObservable(
	observable TransactionObservable,
	observables []Observable) {
	newObservableList := make([]Observable, 0)
	for _, obs := range observables {
		if o, ok := obs.(*TransactionObservable); ok {
			if o.TxID != observable.TxID {
				newObservableList = append(newObservableList, o)
			}
		} else {
			newObservableList = append(newObservableList, o)
		}
	}
	u.observables = newObservableList
}

//GetEventChannel returns Event channel which can be used to "listen" to
//blockchain events
func (u *utxoCrawler) GetEventChannel() chan Event {
	return u.eventChan
}

func (u *utxoCrawler) observeAll(w *sync.WaitGroup) {
	observables := u.getObservable()
	for _, o := range observables {
		w.Add(1)
		go o.observe(w, u.explorerSvc, u.errChan, u.eventChan)
	}
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
	}
	var eventType int
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

func (a *TransactionObservable) observe(
	w *sync.WaitGroup,
	explorerSvc explorer.Service,
	errChan chan error,
	eventChan chan Event,
) {
	defer w.Done()

	if a == nil {
		return
	}

	txStatus, err := explorerSvc.GetTransactionStatus(a.TxID)
	if err != nil {
		errChan <- err
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
		TxID:      a.TxID,
		EventType: trxStatus,
		BlockHash: blockHash,
		BlockTime: blockTime,
	}

	eventChan <- event
}
