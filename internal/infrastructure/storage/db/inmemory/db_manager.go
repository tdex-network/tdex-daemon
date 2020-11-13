package inmemory

import (
	"sync"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

type marketInmemoryStore struct {
	markets         map[int]domain.Market
	accountsByAsset map[string]int
	locker          *sync.Mutex
}

type tradeInmemoryStore struct {
	trades               map[uuid.UUID]domain.Trade
	tradesBySwapAcceptID map[string]uuid.UUID
	tradesByTrader       map[string][]uuid.UUID
	tradesByMarket       map[string][]uuid.UUID
	locker               *sync.Mutex
}

type unspentInmemoryStore struct {
	unspents map[domain.UnspentKey]domain.Unspent
	locker   *sync.RWMutex
}

type vaultInmemoryStore struct {
	vault  *domain.Vault
	locker *sync.Mutex
}

type DbManager struct {
	marketStore  *marketInmemoryStore
	tradeStore   *tradeInmemoryStore
	unspentStore *unspentInmemoryStore
	vaultStore   *vaultInmemoryStore
}

func NewDbManager() *DbManager {
	return &DbManager{
		marketStore: &marketInmemoryStore{
			markets:         map[int]domain.Market{},
			accountsByAsset: map[string]int{},
			locker:          &sync.Mutex{},
		},
		tradeStore: &tradeInmemoryStore{
			trades:               map[uuid.UUID]domain.Trade{},
			tradesBySwapAcceptID: map[string]uuid.UUID{},
			tradesByTrader:       map[string][]uuid.UUID{},
			tradesByMarket:       map[string][]uuid.UUID{},
			locker:               &sync.Mutex{},
		},
		unspentStore: &unspentInmemoryStore{
			unspents: map[domain.UnspentKey]domain.Unspent{},
			locker:   &sync.RWMutex{},
		},
		vaultStore: &vaultInmemoryStore{
			vault:  &domain.Vault{},
			locker: &sync.Mutex{},
		},
	}
}
