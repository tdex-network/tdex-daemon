package crawler

import (
	"sync"
	"time"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"golang.org/x/time/rate"
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
	observables  map[string]*observableHandler
	errorHandler func(err error)
	mutex        *sync.RWMutex
	wg           *sync.WaitGroup
	rateLimiter  *rate.Limiter
}

// Opts defines the parameters needed for creating a crawler service with NewService method
type Opts struct {
	ExplorerSvc explorer.Service
	//crawler interval in milliseconds
	CrawlerInterval time.Duration
	//number of requests per second
	ExplorerLimit int
	//number of bursts tokens permitted
	ExplorerTokenBurst int
	ErrorHandler       func(err error)
}

// NewService returns an utxoCrawelr that is ready for watch for blockchain activites. Use Start and Stop methods to manage it.
func NewService(opts Opts) Service {
	oneSecInMillisecond := 1000
	everyMillisecond := oneSecInMillisecond / opts.ExplorerLimit
	rt := rate.Every(time.Duration(everyMillisecond) * time.Millisecond)
	rateLimiter := rate.NewLimiter(rt, opts.ExplorerTokenBurst)

	return &blockchainCrawler{
		interval:     opts.CrawlerInterval,
		explorerSvc:  opts.ExplorerSvc,
		errChan:      make(chan error, errorQueueMaxSize),
		eventChan:    make(chan Event, eventQueueMaxSize),
		observables:  map[string]*observableHandler{},
		errorHandler: opts.ErrorHandler,
		mutex:        &sync.RWMutex{},
		wg:           &sync.WaitGroup{},
		rateLimiter:  rateLimiter,
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
		go obsHandler.stop()
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
		obsHandler := newObservableHandler(
			observable,
			bc.explorerSvc,
			bc.wg,
			bc.interval,
			bc.eventChan,
			bc.errChan,
			bc.rateLimiter,
		)

		bc.observables[observable.Key()] = obsHandler
		go obsHandler.start()
	}
}

// RemoveObservable stops "watching" given Observable
func (bc *blockchainCrawler) RemoveObservable(observable Observable) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if obsHandler, ok := bc.observables[observable.Key()]; ok {
		obsHandler.stop()
		delete(bc.observables, observable.Key())
	}
}

//IsObservingAddresses returns true if the crawler is observing at least one address given as parameter.
//false in the other case
func (bc *blockchainCrawler) IsObservingAddresses(addresses []string) bool {
	if len(addresses) == 0 {
		return false
	}

	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	for _, addr := range addresses {
		if _, ok := bc.observables[addr]; !ok {
			return false
		}
	}
	return true
}
