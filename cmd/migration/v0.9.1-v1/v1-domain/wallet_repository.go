package v1domain

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/timshannon/badgerhold/v4"
)

type WalletRepository interface {
	InsertWallet(ctx context.Context, wallet *Wallet) error
}

type walletRepositoryImpl struct {
	store *badgerhold.Store
}

func NewWalletRepositoryImpl(store *badgerhold.Store) WalletRepository {
	return &walletRepositoryImpl{store}
}

func (w *walletRepositoryImpl) InsertWallet(
	ctx context.Context, wallet *Wallet,
) error {
	var err error

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = w.store.TxInsert(tx, walletKey, *wallet)
	} else {
		err = w.store.Insert(walletKey, *wallet)
	}
	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return fmt.Errorf("wallet is already initialized")
		}
		return err
	}

	return nil
}
