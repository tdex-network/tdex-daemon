package pricefeederinfra

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v3/options"
	log "github.com/sirupsen/logrus"

	"github.com/dgraph-io/badger/v3"

	"github.com/timshannon/badgerhold/v4"
)

type PriceFeedStore interface {
	// AddPriceFeed adds a new price feed to the repository.
	AddPriceFeed(ctx context.Context, priceFeed PriceFeed) error
	// GetPriceFeed returns the price feed with the given ID.
	GetPriceFeed(ctx context.Context, id string) (*PriceFeed, error)
	// GetPriceFeedsByMarket returns all price feed for a given market.
	GetPriceFeedsByMarket(
		ctx context.Context,
		market Market,
	) (*PriceFeed, error)
	// UpdatePriceFeed updates the price feed.
	UpdatePriceFeed(
		ctx context.Context,
		ID string,
		updateFn func(priceFeed *PriceFeed) (*PriceFeed, error),
	) error
	// RemovePriceFeed removes the price feed with the given ID.
	RemovePriceFeed(ctx context.Context, id string) error
	// GetAllPriceFeeds returns all price feeds of all markets.
	GetAllPriceFeeds(ctx context.Context) ([]PriceFeed, error)
	// GetStartedPriceFeeds returns all price feeds that are started.
	GetStartedPriceFeeds(ctx context.Context) ([]PriceFeed, error)
}

type priceFeedRepositoryImpl struct {
	store *badgerhold.Store
}

func NewPriceFeedStoreImpl(
	baseDbDir string, logger badger.Logger,
) (PriceFeedStore, error) {
	var priceFeederDir string
	if len(baseDbDir) > 0 {
		priceFeederDir = filepath.Join(baseDbDir, "priceFeeder")
	}

	store, err := createDb(priceFeederDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening main db: %w", err)
	}
	return &priceFeedRepositoryImpl{
		store,
	}, nil
}

func (p *priceFeedRepositoryImpl) AddPriceFeed(
	ctx context.Context,
	priceFeed PriceFeed,
) error {
	pf, err := p.GetPriceFeedsByMarket(ctx, priceFeed.Market)
	if err != nil {
		return err
	}

	if pf != nil {
		return ErrPriceFeedAlreadyExists
	}

	if err := p.store.Insert(priceFeed.ID, &priceFeed); err != nil {
		if err == badgerhold.ErrKeyExists {
			return nil
		}

		return err
	}

	return nil
}

func (p *priceFeedRepositoryImpl) GetPriceFeed(
	ctx context.Context,
	id string,
) (*PriceFeed, error) {
	var priceFeed PriceFeed
	if err := p.store.Get(id, &priceFeed); err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, ErrPriceFeedNotFound
		}

		return nil, err
	}

	return &priceFeed, nil
}

func (p *priceFeedRepositoryImpl) GetPriceFeedsByMarket(
	ctx context.Context,
	market Market,
) (*PriceFeed, error) {
	query := badgerhold.Where("Market.BaseAsset").Eq(market.BaseAsset).
		And("Market.QuoteAsset").Eq(market.QuoteAsset)

	findPriceFeed, err := p.findPriceFeed(ctx, query)
	if err != nil {
		return nil, err
	}

	return findPriceFeed, nil
}

func (p *priceFeedRepositoryImpl) UpdatePriceFeed(
	ctx context.Context,
	ID string,
	updateFn func(priceFeed *PriceFeed) (*PriceFeed, error),
) error {
	priceFeed, err := p.GetPriceFeed(ctx, ID)
	if err != nil {
		return err
	}

	oldBaseAsset := priceFeed.Market.BaseAsset
	oldQuoteAsset := priceFeed.Market.QuoteAsset

	updatedPriceFeed, err := updateFn(priceFeed)
	if err != nil {
		return err
	}

	if oldBaseAsset != updatedPriceFeed.Market.BaseAsset ||
		oldQuoteAsset != updatedPriceFeed.Market.QuoteAsset {
		return ErrPriceFeedMarketCannotBeChanged
	}

	if err := p.store.Update(ID, updatedPriceFeed); err != nil {
		return err
	}

	return nil
}

func (p *priceFeedRepositoryImpl) findPriceFeed(
	ctx context.Context, query *badgerhold.Query,
) (*PriceFeed, error) {
	var priceFeed []PriceFeed
	if err := p.store.Find(&priceFeed, query); err != nil {
		return nil, err
	}

	if len(priceFeed) == 0 {
		return nil, nil
	}

	return &priceFeed[0], nil
}

func (p *priceFeedRepositoryImpl) RemovePriceFeed(
	ctx context.Context,
	id string,
) error {
	return p.store.Delete(id, PriceFeed{})
}

func (p *priceFeedRepositoryImpl) GetAllPriceFeeds(
	ctx context.Context,
) ([]PriceFeed, error) {
	var priceFeeds []PriceFeed
	if err := p.store.Find(&priceFeeds, nil); err != nil {
		return nil, err
	}

	return priceFeeds, nil
}

func (p *priceFeedRepositoryImpl) GetStartedPriceFeeds(
	ctx context.Context) ([]PriceFeed, error) {
	query := badgerhold.Where("Started").Eq(true)

	var priceFeeds []PriceFeed
	if err := p.store.Find(&priceFeeds, query); err != nil {
		return nil, err
	}

	return priceFeeds, nil
}

func createDb(dbDir string, logger badger.Logger) (*badgerhold.Store, error) {
	isInMemory := len(dbDir) <= 0

	opts := badger.DefaultOptions(dbDir)
	opts.Logger = logger

	if isInMemory {
		opts.InMemory = true
	} else {
		opts.Compression = options.ZSTD
	}

	db, err := badgerhold.Open(badgerhold.Options{
		Encoder:          badgerhold.DefaultEncode,
		Decoder:          badgerhold.DefaultDecode,
		SequenceBandwith: 100,
		Options:          opts,
	})
	if err != nil {
		return nil, err
	}

	if !isInMemory {
		ticker := time.NewTicker(30 * time.Minute)

		go func() {
			for {
				<-ticker.C
				if err := db.Badger().RunValueLogGC(0.5); err != nil &&
					err != badger.ErrNoRewrite {
					log.Error(err)
				}
			}
		}()
	}

	return db, nil
}
