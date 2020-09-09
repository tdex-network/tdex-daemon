package crawler

import (
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/constant"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"sync"
	"time"
)

type Service interface {
	Start()
	Stop()
	AddObservable(observable Observable)
	RemoveObservable(observable Observable)
	GetEventChannel() chan event
}

const (
	FeeAccountDeposit = iota
	MarketAccountDeposit
)

type event struct {
	EventType   int
	AccountType int
	Address     string
	AssetHash   string
	Utxos       []explorer.Utxo
}

type Observable struct {
	AccountType int
	AssetHash   string
	Address     string
	BlindingKey [][]byte
}

type utxoCrawler struct {
	explorerApiUrl string
	interval       *time.Ticker
	explorerSvc    explorer.Service
	errChan        chan error
	quitChan       chan int
	eventChan      chan event
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
		eventChan:      make(chan event),
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
	u.mutex.Lock()
	defer u.mutex.Unlock()
	newObservableList := make([]Observable, 0)
	for _, o := range u.observables {
		if o.Address != observable.Address {
			newObservableList = append(newObservableList, o)
		}
	}
	u.observables = newObservableList
}

func (u *utxoCrawler) GetEventChannel() chan event {
	return u.eventChan
}

func (u *utxoCrawler) observeAll(w *sync.WaitGroup) {
	observables := u.getObservable()
	for _, a := range observables {
		w.Add(1)
		go u.observe(a, w)
	}
}

func (u *utxoCrawler) observe(observe Observable, w *sync.WaitGroup) {
	defer w.Done()
	unspents, err := u.explorerSvc.GetUnSpents(observe.Address, observe.BlindingKey)
	if err != nil {
		u.errChan <- err
	}
	var eventType int
	switch observe.AccountType {
	case constant.FeeAccount:
		eventType = FeeAccountDeposit
	default:
		eventType = MarketAccountDeposit
	}
	event := event{
		EventType:   eventType,
		AccountType: observe.AccountType,
		Address:     observe.Address,
		AssetHash:   observe.AssetHash,
		Utxos:       unspents,
	}
	u.eventChan <- event
}
