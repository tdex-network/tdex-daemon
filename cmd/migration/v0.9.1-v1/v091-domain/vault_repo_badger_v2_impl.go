package v091domain

import (
	"path/filepath"

	"github.com/sekulicd/badger/v2"
	"github.com/sekulicd/badger/v2/options"

	"github.com/sekulicd/badgerhold/v2"
)

const (
	vaultKey = "vault"
)

type Repository interface {
	GetVaultRepository() VaultRepository
}

type repoManager struct {
	vaultRepository VaultRepository
}

func NewRepositoryImpl(
	dbDir string, logger badger.Logger,
) (Repository, error) {
	mainDb, err := createDb(filepath.Join(dbDir, "main"), logger)
	if err != nil {
		return nil, err
	}

	return &repoManager{
		vaultRepository: NewVaultRepositoryImpl(mainDb),
	}, nil
}

func (r *repoManager) GetVaultRepository() VaultRepository {
	return r.vaultRepository
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
