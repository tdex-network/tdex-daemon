package crawler

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

// Event are emitted through a channel during observation.
type Event interface {
	Type() EventType
}

// Observable represent object that can be observe on the blockchain.
type Observable interface {
	observe(
		w *sync.WaitGroup,
		explorerSvc explorer.Service,
		errChan chan error,
		eventChan chan Event,
	)
	isEqual(observable Observable) bool
}

// Service is the interface for Crawler
type Service interface {
	Start()
	Stop()
	AddObservable(observable Observable)
	RemoveObservable(observable Observable)
	IsObservingAddresses(addresses []string) bool
	GetEventChannel() chan Event
}

type utxoCrawler struct {
	interval     *time.Ticker
	explorerSvc  explorer.Service
	errChan      chan error
	quitChan     chan int
	eventChan    chan Event
	observables  []Observable
	errorHandler func(err error)
	mutex        *sync.RWMutex
}

// Opts defines the parameters needed for creating a crawler service with NewService method
type Opts struct {
	ExplorerSvc            explorer.Service
	IntervalInMilliseconds int
	Observables            []Observable
	ErrorHandler           func(err error)
}

// NewService returns an utxoCrawelr that is ready for watch for blockchain activites. Use Start and Stop methods to manage it.
func NewService(opts Opts) Service {

	interval := time.NewTicker(time.Duration(opts.IntervalInMilliseconds) * time.Millisecond)

	return &utxoCrawler{
		interval:     interval,
		explorerSvc:  opts.ExplorerSvc,
		errChan:      make(chan error),
		quitChan:     make(chan int),
		eventChan:    make(chan Event),
		observables:  opts.Observables,
		errorHandler: opts.ErrorHandler,
		mutex:        &sync.RWMutex{},
	}
}

// Start starts crawler which periodically "scans" blockchain for specific
// events/Observable object
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

// Stop stops crawler
func (u *utxoCrawler) Stop() {
	u.quitChan <- 1
}

// GetEventChannel returns Event channel which can be used to "listen" to
// blockchain events
func (u *utxoCrawler) GetEventChannel() chan Event {
	return u.eventChan
}

// AddObservable adds new Observable to the list of Observables to be "watched
// over" only if the same Observable is not already in the list
func (u *utxoCrawler) AddObservable(observable Observable) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if !contains(u.observables, observable) {
		obs, ok := observable.(*AddressObservable)
		if ok {
			log.Debug("Start observing new account: " + fmt.Sprint(obs.AccountIndex))
		}

		u.observables = append([]Observable{observable}, u.observables...)
	}
}

// RemoveObservable stops "watching" given Observable
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

//IsObservingAddresses returns true if the crawler is observing at least one address given as parameter.
//false in the other case
func (u *utxoCrawler) IsObservingAddresses(addresses []string) bool {
	if len(addresses) == 0 {
		return false
	}
	observables := u.getObservable()
	for _, observable := range observables {
		switch observable := observable.(type) {
		case *AddressObservable:
			for _, addr := range addresses {
				if observable.Address == addr {
					return true
				}
			}
			continue
		default:
			continue
		}
	}
	return false
}

func (u *utxoCrawler) observeAll(w *sync.WaitGroup) {
	observables := u.getObservable()
	for _, o := range observables {
		w.Add(1)
		go o.observe(w, u.explorerSvc, u.errChan, u.eventChan)
	}
}

func (u *utxoCrawler) getObservable() []Observable {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.observables
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

func contains(observables []Observable, observable Observable) bool {
	for _, o := range observables {
		if o.isEqual(observable) {
			return true
		}
	}
	return false
}
