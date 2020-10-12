package dbbadger

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/timshannon/badgerhold"
)

type DbManager struct {
	Store *badgerhold.Store
}

func NewDbManager(dbDir string) (*DbManager, error) {
	db, err := badgerhold.Open(badgerhold.Options{
		Encoder:          badgerhold.DefaultEncode,
		Decoder:          badgerhold.DefaultDecode,
		SequenceBandwith: 100,
		Options:          badger.DefaultOptions(dbDir),
	})
	if err != nil {
		fmt.Println("Error instance db: ", err)
	}

	return &DbManager{
		Store: db,
	}, nil
}

func (d DbManager) NewTransaction() ports.Transaction {
	return d.Store.Badger().NewTransaction(true)
}
