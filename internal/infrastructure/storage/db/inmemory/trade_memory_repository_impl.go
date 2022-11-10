package inmemory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

type tradeInmemoryStore struct {
	trades               map[string]domain.Trade
	tradesBySwapAcceptId map[string]string
	tradesByMarket       map[string][]string
	locker               *sync.Mutex
}

type tradeRepositoryImpl struct {
	store *tradeInmemoryStore
}

// NewTradeRepositoryImpl returns a new inmemory TradeRepository implementation.
func NewTradeRepositoryImpl() domain.TradeRepository {
	return &tradeRepositoryImpl{&tradeInmemoryStore{
		trades:               map[string]domain.Trade{},
		tradesBySwapAcceptId: map[string]string{},
		tradesByMarket:       map[string][]string{},
		locker:               &sync.Mutex{},
	}}
}

func (r *tradeRepositoryImpl) AddTrade(
	_ context.Context, trade *domain.Trade,
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.addTrade(trade)
}

func (r *tradeRepositoryImpl) GetTradeById(
	_ context.Context, tradeId string,
) (*domain.Trade, error) {
	return r.getTrade(tradeId)
}

func (r *tradeRepositoryImpl) GetAllTrades(
	_ context.Context, page domain.Page,
) ([]domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getAllTrades(page)
}

func (r *tradeRepositoryImpl) GetAllTradesByMarket(
	_ context.Context, marketName string, page domain.Page,
) ([]domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getAllTradesByMarket(marketName, page)
}

func (r *tradeRepositoryImpl) GetCompletedTradesByMarket(
	_ context.Context, marketName string, page domain.Page,
) ([]domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	tradesByMarkets, err := r.getAllTradesByMarket(marketName, page)
	if err != nil {
		return nil, err
	}

	completedTrades := make([]domain.Trade, 0)
	for _, trade := range tradesByMarkets {
		if trade.Status.Code >= domain.TradeStatusCodeCompleted && !trade.Status.Failed {
			completedTrades = append(completedTrades, trade)
		}
	}

	return completedTrades, nil
}

func (r *tradeRepositoryImpl) GetTradeByTxId(
	ctx context.Context, txid string,
) (*domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	trades, err := r.getAllTrades(nil)
	if err != nil {
		return nil, err
	}
	for _, t := range trades {
		if t.TxId == txid {
			return &t, nil
		}
	}

	return nil, nil
}

func (r *tradeRepositoryImpl) GetTradeBySwapAcceptId(
	_ context.Context, swapAcceptId string,
) (*domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getTradeBySwapAcceptId(swapAcceptId)
}

func (r *tradeRepositoryImpl) UpdateTrade(
	ctx context.Context,
	tradeId string, updateFn func(t *domain.Trade) (*domain.Trade, error),
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	currentTrade, err := r.getTrade(tradeId)
	if err != nil {
		return err
	}

	updatedTrade, err := updateFn(currentTrade)
	if err != nil {
		return err
	}

	if swapAccept := updatedTrade.SwapAccept; swapAccept != nil {
		if _, ok := r.store.tradesBySwapAcceptId[swapAccept.Id]; !ok {
			r.store.tradesBySwapAcceptId[swapAccept.Id] = currentTrade.Id
		}
	}

	r.addTradeByMarket(updatedTrade.MarketName, currentTrade.Id)

	r.store.trades[updatedTrade.Id] = *updatedTrade

	return nil
}

func (r *tradeRepositoryImpl) addTrade(trade *domain.Trade) error {
	if _, ok := r.store.trades[trade.Id]; ok {
		return fmt.Errorf("trade with id %s already exists", trade.Id)
	}
	r.store.trades[trade.Id] = *trade
	r.store.tradesByMarket[trade.MarketName] = append(r.store.tradesByMarket[trade.MarketName], trade.Id)
	return nil
}

func (r *tradeRepositoryImpl) getTrade(id string) (*domain.Trade, error) {
	trade, ok := r.store.trades[id]
	if !ok {
		return nil, fmt.Errorf("trade with id %s not found", id)
	}
	return &trade, nil
}

func (r *tradeRepositoryImpl) getAllTrades(
	page domain.Page,
) ([]domain.Trade, error) {
	trades := make([]domain.Trade, 0, len(r.store.trades))
	for _, trade := range r.store.trades {
		trades = append(trades, trade)
	}
	sort.SliceStable(trades, func(i, j int) bool {
		return trades[i].SwapRequest.Timestamp > trades[j].SwapRequest.Timestamp
	})

	if page == nil {
		return trades, nil
	}

	pagTrades := make([]domain.Trade, 0)
	startIndex := int(page.GetNumber()*page.GetSize() - page.GetSize())
	endIndex := int(page.GetNumber() * page.GetSize())
	for i, trade := range trades {
		if i >= startIndex && i < endIndex {
			pagTrades = append(pagTrades, trade)
		}
	}
	return pagTrades, nil
}

func (r *tradeRepositoryImpl) getAllTradesByMarket(
	marketName string, page domain.Page,
) ([]domain.Trade, error) {
	tradeIds, ok := r.store.tradesByMarket[marketName]
	if !ok {
		return nil, nil
	}

	tradeList := tradesFromIds(r.store.trades, tradeIds, page)
	return tradeList, nil
}

func (r *tradeRepositoryImpl) getTradeBySwapAcceptId(
	swapAcceptId string,
) (*domain.Trade, error) {
	tradeId, ok := r.store.tradesBySwapAcceptId[swapAcceptId]
	if !ok {
		return nil, fmt.Errorf(
			"trade with swap accept id %s not found", swapAcceptId,
		)
	}
	trade := r.store.trades[tradeId]
	return &trade, nil
}

func (r *tradeRepositoryImpl) addTradeByMarket(key string, val string) {
	trades, ok := r.store.tradesByMarket[key]
	if !ok {
		r.store.tradesByMarket[key] = []string{val}
		return
	}

	if !contain(trades, val) {
		r.store.tradesByMarket[key] = append(r.store.tradesByMarket[key], val)
	}
}

func tradesFromIds(
	trades map[string]domain.Trade, tradeIds []string, page domain.Page,
) []domain.Trade {
	tradesById := make([]domain.Trade, 0, len(trades))
	for _, tradeId := range tradeIds {
		trade := trades[tradeId]
		tradesById = append(tradesById, trade)
	}
	sort.SliceStable(tradesById, func(i, j int) bool {
		return tradesById[i].SwapRequest.Timestamp > tradesById[j].SwapRequest.Timestamp
	})

	if page == nil {
		return tradesById
	}

	pagTrades := make([]domain.Trade, 0, page.GetSize())
	startIndex := int(page.GetNumber()*page.GetSize() - page.GetSize())
	endIndex := int(page.GetNumber() * page.GetSize())
	for i, trade := range tradesById {
		if i >= startIndex && i < endIndex {
			pagTrades = append(pagTrades, trade)
		}
	}
	return pagTrades

}

func contain(list []string, id string) bool {
	for _, l := range list {
		if id == l {
			return true
		}
	}
	return false
}
