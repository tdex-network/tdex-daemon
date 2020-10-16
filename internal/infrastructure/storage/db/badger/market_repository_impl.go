package dbbadger

import (
	"context"
	"github.com/dgraph-io/badger"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold"
)

type marketRepositoryImpl struct {
	db *DbManager
}

func NewMarketRepositoryImpl(db *DbManager) domain.MarketRepository {
	return marketRepositoryImpl{
		db: db,
	}
}

func (m marketRepositoryImpl) GetMarketByAsset(
	ctx context.Context,
	quoteAsset string,
) (market *domain.Market, accountIndex int, err error) {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("QuoteAsset").Eq(quoteAsset)
	markets, err := m.findMarkets(tx, query)
	if err != nil {
		return
	}

	if len(markets) > 0 {
		market = &markets[0]
		accountIndex = market.AccountIndex
	}

	return
}

func (m marketRepositoryImpl) GetLatestMarket(
	ctx context.Context,
) (market *domain.Market, accountIndex int, err error) {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("AccountIndex").
		Ge(domain.MarketAccountStart).
		SortBy("AccountIndex").
		Reverse()
	markets, err := m.findMarkets(tx, query)
	if err != nil {
		return
	}

	if len(markets) > 0 {
		market = &markets[0]
		accountIndex = market.AccountIndex
	}

	return
}

func (m marketRepositoryImpl) GetOrCreateMarket(
	ctx context.Context,
	accountIndex int,
) (*domain.Market, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	return m.getOrCreateMarket(tx, accountIndex)
}

func (m marketRepositoryImpl) GetTradableMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("AccountIndex").
		Ge(domain.MarketAccountStart).
		And("Tradable").
		Eq(true)

	return m.findMarkets(tx, query)
}

func (m marketRepositoryImpl) GetAllMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("AccountIndex").Ge(domain.MarketAccountStart)

	return m.findMarkets(tx, query)
}

func (m marketRepositoryImpl) UpdateMarket(
	ctx context.Context,
	accountIndex int,
	updateFn func(m *domain.Market) (*domain.Market, error),
) error {
	tx := ctx.Value("tx").(*badger.Txn)

	currentMarket, err := m.getOrCreateMarket(tx, accountIndex)
	if err != nil {
		return err
	}

	updatedMarket, err := updateFn(currentMarket)
	if err != nil {
		return err
	}

	return m.updateMarket(tx, accountIndex, *updatedMarket)
}

func (m marketRepositoryImpl) OpenMarket(
	ctx context.Context,
	quoteAsset string,
) error {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("QuoteAsset").Eq(quoteAsset)
	markets, err := m.findMarkets(tx, query)
	if err != nil {
		return err
	}

	if len(markets) > 0 {
		market := markets[0]
		if market.IsTradable() {
			return nil
		}

		err = market.MakeTradable()
		if err != nil {
			return err
		}

		err := m.updateMarket(tx, market.AccountIndex, market)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m marketRepositoryImpl) CloseMarket(
	ctx context.Context,
	quoteAsset string,
) error {
	tx := ctx.Value("tx").(*badger.Txn)

	query := badgerhold.Where("QuoteAsset").Eq(quoteAsset)
	markets, err := m.findMarkets(tx, query)
	if err != nil {
		return err
	}

	if len(markets) > 0 {
		market := markets[0]
		if !market.IsTradable() {
			return nil
		}

		err = market.MakeNotTradable()
		if err != nil {
			return err
		}

		err := m.updateMarket(tx, market.AccountIndex, market)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m marketRepositoryImpl) getOrCreateMarket(
	tx *badger.Txn,
	accountIndex int,
) (*domain.Market, error) {
	market, err := m.getMarket(tx, accountIndex)
	if err != nil {
		return nil, err
	}

	if market == nil {
		market, err = domain.NewMarket(accountIndex)
		if err != nil {
			return nil, err
		}

		err = m.insertMarket(tx, accountIndex, *market)
		if err != nil {
			return nil, err
		}
	}

	return market, nil
}

func (m marketRepositoryImpl) insertMarket(
	tx *badger.Txn,
	accountIndex int,
	market domain.Market,
) error {
	var err error
	if tx != nil {
		err = m.db.Store.TxInsert(
			tx,
			accountIndex,
			&market,
		)
	} else {
		err = m.db.Store.Insert(
			vaultKey,
			&market,
		)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}

	return nil
}

func (m marketRepositoryImpl) getMarket(
	tx *badger.Txn,
	accountIndex int,
) (*domain.Market, error) {
	var err error
	var market domain.Market
	if tx != nil {
		err = m.db.Store.TxGet(
			tx,
			accountIndex,
			&market,
		)
	} else {
		err = m.db.Store.Get(
			accountIndex,
			&market,
		)
	}

	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &market, nil
}

func (m marketRepositoryImpl) updateMarket(
	tx *badger.Txn,
	accountIndex int,
	market domain.Market,
) error {
	var err error
	if tx != nil {
		err = m.db.Store.TxUpdate(
			tx,
			accountIndex,
			market,
		)
	} else {
		err = m.db.Store.Update(
			accountIndex,
			market,
		)
	}
	return err
}

func (m marketRepositoryImpl) findMarkets(
	tx *badger.Txn,
	query *badgerhold.Query,
) ([]domain.Market, error) {
	var markets []domain.Market
	var err error
	if tx != nil {
		err = m.db.Store.TxFind(
			tx,
			&markets,
			query,
		)
	} else {
		err = m.db.Store.Find(
			&markets,
			query,
		)
	}

	return markets, err
}
