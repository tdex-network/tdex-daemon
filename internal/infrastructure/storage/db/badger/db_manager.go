package dbbadger

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v2"
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
// It creates a dedicated directory for swap, price and unspent.
func NewDbManager(baseDbDir string, logger badger.Logger) (*DbManager, error) {
	swapDb, err := createDb(baseDbDir+"/swap", logger)
	if err != nil {
		return nil, fmt.Errorf("opening swap db: %w", err)
	}

	priceDb, err := createDb(baseDbDir+"/price", logger)
	if err != nil {
		return nil, fmt.Errorf("opening price db: %w", err)
	}

	unspentDb, err := createDb(baseDbDir+"/unspent", logger)
	if err != nil {
		return nil, fmt.Errorf("opening unspent db: %w", err)
	}

	return &DbManager{
		Store:        swapDb,
		PriceStore:   priceDb,
		UnspentStore: unspentDb,
	}, nil
}

// NewTransaction implements the DbManager interface
func (d DbManager) NewTransaction() ports.Transaction {
	return d.Store.Badger().NewTransaction(true)
}

// NewPriceTransaction implements the DbManager interface
func (d DbManager) NewPriceTransaction() ports.Transaction {
	return d.PriceStore.Badger().NewTransaction(true)
}

// NewUnspentTransaction implements the DbManager interface
func (d DbManager) NewUnspentTransaction() ports.Transaction {
	return d.UnspentStore.Badger().NewTransaction(true)
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

func createDb(dbDir string, logger badger.Logger) (db *badgerhold.Store, err error) {
	opts := badger.DefaultOptions(dbDir)
	opts.Logger = logger

	db, err = badgerhold.Open(badgerhold.Options{
		Encoder:          JSONEncode,
		Decoder:          JSONDecode,
		SequenceBandwith: 100,
		Options:          opts,
	})

	return
}
