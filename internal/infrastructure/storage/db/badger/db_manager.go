package dbbadger

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/timshannon/badgerhold/v4"
)

// repoManager holds all the badgerhold stores in a single data structure.
type repoManager struct {
	marketStore *badgerhold.Store
	priceStore  *badgerhold.Store
	tradeStore  *badgerhold.Store
	txStore     *badgerhold.Store

	marketRepository     domain.MarketRepository
	tradeRepository      domain.TradeRepository
	depositRepository    domain.DepositRepository
	withdrawalRepository domain.WithdrawalRepository
}

// NewRepoManager opens (or creates if not exists) the badger store on disk.
// It expects a base data dir and an optional logger.
// It creates a dedicated directory for main and prices stores, while the
// unspent repository lives in memory.
func NewRepoManager(baseDbDir string, logger badger.Logger) (ports.RepoManager, error) {
	var marketsDir, pricesDir, tradesDir, txsDir string
	if len(baseDbDir) > 0 {
		marketsDir = filepath.Join(baseDbDir, "markets")
		pricesDir = filepath.Join(baseDbDir, "prices")
		tradesDir = filepath.Join(baseDbDir, "trades")
		txsDir = filepath.Join(baseDbDir, "transactions")
	}

	marketDb, err := createDb(marketsDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening main db: %w", err)
	}
	priceDb, err := createDb(pricesDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening prices db: %w", err)
	}
	tradeDb, err := createDb(tradesDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening unspents db: %w", err)
	}
	txDb, err := createDb(txsDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening unspents db: %w", err)
	}

	marketRepo := NewMarketRepositoryImpl(marketDb, priceDb)
	tradeRepo := NewTradeRepositoryImpl(tradeDb)
	depositRepository := NewDepositRepositoryImpl(txDb)
	withdrawalRepository := NewWithdrawalRepositoryImpl(txDb)

	return &repoManager{
		marketStore:          marketDb,
		priceStore:           priceDb,
		tradeStore:           tradeDb,
		txStore:              txDb,
		marketRepository:     marketRepo,
		tradeRepository:      tradeRepo,
		depositRepository:    depositRepository,
		withdrawalRepository: withdrawalRepository,
	}, nil
}

func (d *repoManager) MarketRepository() domain.MarketRepository {
	return d.marketRepository
}

func (d *repoManager) TradeRepository() domain.TradeRepository {
	return d.tradeRepository
}

func (d *repoManager) DepositRepository() domain.DepositRepository {
	return d.depositRepository
}

func (d *repoManager) WithdrawalRepository() domain.WithdrawalRepository {
	return d.withdrawalRepository
}

func (d *repoManager) Close() {
	d.marketStore.Close()
	d.priceStore.Close()
	d.tradeStore.Close()
	d.txStore.Close()
}

// isTransactionConflict returns whether the error occurred when committing a
// transacton is a conflict
func isTransactionConflict(err error) bool {
	return err == badger.ErrConflict
}

// EncodeKey encodes key values with a type prefix which allows multiple
// different types to exist in the badger DB
func EncodeKey(key interface{}, typeName string) ([]byte, error) {
	encoded, err := badgerhold.DefaultEncode(key)
	if err != nil {
		return nil, err
	}

	return append([]byte(typeName), encoded...), nil
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
				if err := db.Badger().RunValueLogGC(0.5); err != nil && err != badger.ErrNoRewrite {
					log.Error(err)
				}
			}
		}()
	}

	return db, nil
}
