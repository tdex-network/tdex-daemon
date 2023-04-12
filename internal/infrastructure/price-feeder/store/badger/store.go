package pricefeederstore

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	log "github.com/sirupsen/logrus"
	pricefeeder "github.com/tdex-network/tdex-daemon/internal/infrastructure/price-feeder"
	"github.com/timshannon/badgerhold/v4"
)

type priceFeedStore struct {
	store *badgerhold.Store
}

func NewPriceFeedStore(
	baseDbDir string, logger badger.Logger,
) (pricefeeder.PriceFeedStore, error) {
	var priceFeederDir string
	if len(baseDbDir) > 0 {
		priceFeederDir = filepath.Join(baseDbDir, "feeder")
	}

	store, err := createDb(priceFeederDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening main db: %w", err)
	}
	return &priceFeedStore{store}, nil
}

func (p *priceFeedStore) AddPriceFeed(
	ctx context.Context, priceFeed pricefeeder.PriceFeedInfo,
) error {
	query := badgerhold.Where("Market.BaseAsset").Eq(priceFeed.Market.BaseAsset).
		And("Market.QuoteAsset").Eq(priceFeed.Market.QuoteAsset)

	pf, err := p.findPriceFeed(ctx, query)
	if err != nil {
		return err
	}
	if pf != nil {
		return fmt.Errorf("price feed already exists")
	}

	if err := p.store.Insert(priceFeed.ID, &priceFeed); err != nil {
		if err == badgerhold.ErrKeyExists {
			return fmt.Errorf("price feed already exists")
		}
		return err
	}

	return nil
}

func (p *priceFeedStore) GetPriceFeed(
	ctx context.Context, id string,
) (*pricefeeder.PriceFeedInfo, error) {
	var priceFeed pricefeeder.PriceFeedInfo
	if err := p.store.Get(id, &priceFeed); err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, fmt.Errorf("price feed not found")
		}
		return nil, err
	}

	return &priceFeed, nil
}

func (p *priceFeedStore) UpdatePriceFeed(
	ctx context.Context, id string, updateFn func(
		*pricefeeder.PriceFeedInfo,
	) (*pricefeeder.PriceFeedInfo, error),
) error {
	priceFeed, err := p.GetPriceFeed(ctx, id)
	if err != nil {
		return err
	}

	updatedPriceFeed, err := updateFn(priceFeed)
	if err != nil {
		return err
	}

	return p.store.Update(id, updatedPriceFeed)
}

func (p *priceFeedStore) RemovePriceFeed(
	ctx context.Context, id string,
) error {
	if err := p.store.Delete(id, pricefeeder.PriceFeedInfo{}); err != nil {
		if err == badgerhold.ErrNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (p *priceFeedStore) GetAllPriceFeeds(
	ctx context.Context,
) ([]pricefeeder.PriceFeedInfo, error) {
	var priceFeeds []pricefeeder.PriceFeedInfo
	if err := p.store.Find(&priceFeeds, nil); err != nil {
		return nil, err
	}

	return priceFeeds, nil
}

func (p *priceFeedStore) Close() {
	p.store.Close()
}

func (p *priceFeedStore) findPriceFeed(
	ctx context.Context, query *badgerhold.Query,
) (*pricefeeder.PriceFeedInfo, error) {
	var priceFeed []pricefeeder.PriceFeedInfo
	if err := p.store.Find(&priceFeed, query); err != nil {
		return nil, err
	}

	if len(priceFeed) == 0 {
		return nil, nil
	}

	return &priceFeed[0], nil
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
