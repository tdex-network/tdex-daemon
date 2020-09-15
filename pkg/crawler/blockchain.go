package crawler

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
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
	EventType   int
	AccountType int
	Address     string
	AssetHash   string
	Utxos       []explorer.Utxo
}

func (a AddressEvent) Type() int {
	return a.EventType
}

type TransactionEvent struct {
	TxID      string
	EventType int
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
	AccountType int
	AssetHash   string
	Address     string
	BlindingKey [][]byte
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

func (u *utxoCrawler) Stop() {
	u.quitChan <- 1
}

func (u *utxoCrawler) getObservable() []Observable {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.observables
}

func (u *utxoCrawler) AddObservable(observable Observable) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.observables = append(u.observables, observable)
}

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
	//observables := u.getObservable()
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
	//observables := u.getObservable()
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

	unspents, err := explorerSvc.GetUnSpents(a.Address, a.BlindingKey)
	if err != nil {
		errChan <- err
	}
	var eventType int
	switch a.AccountType {
	case vault.FeeAccount:
		eventType = FeeAccountDeposit
	default:
		eventType = MarketAccountDeposit
	}
	event := AddressEvent{
		EventType:   eventType,
		AccountType: a.AccountType,
		Address:     a.Address,
		AssetHash:   a.AssetHash,
		Utxos:       unspents,
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

	confirmed, err := explorerSvc.IsTransactionConfirmed(a.TxID)
	if err != nil {
		errChan <- err
	}

	trxStatus := TransactionUnConfirmed
	if confirmed {
		trxStatus = TransactionConfirmed
	}

	event := TransactionEvent{
		TxID:      a.TxID,
		EventType: trxStatus,
	}

	eventChan <- event
}
