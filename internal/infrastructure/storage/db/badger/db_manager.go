package dbbadger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/timshannon/badgerhold"
)

type DbManager struct {
	Store *badgerhold.Store
}

func NewDbManager(dbDir string) (*DbManager, error) {
	opts := badger.DefaultOptions(dbDir)
	opts.Logger = nil
	db, err := badgerhold.Open(badgerhold.Options{
		Encoder:          JsonEncode,
		Decoder:          JsonDecode,
		SequenceBandwith: 100,
		Options:          opts,
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
