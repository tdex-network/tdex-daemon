package dbbadger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/timshannon/badgerhold/v2"
)

// repoManager holds all the badgerhold stores in a single data structure.
type repoManager struct {
	store        *badgerhold.Store
	priceStore   *badgerhold.Store
	unspentStore *badgerhold.Store

	marketRepository     domain.MarketRepository
	unspentRepository    domain.UnspentRepository
	tradeRepository      domain.TradeRepository
	vaultRepository      domain.VaultRepository
	depositRepository    domain.DepositRepository
	withdrawalRepository domain.WithdrawalRepository
}

// NewRepoManager opens (or creates if not exists) the badger store on disk.
// It expects a base data dir and an optional logger.
// It creates a dedicated directory for main and prices stores, while the
// unspent repository lives in memory.
func NewRepoManager(baseDbDir string, logger badger.Logger) (ports.RepoManager, error) {
	var maindbDir, pricedbDir, unspentDir string
	if len(baseDbDir) > 0 {
		maindbDir = filepath.Join(baseDbDir, "main")
		pricedbDir = filepath.Join(baseDbDir, "prices")
		unspentDir = filepath.Join(baseDbDir, "unspents")
	}

	mainDb, err := createDb(maindbDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening main db: %w", err)
	}

	priceDb, err := createDb(pricedbDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening prices db: %w", err)
	}

	unspentDb, err := createDb(unspentDir, logger)
	if err != nil {
		return nil, fmt.Errorf("opening unspents db: %w", err)
	}

	marketRepo := NewMarketRepositoryImpl(mainDb, priceDb)
	unspentRepo := NewUnspentRepositoryImpl(unspentDb, mainDb)
	tradeRepo := NewTradeRepositoryImpl(mainDb)
	vaultRepo := NewVaultRepositoryImpl(mainDb)
	depositRepository := NewDepositRepositoryImpl(mainDb)
	withdrawalRepository := NewWithdrawalRepositoryImpl(mainDb)

	return &repoManager{
		store:                mainDb,
		priceStore:           priceDb,
		unspentStore:         unspentDb,
		marketRepository:     marketRepo,
		unspentRepository:    unspentRepo,
		tradeRepository:      tradeRepo,
		vaultRepository:      vaultRepo,
		depositRepository:    depositRepository,
		withdrawalRepository: withdrawalRepository,
	}, nil
}

func (d *repoManager) MarketRepository() domain.MarketRepository {
	return d.marketRepository
}

func (d *repoManager) UnspentRepository() domain.UnspentRepository {
	return d.unspentRepository
}

func (d *repoManager) TradeRepository() domain.TradeRepository {
	return d.tradeRepository
}

func (d *repoManager) VaultRepository() domain.VaultRepository {
	return d.vaultRepository
}

func (d *repoManager) DepositRepository() domain.DepositRepository {
	return d.depositRepository
}

func (d *repoManager) WithdrawalRepository() domain.WithdrawalRepository {
	return d.withdrawalRepository
}

func (d *repoManager) Close() {
	d.store.Close()
	d.priceStore.Close()
	d.unspentStore.Close()
}

// NewTransaction implements the RepoManager interface
func (d *repoManager) NewTransaction() ports.Transaction {
	return d.store.Badger().NewTransaction(true)
}

// NewPricesTransaction implements the RepoManager interface
func (d *repoManager) NewPricesTransaction() ports.Transaction {
	return d.priceStore.Badger().NewTransaction(true)
}

// NewUnspentsTransaction implements the RepoManager interface
func (d *repoManager) NewUnspentsTransaction() ports.Transaction {
	return d.unspentStore.Badger().NewTransaction(true)
}

// RunTransaction invokes the given handler and retries in case the transaction
// returns a conflict error
func (d *repoManager) RunTransaction(
	ctx context.Context,
	readOnly bool,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	ctxMaker := func() (ports.Transaction, context.Context) {
		tx := d.NewTransaction()
		_ctx := context.WithValue(ctx, "tx", tx)
		return tx, _ctx
	}
	return d.runTransaction(runTransactionArgs{
		ctxMaker: ctxMaker,
		readOnly: readOnly,
		handler:  handler,
	})
}

// RunUnspentsTransaction invokes the given handler and retries in case the
// unspents transaction returns a conflict error
func (d *repoManager) RunUnspentsTransaction(
	ctx context.Context,
	readOnly bool,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	ctxMaker := func() (ports.Transaction, context.Context) {
		tx := d.NewUnspentsTransaction()
		_ctx := context.WithValue(ctx, "utx", tx)
		return tx, _ctx
	}

	return d.runTransaction(runTransactionArgs{
		ctxMaker: ctxMaker,
		readOnly: readOnly,
		handler:  handler,
	})
}

// RunPricesTransaction invokes the given handler and retries in case the
// unspents transaction returns a conflict error
func (d *repoManager) RunPricesTransaction(
	ctx context.Context,
	readOnly bool,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	ctxMaker := func() (ports.Transaction, context.Context) {
		tx := d.NewPricesTransaction()
		_ctx := context.WithValue(ctx, "ptx", tx)
		return tx, _ctx
	}

	return d.runTransaction(runTransactionArgs{
		ctxMaker: ctxMaker,
		readOnly: readOnly,
		handler:  handler,
	})
}

type runTransactionArgs struct {
	ctxMaker func() (ports.Transaction, context.Context)
	readOnly bool
	handler  func(ctx context.Context) (interface{}, error)
}

func (d *repoManager) runTransaction(
	args runTransactionArgs,
) (interface{}, error) {
	for {
		tx, ctx := args.ctxMaker()
		res, err := args.handler(ctx)
		if err != nil {
			if args.readOnly && isTransactionConflict(err) {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			return nil, err
		}

		if !args.readOnly {
			if err := tx.Commit(); err != nil {
				if !isTransactionConflict(err) {
					return nil, err
				}
				time.Sleep(50 * time.Millisecond)
				continue
			}
		}
		return res, nil
	}
}

// isTransactionConflict returns wheter the error occured when commiting a
// transacton is a conflict
func isTransactionConflict(err error) bool {
	return err == badger.ErrConflict
}

// JSONEncode is a custom JSON based encoder for badger
func JSONEncode(value interface{}) ([]byte, error) {
	var buff bytes.Buffer

	en := json.NewEncoder(&buff)

	err := en.Encode(value)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// JSONDecode is a custom JSON based decoder for badger
func JSONDecode(data []byte, value interface{}) error {
	var buff bytes.Buffer
	de := json.NewDecoder(&buff)

	_, err := buff.Write(data)
	if err != nil {
		return err
	}

	return de.Decode(value)
}

// EncodeKey encodes key values with a type prefix which allows multiple
//different types to exist in the badger DB
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
		opts.ValueLogLoadingMode = options.FileIO
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
