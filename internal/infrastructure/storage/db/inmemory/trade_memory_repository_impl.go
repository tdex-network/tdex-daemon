package inmemory

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/uow"
	"sync"

	"github.com/google/uuid"
)

// TradeRepositoryImpl represents an in memory storage
type TradeRepositoryImpl struct {
	trades               map[uuid.UUID]*domain.Trade
	tradesBySwapAcceptID map[string]uuid.UUID
	tradesByTrader       map[string][]uuid.UUID
	tradesByMarket       map[string][]uuid.UUID

	lock *sync.RWMutex
}

// NewTradeRepositoryImpl returns a new empty TradeRepositoryImpl
func NewTradeRepositoryImpl() *TradeRepositoryImpl {
	return &TradeRepositoryImpl{
		trades:               map[uuid.UUID]*domain.Trade{},
		tradesBySwapAcceptID: map[string]uuid.UUID{},
		tradesByTrader:       map[string][]uuid.UUID{},
		tradesByMarket:       map[string][]uuid.UUID{},
		lock:                 &sync.RWMutex{},
	}
}

// GetOrCreateTrade gets a trade with a given swapID that can be either a
// request, accept, or complete ID. They all identify the same Trade.
// If not found, a new entry is inserted
func (r TradeRepositoryImpl) GetOrCreateTrade(ctx context.Context, tradeID *uuid.UUID) (*domain.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	trades, _, _, _ := r.storageByContext(ctx)
	return getOrCreateTrade(trades, tradeID)
}

// GetAllTrades returns all the trades processed by the daemon
func (r TradeRepositoryImpl) GetAllTrades(ctx context.Context) ([]*domain.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	trades, _, _, _ := r.storageByContext(ctx)
	return getAllTrades(trades)
}

// GetAllTradesByMarket returns all the trades processed for the given market
func (r TradeRepositoryImpl) GetAllTradesByMarket(ctx context.Context, marketQuoteAsset string) ([]*domain.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	trades, _, tradesByMarket, _ := r.storageByContext(ctx)
	return getAllTradesByMarket(trades, tradesByMarket, marketQuoteAsset)
}

// GetAllTradesByTrader returns all the trades processed for the given trader
func (r TradeRepositoryImpl) GetAllTradesByTrader(ctx context.Context, traderID string) ([]*domain.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	trades, _, _, tradesByTrader := r.storageByContext(ctx)
	return getAllTradesByTrader(trades, tradesByTrader, traderID)
}

// GetTradeBySwapAcceptID returns the trade idetified by the swap accept message's id
func (r TradeRepositoryImpl) GetTradeBySwapAcceptID(ctx context.Context, swapAcceptID string) (*domain.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	trades, tradesBySwapAcceptID, _, _ := r.storageByContext(ctx)
	return getTradeBySwapAcceptID(trades, tradesBySwapAcceptID, swapAcceptID)
}

// UpdateTrade updates data to a trade identified by any of its swap ids (request, accept, complete) passing an update function
func (r TradeRepositoryImpl) UpdateTrade(
	ctx context.Context,
	tradeID *uuid.UUID,
	updateFn func(t *domain.Trade) (*domain.Trade, error),
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
			tradesBySwapAcceptID[swapAccept.GetId()] = currentTrade.ID
		}
	}

	addTradeIDByKeyString(tradesByMarket, updatedTrade.MarketQuoteAsset, currentTrade.ID)
	addTradeIDByKeyString(tradesByTrader, hex.EncodeToString(updatedTrade.TraderPubkey), currentTrade.ID)
	return nil
}

// Begin returns a new TradeRepositoryTx
func (r TradeRepositoryImpl) Begin() (uow.Tx, error) {
	tx := &TradeRepositoryTx{
		root:                 r,
		trades:               map[uuid.UUID]*domain.Trade{},
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
func (r TradeRepositoryImpl) ContextKey() interface{} {
	return uow.InMemoryContextKey
}

func (r TradeRepositoryImpl) storageByContext(ctx context.Context) (
	trades map[uuid.UUID]*domain.Trade,
	tradesBySwapAcceptID map[string]uuid.UUID,
	tradesByMarket map[string][]uuid.UUID,
	tradesByTrader map[string][]uuid.UUID,
) {
	trades = r.trades
	tradesBySwapAcceptID = r.tradesBySwapAcceptID
	tradesByTrader = r.tradesByTrader
	tradesByMarket = r.tradesByMarket
	if tx, ok := ctx.Value(r).(*TradeRepositoryTx); ok {
		trades = tx.trades
		tradesBySwapAcceptID = tx.tradesBySwapAcceptID
		tradesByMarket = tx.tradesByMarket
		tradesByTrader = tx.tradesByTrader
	}
	return
}

func getOrCreateTrade(trades map[uuid.UUID]*domain.Trade, tradeID *uuid.UUID) (*domain.Trade, error) {
	if tradeID == nil {
		t := domain.NewTrade()
		trades[t.ID] = t
		return t, nil
	}

	currentTrade, ok := trades[*tradeID]
	if !ok {
		t := domain.NewTrade()
		trades[t.ID] = t
		return t, nil
	}

	return currentTrade, nil
}

func getAllTrades(trades map[uuid.UUID]*domain.Trade) ([]*domain.Trade, error) {
	allTrades := make([]*domain.Trade, 0, len(trades))
	for _, trade := range trades {
		allTrades = append(allTrades, trade)
	}
	return allTrades, nil
}

func getAllTradesByMarket(
	trades map[uuid.UUID]*domain.Trade,
	tradesByMarket map[string][]uuid.UUID,
	marketQuoteAsset string,
) ([]*domain.Trade, error) {
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
	trades map[uuid.UUID]*domain.Trade,
	tradesByTrader map[string][]uuid.UUID,
	traderID string,
) ([]*domain.Trade, error) {
	tradeIDs, ok := tradesByTrader[traderID]
	if !ok {
		return nil, fmt.Errorf(
			"no trades found for trader with pubkey '%s'", traderID,
		)
	}

	tradeList := tradesFromIDs(trades, tradeIDs)
	return tradeList, nil
}

func getTradeBySwapAcceptID(
	trades map[uuid.UUID]*domain.Trade,
	tradesBySwapAcceptID map[string]uuid.UUID,
	swapAcceptID string,
) (*domain.Trade, error) {
	tradeID, ok := tradesBySwapAcceptID[swapAcceptID]
	if !ok {
		return nil, fmt.Errorf(
			"trade not found for swap accept message with id '%s'", swapAcceptID,
		)
	}
	return trades[tradeID], nil
}

func tradesFromIDs(trades map[uuid.UUID]*domain.Trade, tradeIDs []uuid.UUID) []*domain.Trade {
	tradesByID := make([]*domain.Trade, 0, len(tradeIDs))
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

// TradeRepositoryTx allows to make transactional read/write operation
// on the in-memory repository
type TradeRepositoryTx struct {
	root                 TradeRepositoryImpl
	trades               map[uuid.UUID]*domain.Trade
	tradesBySwapAcceptID map[string]uuid.UUID
	tradesByMarket       map[string][]uuid.UUID
	tradesByTrader       map[string][]uuid.UUID
}

// Commit applies the updates made to the state of the transaction to its root
func (tx *TradeRepositoryTx) Commit() error {
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
func (tx *TradeRepositoryTx) Rollback() error {
	tx.trades = map[uuid.UUID]*domain.Trade{}
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
