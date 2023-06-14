package v1domain

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"github.com/timshannon/badgerhold/v4"
)

const (
	walletKey = "wallet"
)

type Repository interface {
	GetWalletRepository() WalletRepository
}

type repoManager struct {
	walletRepository WalletRepository
}

func NewRepositoryImpl(
	dbDir string, logger badger.Logger,
) (Repository, error) {
	//TODO ocean datadir
	opts := badger.DefaultOptions(dbDir)
	opts.Logger = logger
	opts.Compression = options.ZSTD

	store, err := badgerhold.Open(badgerhold.Options{
		Encoder:          badgerhold.DefaultEncode,
		Decoder:          badgerhold.DefaultDecode,
		SequenceBandwith: 100,
		Options:          opts,
	})
	if err != nil {
		return nil, err
	}

	return &repoManager{
		walletRepository: NewWalletRepositoryImpl(store),
	}, nil
}

func (r *repoManager) GetWalletRepository() WalletRepository {
	return r.walletRepository
}

func createDb(dbDir string, logger badger.Logger) (*badgerhold.Store, error) {
	opts := badger.DefaultOptions(dbDir)
	opts.Logger = logger
	opts.Compression = options.ZSTD

	return badgerhold.Open(badgerhold.Options{
		Encoder:          badgerhold.DefaultEncode,
		Decoder:          badgerhold.DefaultDecode,
		SequenceBandwith: 100,
		Options:          opts,
	})
}
