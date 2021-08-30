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

type depositInmemoryStore struct {
	deposits map[domain.DepositKey]domain.Deposit
	locker   *sync.RWMutex
}

type withdrawalInmemoryStore struct {
	withdrawals map[string]domain.Withdrawal
	locker      *sync.RWMutex
}

type RepoManager struct {
	marketStore     *marketInmemoryStore
	tradeStore      *tradeInmemoryStore
	unspentStore    *unspentInmemoryStore
	vaultStore      *vaultInmemoryStore
	depositStore    *depositInmemoryStore
	withdrawalStore *withdrawalInmemoryStore

	marketRepository      domain.MarketRepository
	unspentRepository     domain.UnspentRepository
	tradeRepository       domain.TradeRepository
	vaultRepository       domain.VaultRepository
	depositRepository     domain.DepositRepository
	withdrawalsRepository domain.WithdrawalRepository
}

type InmemoryTx struct {
	db      *RepoManager
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

func NewRepoManager() ports.RepoManager {
	marketStore := &marketInmemoryStore{
		markets:         map[int]domain.Market{},
		accountsByAsset: map[string]int{},
		locker:          &sync.Mutex{},
	}
	tradeStore := &tradeInmemoryStore{
		trades:               map[uuid.UUID]domain.Trade{},
		tradesBySwapAcceptID: map[string]uuid.UUID{},
		tradesByMarket:       map[string][]uuid.UUID{},
		locker:               &sync.Mutex{},
	}
	unspentStore := &unspentInmemoryStore{
		unspents: map[domain.UnspentKey]domain.Unspent{},
		locker:   &sync.RWMutex{},
	}
	vaultStore := &vaultInmemoryStore{
		vault:  &domain.Vault{},
		locker: &sync.Mutex{},
	}
	depositStore := &depositInmemoryStore{
		deposits: map[domain.DepositKey]domain.Deposit{},
		locker:   &sync.RWMutex{},
	}
	withdrawalStore := &withdrawalInmemoryStore{
		withdrawals: map[string]domain.Withdrawal{},
		locker:      &sync.RWMutex{},
	}

	marketRepo := NewMarketRepositoryImpl(marketStore)
	tradeRepo := NewTradeRepositoryImpl(tradeStore)
	unspentRepo := NewUnspentRepositoryImpl(unspentStore)
	vaultRepo := NewVaultRepositoryImpl(vaultStore)
	depositRepo := NewDepositRepositoryImpl(depositStore)
	withdrawalRepo := NewWithdrawalRepositoryImpl(withdrawalStore)

	return &RepoManager{
		marketStore:           marketStore,
		tradeStore:            tradeStore,
		unspentStore:          unspentStore,
		vaultStore:            vaultStore,
		depositStore:          depositStore,
		withdrawalStore:       withdrawalStore,
		marketRepository:      marketRepo,
		tradeRepository:       tradeRepo,
		unspentRepository:     unspentRepo,
		vaultRepository:       vaultRepo,
		depositRepository:     depositRepo,
		withdrawalsRepository: withdrawalRepo,
	}
}

func (d *RepoManager) MarketRepository() domain.MarketRepository {
	return d.marketRepository
}

func (d *RepoManager) UnspentRepository() domain.UnspentRepository {
	return d.unspentRepository
}

func (d *RepoManager) TradeRepository() domain.TradeRepository {
	return d.tradeRepository
}

func (d *RepoManager) VaultRepository() domain.VaultRepository {
	return d.vaultRepository
}

func (d *RepoManager) DepositRepository() domain.DepositRepository {
	return d.depositRepository
}

func (d *RepoManager) WithdrawalRepository() domain.WithdrawalRepository {
	return d.withdrawalsRepository
}

func (d *RepoManager) Close() {}

func (db *RepoManager) NewTransaction() ports.Transaction {
	return &InmemoryTx{
		db:      db,
		success: false,
	}
}

func (db *RepoManager) NewUnspentsTransaction() ports.Transaction {
	return db.NewTransaction()
}

func (db *RepoManager) NewPricesTransaction() ports.Transaction {
	return db.NewTransaction()
}

func (db *RepoManager) RunTransaction(
	ctx context.Context,
	_ bool,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	return db.runTransaction(ctx, handler)
}

func (db *RepoManager) RunUnspentsTransaction(
	ctx context.Context,
	readOnly bool,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	return db.RunTransaction(ctx, readOnly, handler)
}

func (db *RepoManager) RunPricesTransaction(
	ctx context.Context,
	readOnly bool,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	return db.RunTransaction(ctx, readOnly, handler)
}

func (db *RepoManager) runTransaction(
	ctx context.Context,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	res, err := handler(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}
