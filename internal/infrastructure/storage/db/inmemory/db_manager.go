package inmemory

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type marketInmemoryStore struct {
	markets         map[int]domain.Market
	accountsByAsset map[string]int
	locker          *sync.Mutex
}

type tradeInmemoryStore struct {
	trades               map[uuid.UUID]domain.Trade
	tradesBySwapAcceptID map[string]uuid.UUID
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

type InmemoryTx struct {
	db      *DbManager
	success bool
}

func (tx *InmemoryTx) Commit() error {
	if tx.db == nil {
		return errors.New("the transaction has no associated database.")
	}
	tx.success = true
	return nil
}

func (tx *InmemoryTx) Discard() {
	tx.success = false
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

func (db *DbManager) NewTransaction() ports.Transaction {
	return &InmemoryTx{
		db:      db,
		success: false,
	}
}

func (db *DbManager) NewUnspentsTransaction() ports.Transaction {
	return db.NewTransaction()
}

func (db *DbManager) NewPricesTransaction() ports.Transaction {
	return db.NewTransaction()
}

func (db *DbManager) RunTransaction(
	ctx context.Context,
	_ bool,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	return db.runTransaction(ctx, handler)
}

func (db *DbManager) RunUnspentsTransaction(
	ctx context.Context,
	readOnly bool,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	return db.RunTransaction(ctx, readOnly, handler)
}

func (db *DbManager) RunPricesTransaction(
	ctx context.Context,
	readOnly bool,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	return db.RunTransaction(ctx, readOnly, handler)
}

func (db *DbManager) runTransaction(
	ctx context.Context,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	res, err := handler(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}
