package dbbadger

import (
	"bytes"
	"encoding/json"
	"path/filepath"

	"github.com/dgraph-io/badger/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/timshannon/badgerhold/v2"
)

type DbManager struct {
	Store        *badgerhold.Store
	UnspentStore *badgerhold.Store
}

func NewDbManager(dbDir string, logger badger.Logger) (*DbManager, error) {
	db, err := newDb(filepath.Join(dbDir, "daemon"), logger)
	if err != nil {
		return nil, err
	}
	udb, err := newDb(filepath.Join(dbDir, "unspents"), logger)
	if err != nil {
		return nil, err
	}

	return &DbManager{
		Store:        db,
		UnspentStore: udb,
	}, nil
}

func (d DbManager) NewTransaction() ports.Transaction {
	return d.Store.Badger().NewTransaction(true)
}

func (d DbManager) NewUnspentsTransaction() ports.Transaction {
	return d.UnspentStore.Badger().NewTransaction(true)
}

func JsonEncode(value interface{}) ([]byte, error) {
	var buff bytes.Buffer

	en := json.NewEncoder(&buff)

	err := en.Encode(value)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func JsonDecode(data []byte, value interface{}) error {
	var buff bytes.Buffer
	de := json.NewDecoder(&buff)

	_, err := buff.Write(data)
	if err != nil {
		return err
	}

	return de.Decode(value)
}

func newDb(dbDir string, logger badger.Logger) (*badgerhold.Store, error) {
	opts := badger.DefaultOptions(dbDir)
	opts.Logger = logger
	return badgerhold.Open(badgerhold.Options{
		Encoder:          JsonEncode,
		Decoder:          JsonDecode,
		SequenceBandwith: 100,
		Options:          opts,
	})
}
