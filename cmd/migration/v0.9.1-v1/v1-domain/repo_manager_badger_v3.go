package v1domain

import (
	"fmt"
	"path/filepath"

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
	oceanDbDir, tdexdDbDir string, logger badger.Logger,
) (Repository, error) {
	walletDb, err := createDb(filepath.Join(oceanDbDir, "wallet"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening wallet db: %w", err)
	}

	return &repoManager{
		walletRepository: NewWalletRepositoryImpl(walletDb),
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
