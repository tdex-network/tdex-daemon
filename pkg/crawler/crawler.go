package crawler

import (
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

// Event are emitted through a channel during observation.
type Event interface {
	Type() EventType
}

// Observable represent object that can be observe on the blockchain.
type Observable interface {
	Observe(
		explorerSvc explorer.Service,
		errChan chan error,
		eventChan chan Event,
	)
	Key() string
}

// Service is the interface for Crawler
type Service interface {
	Start()
	Stop()
	AddObservable(observable Observable)
	RemoveObservable(observable Observable)
	GetEventChannel() chan Event
}
