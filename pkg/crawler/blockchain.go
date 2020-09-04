package crawler

import (
	"fmt"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
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
	EventType  int
	AccuntType int
	Address    string
	AssetHash  string
	Utxos      []explorer.Utxo
}

type Observable struct {
	AccountType int
	AssetHash   string
	Address     string
}

type utxoCrawler struct {
	explorerApiUrl string
	interval       *time.Ticker
	explorerSvc    explorer.Service
	errChan        chan error
	quitChan       chan int
	eventChan      chan event
	observables    []Observable
}

func NewService(
	explorerSvc explorer.Service,
	observables []Observable,
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
	}
}

func (u *utxoCrawler) Start() {
	fmt.Println("start observe")
	for {
		select {
		case <-u.interval.C:
			fmt.Println("observe interval")
			go u.observeAll(u.observables)
		case err := <-u.errChan:
			fmt.Println(err)
		case <-u.quitChan:
			fmt.Println("stop observe")
			close(u.eventChan)
			return
		}
	}
}

func (u *utxoCrawler) Stop() {
	u.quitChan <- 1
}

func (u *utxoCrawler) AddObservable(observable Observable) {
	u.observables = append(u.observables, observable)
}

func (u *utxoCrawler) RemoveObservable(observable Observable) {
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

func (u *utxoCrawler) observeAll(observables []Observable) {
	for _, a := range observables {
		go u.observe(a)
	}
}

func (u *utxoCrawler) observe(observe Observable) {
	unspents, err := u.explorerSvc.GetUnSpents(observe.Address)
	if err != nil {
		u.errChan <- err
	}
	var eventType int
	switch observe.AccountType {
	case wallet.FeeAccount:
		eventType = FeeAccountDeposit
	default:
		eventType = MarketAccountDeposit
	}
	event := event{
		EventType:  eventType,
		AccuntType: observe.AccountType,
		Address:    observe.Address,
		AssetHash:  observe.AssetHash,
		Utxos:      unspents,
	}
	u.eventChan <- event
}
