package storage

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/domain/trade"
	"github.com/tdex-network/tdex-daemon/internal/storageutil/uow"
)

// InMemoryTradeRepository represents an in memory storage
type InMemoryTradeRepository struct {
	trades               map[uuid.UUID]*trade.Trade
	tradesBySwapAcceptID map[string]uuid.UUID
	tradesByTrader       map[string][]uuid.UUID
	tradesByMarket       map[string][]uuid.UUID

	lock *sync.RWMutex
}

// NewInMemoryTradeRepository returns a new empty InMemoryTradeRepository
func NewInMemoryTradeRepository() InMemoryTradeRepository {
	return InMemoryTradeRepository{
		trades:               map[uuid.UUID]*trade.Trade{},
		tradesBySwapAcceptID: map[string]uuid.UUID{},
		tradesByTrader:       map[string][]uuid.UUID{},
		tradesByMarket:       map[string][]uuid.UUID{},
		lock:                 &sync.RWMutex{},
	}
}

// GetOrCreateTrade gets a trade with a given swapID that can be either a
// request, accept, or complete ID. They all identify the same Trade.
// If not found, a new entry is inserted
func (r InMemoryTradeRepository) GetOrCreateTrade(ctx context.Context, tradeID *uuid.UUID) (*trade.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	trades, _, _, _ := r.storageByContext(ctx)
	return getOrCreateTrade(trades, tradeID)
}

// GetAllTrades returns all the trades processed by the daemon
func (r InMemoryTradeRepository) GetAllTrades(ctx context.Context) ([]*trade.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	trades, _, _, _ := r.storageByContext(ctx)
	return getAllTrades(trades)
}

// GetAllTradesByMarket returns all the trades processed for the given market
func (r InMemoryTradeRepository) GetAllTradesByMarket(ctx context.Context, marketQuoteAsset string) ([]*trade.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	trades, _, tradesByMarket, _ := r.storageByContext(ctx)
	return getAllTradesByMarket(trades, tradesByMarket, marketQuoteAsset)
}

// GetAllTradesByTrader returns all the trades processed for the given trader
func (r InMemoryTradeRepository) GetAllTradesByTrader(ctx context.Context, traderID string) ([]*trade.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	trades, _, _, tradesByTrader := r.storageByContext(ctx)
	return getAllTradesByTrader(trades, tradesByTrader, traderID)
}

// UpdateTrade updates data to a trade identified by any of its swap ids (request, accept, complete) passing an update function
func (r InMemoryTradeRepository) UpdateTrade(
	ctx context.Context,
	tradeID *uuid.UUID,
	updateFn func(t *trade.Trade) (*trade.Trade, error),
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	trades, tradesBySwapAcceptID, tradesByTrader, tradesByMarket := r.storageByContext(ctx)

	currentTrade, err := getOrCreateTrade(trades, tradeID)
	if err != nil {
		return err
	}

	updatedTrade, err := updateFn(currentTrade)
	if err != nil {
		return err
	}

	if swapAccept := updatedTrade.SwapAcceptMessage(); swapAccept != nil {
		if _, ok := tradesBySwapAcceptID[swapAccept.GetId()]; !ok {
			tradesBySwapAcceptID[swapAccept.GetId()] = currentTrade.ID()
		}
	}

	addTradeIDByKeyString(tradesByMarket, updatedTrade.MarketQuoteAsset(), currentTrade.ID())
	addTradeIDByKeyString(tradesByTrader, hex.EncodeToString(updatedTrade.TraderPubkey()), currentTrade.ID())
	return nil
}

// Begin returns a new InMemoryTradeRepositoryTx
func (r InMemoryTradeRepository) Begin() (uow.Tx, error) {
	tx := &InMemoryTradeRepositoryTx{
		root:                 r,
		trades:               map[uuid.UUID]*trade.Trade{},
		tradesBySwapAcceptID: map[string]uuid.UUID{},
		tradesByMarket:       map[string][]uuid.UUID{},
		tradesByTrader:       map[string][]uuid.UUID{},
	}

	// copy the current state of the repo into the transaction
	for k, v := range r.trades {
		tx.trades[k] = v
	}
	for k, v := range r.tradesBySwapAcceptID {
		tx.tradesBySwapAcceptID[k] = v
	}
	for k, v := range r.tradesByTrader {
		cv := make([]uuid.UUID, len(v))
		copy(cv, v)
		tx.tradesByTrader[k] = cv
	}
	for k, v := range r.tradesByMarket {
		cv := make([]uuid.UUID, len(v))
		copy(cv, v)
		tx.tradesByMarket[k] = cv
	}

	return tx, nil
}

// ContextKey returns the context key shared between in-memory repositories
func (r InMemoryTradeRepository) ContextKey() interface{} {
	return uow.InMemoryContextKey
}

func (r InMemoryTradeRepository) storageByContext(ctx context.Context) (
	trades map[uuid.UUID]*trade.Trade,
	tradesBySwapAcceptID map[string]uuid.UUID,
	tradesByMarket map[string][]uuid.UUID,
	tradesByTrader map[string][]uuid.UUID,
) {
	trades = r.trades
	tradesBySwapAcceptID = r.tradesBySwapAcceptID
	tradesByTrader = r.tradesByTrader
	tradesByMarket = r.tradesByMarket
	if tx, ok := ctx.Value(r).(*InMemoryTradeRepositoryTx); ok {
		trades = tx.trades
		tradesBySwapAcceptID = tx.tradesBySwapAcceptID
		tradesByMarket = tx.tradesByMarket
		tradesByTrader = tx.tradesByTrader
	}
	return
}

func getOrCreateTrade(trades map[uuid.UUID]*trade.Trade, tradeID *uuid.UUID) (*trade.Trade, error) {
	if tradeID == nil {
		t := trade.NewTrade()
		trades[t.ID()] = t
		return t, nil
	}

	currentTrade, ok := trades[*tradeID]
	if !ok {
		t := trade.NewTrade()
		trades[t.ID()] = t
		return t, nil
	}

	return currentTrade, nil
}

func getAllTrades(trades map[uuid.UUID]*trade.Trade) ([]*trade.Trade, error) {
	allTrades := make([]*trade.Trade, 0, len(trades))
	for _, trade := range trades {
		allTrades = append(allTrades, trade)
	}
	return allTrades, nil
}

func getAllTradesByMarket(
	trades map[uuid.UUID]*trade.Trade,
	tradesByMarket map[string][]uuid.UUID,
	marketQuoteAsset string,
) ([]*trade.Trade, error) {
	tradeIDs, ok := tradesByMarket[marketQuoteAsset]
	if !ok {
		return nil, fmt.Errorf(
			"no trades found for market with quote asset '%s'", marketQuoteAsset,
		)
	}

	tradeList := tradesFromIDs(trades, tradeIDs)
	return tradeList, nil
}

func getAllTradesByTrader(
	trades map[uuid.UUID]*trade.Trade,
	tradesByTrader map[string][]uuid.UUID,
	traderID string,
) ([]*trade.Trade, error) {
	tradeIDs, ok := tradesByTrader[traderID]
	if !ok {
		return nil, fmt.Errorf(
			"no trades found for trader with pubkey '%s'", traderID,
		)
	}

	tradeList := tradesFromIDs(trades, tradeIDs)
	return tradeList, nil
}

func tradesFromIDs(trades map[uuid.UUID]*trade.Trade, tradeIDs []uuid.UUID) []*trade.Trade {
	tradesByID := make([]*trade.Trade, 0, len(tradeIDs))
	for _, tradeID := range tradeIDs {
		tradesByID = append(tradesByID, trades[tradeID])
	}
	return tradesByID
}

func addTradeIDByKeyString(tradeMap map[string][]uuid.UUID, key string, val uuid.UUID) {
	trades, ok := tradeMap[key]
	if !ok {
		tradeMap[key] = []uuid.UUID{val}
		return
	}

	if !contain(trades, val) {
		tradeMap[key] = append(
			tradeMap[key],
			val,
		)
	}
}

func contain(list []uuid.UUID, id uuid.UUID) bool {
	for _, l := range list {
		if id == l {
			return true
		}
	}
	return false
}

// InMemoryTradeRepositoryTx allows to make transactional read/write operation
// on the in-memory repository
type InMemoryTradeRepositoryTx struct {
	root                 InMemoryTradeRepository
	trades               map[uuid.UUID]*trade.Trade
	tradesBySwapAcceptID map[string]uuid.UUID
	tradesByMarket       map[string][]uuid.UUID
	tradesByTrader       map[string][]uuid.UUID
}

// Commit applies the updates made to the state of the transaction to its root
func (tx *InMemoryTradeRepositoryTx) Commit() error {
	// update trades
	// TODO: check if trade has changed and update overwrite only updated trades
	for k, v := range tx.trades {
		tx.root.trades[k] = v
	}

	// update tradesBySwapAcceptID by adding only new entries, if any
	for k, v := range tx.tradesBySwapAcceptID {
		if _, ok := tx.root.tradesBySwapAcceptID[k]; !ok {
			tx.root.tradesBySwapAcceptID[k] = v
		}
	}

	// update tradesByTrader by either adding new entries or updating the list
	// of tradeIDs for each trader, if any has been added
	for key, txTradeIds := range tx.tradesByTrader {
		if rootTradeIDs, ok := tx.root.tradesByTrader[key]; !ok {
			tx.root.tradesByTrader[key] = txTradeIds
		} else {
			if len(txTradeIds) > len(rootTradeIDs) {
				for _, tradeID := range txTradeIds[len(rootTradeIDs):] {
					tx.root.tradesByTrader[key] = append(tx.root.tradesByTrader[key], tradeID)
				}
			}
		}
	}

	// update tradesByMarket by either adding new entries or updating the list
	// of tradeIDs for each market, if any has been added
	for key, txTradeIDs := range tx.tradesByMarket {
		if rootTradeIDs, ok := tx.root.tradesByMarket[key]; !ok {
			tx.root.tradesByMarket[key] = txTradeIDs
		} else {
			if len(txTradeIDs) > len(rootTradeIDs) {
				for _, tradeID := range txTradeIDs[len(rootTradeIDs):] {
					tx.root.tradesByMarket[key] = append(tx.root.tradesByTrader[key], tradeID)
				}
			}
		}
	}

	return nil
}

// Rollback resets the state of the transaction to the state of its root
func (tx *InMemoryTradeRepositoryTx) Rollback() error {
	tx.trades = map[uuid.UUID]*trade.Trade{}
	for k, v := range tx.root.trades {
		tx.trades[k] = v
	}

	tx.tradesBySwapAcceptID = map[string]uuid.UUID{}
	for k, v := range tx.root.tradesBySwapAcceptID {
		tx.tradesBySwapAcceptID[k] = v
	}

	tx.tradesByTrader = map[string][]uuid.UUID{}
	for k, v := range tx.root.tradesByTrader {
		cv := make([]uuid.UUID, len(v))
		copy(cv, v)
		tx.tradesByTrader[k] = cv
	}

	tx.tradesByMarket = map[string][]uuid.UUID{}
	for k, v := range tx.root.tradesByMarket {
		cv := make([]uuid.UUID, len(v))
		copy(cv, v)
		tx.tradesByMarket[k] = cv
	}
	return nil
}
