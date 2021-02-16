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
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/timshannon/badgerhold/v2"
)

// DbManager holds all the badgerhold stores in a single data structure.
type DbManager struct {
	Store        *badgerhold.Store
	PriceStore   *badgerhold.Store
	UnspentStore *badgerhold.Store
}

// NewDbManager opens (or creates if not exists) the badger store on disk. It expects a base data dir and an optional logger.
// It creates a dedicated directory for main, price and unspent.
func NewDbManager(baseDbDir string, logger badger.Logger) (*DbManager, error) {
	mainDb, err := createDb(filepath.Join(baseDbDir, "main"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening main db: %w", err)
	}

	priceDb, err := createDb(filepath.Join(baseDbDir, "prices"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening prices db: %w", err)
	}

	unspentDb, err := createDb(filepath.Join(baseDbDir, "unspents"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening unspents db: %w", err)
	}

	return &DbManager{
		Store:        mainDb,
		PriceStore:   priceDb,
		UnspentStore: unspentDb,
	}, nil
}

// NewTransaction implements the DbManager interface
func (d DbManager) NewTransaction() ports.Transaction {
	return d.Store.Badger().NewTransaction(true)
}

// NewPricesTransaction implements the DbManager interface
func (d DbManager) NewPricesTransaction() ports.Transaction {
	return d.PriceStore.Badger().NewTransaction(true)
}

// NewUnspentsTransaction implements the DbManager interface
func (d DbManager) NewUnspentsTransaction() ports.Transaction {
	return d.UnspentStore.Badger().NewTransaction(true)
}

// RunTransaction invokes the given handler and retries in case the transaction
// returns a conflict error
func (d DbManager) RunTransaction(
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
func (d DbManager) RunUnspentsTransaction(
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
func (d DbManager) RunPricesTransaction(
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

func (d DbManager) runTransaction(
	args runTransactionArgs,
) (interface{}, error) {
	for {
		tx, ctx := args.ctxMaker()
		res, err := args.handler(ctx)
		if err != nil {
			if args.readOnly && d.isTransactionConflict(err) {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			return nil, err
		}

		if !args.readOnly {
			if err := tx.Commit(); err != nil {
				if !d.isTransactionConflict(err) {
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
func (d DbManager) isTransactionConflict(err error) bool {
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
	opts := badger.DefaultOptions(dbDir)
	opts.Logger = logger
	opts.ValueLogLoadingMode = options.FileIO
	opts.Compression = options.ZSTD

	db, err := badgerhold.Open(badgerhold.Options{
		Encoder:          badgerhold.DefaultEncode,
		Decoder:          badgerhold.DefaultDecode,
		SequenceBandwith: 100,
		Options:          opts,
	})
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(30 * time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := db.Badger().RunValueLogGC(0.5); err != nil && err != badger.ErrNoRewrite {
					log.Error(err)
				}
			}
		}
	}()

	return db, nil
}
