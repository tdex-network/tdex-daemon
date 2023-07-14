package v1domain

import (
	"github.com/timshannon/badgerhold/v4"
)

type WalletRepository interface {
	InsertWallet(wallet *Wallet) error
	GetWallet() (*Wallet, error)
}

type walletRepositoryImpl struct {
	store *badgerhold.Store
}

func NewWalletRepositoryImpl(store *badgerhold.Store) WalletRepository {
	return &walletRepositoryImpl{store}
}

func (w *walletRepositoryImpl) InsertWallet(wallet *Wallet) error {
	if err := w.store.Insert(walletKey, *wallet); err != nil {
		return err
	}

	return nil
}

func (w *walletRepositoryImpl) GetWallet() (*Wallet, error) {
	var wallet Wallet
	if err := w.store.Get(walletKey, &wallet); err != nil {
		return nil, err
	}

	return &wallet, nil
}
