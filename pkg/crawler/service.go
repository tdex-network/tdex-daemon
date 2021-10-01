package crawler

import (
	"sync"
	"time"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

const (
	eventQueueMaxSize = 100
	errorQueueMaxSize = 10
)

type blockchainCrawler struct {
	interval     time.Duration
	explorerSvc  explorer.Service
	errChan      chan error
	eventChan    chan Event
	observables  map[string]*ObservableHandler
	errorHandler func(err error)
	mutex        *sync.RWMutex
	wg           *sync.WaitGroup
}

// Opts defines the parameters needed for creating a crawler service with NewService method
type Opts struct {
	ExplorerSvc explorer.Service
	// Crawler interval in milliseconds
	CrawlerInterval time.Duration
	ErrorHandler    func(err error)
}

// NewService returns an utxoCrawelr that is ready for watch for blockchain activites. Use Start and Stop methods to manage it.
func NewService(opts Opts) Service {
	return &blockchainCrawler{
		interval:     opts.CrawlerInterval,
		explorerSvc:  opts.ExplorerSvc,
		errChan:      make(chan error, errorQueueMaxSize),
		eventChan:    make(chan Event, eventQueueMaxSize),
		observables:  map[string]*ObservableHandler{},
		errorHandler: opts.ErrorHandler,
		mutex:        &sync.RWMutex{},
		wg:           &sync.WaitGroup{},
	}
}

// Start starts crawler which periodically "scans" blockchain for specific
// events/Observable object
func (bc *blockchainCrawler) Start() {
	for {
		err, more := <-bc.errChan
		if !more {
			return
		}
		go bc.errorHandler(err)
	}
}

// Stop stops crawler
func (bc *blockchainCrawler) Stop() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	for _, obsHandler := range bc.observables {
		go obsHandler.Stop()
	}
	bc.wg.Wait()
	bc.eventChan <- CloseEvent{}
	close(bc.errChan)
	return
}

// GetEventChannel returns Event channel which can be used to "listen" to
// blockchain events
func (bc *blockchainCrawler) GetEventChannel() chan Event {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	return bc.eventChan
}

// AddObservable adds new Observable to the list of Observables to be "watched
// over" only if the same Observable is not already in the list
func (bc *blockchainCrawler) AddObservable(observable Observable) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if _, ok := bc.observables[observable.Key()]; !ok {
		obsHandler := NewObservableHandler(
			observable,
			bc.explorerSvc,
			bc.wg,
			bc.interval,
			bc.eventChan,
			bc.errChan,
		)

		bc.observables[observable.Key()] = obsHandler
		go obsHandler.Start()
	}
}

// RemoveObservable stops "watching" given Observable
func (bc *blockchainCrawler) RemoveObservable(observable Observable) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if obsHandler, ok := bc.observables[observable.Key()]; ok {
		obsHandler.Stop()
		delete(bc.observables, observable.Key())
	}
}
