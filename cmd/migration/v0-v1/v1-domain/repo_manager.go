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

type OceanRepoManager interface {
	WalletRepository() WalletRepository
	UtxoRepository() UtxoRepository
	TransactionRepository() TransactionRepository
}

type repoManager struct {
	walletRepository      WalletRepository
	utxoRepository        UtxoRepository
	transactionRepository TransactionRepository
}

func NewOceanRepoManager(oceanDbDir string) (OceanRepoManager, error) {
	walletDb, err := createDb(filepath.Join(oceanDbDir, "wallet"), nil)
	if err != nil {
		return nil, fmt.Errorf("opening wallet db: %w", err)
	}

	utxoDb, err := createDb(filepath.Join(oceanDbDir, "utxos"), nil)
	if err != nil {
		return nil, fmt.Errorf("opening utxo db: %w", err)
	}

	txDb, err := createDb(filepath.Join(oceanDbDir, "txs"), nil)
	if err != nil {
		return nil, fmt.Errorf("opening tx db: %w", err)
	}

	return &repoManager{
		walletRepository:      NewWalletRepositoryImpl(walletDb),
		utxoRepository:        NewUtxoRepositoryImpl(utxoDb),
		transactionRepository: NewTransactionRepositoryImpl(txDb),
	}, nil
}

func (r *repoManager) WalletRepository() WalletRepository {
	return r.walletRepository
}

func (r *repoManager) UtxoRepository() UtxoRepository {
	return r.utxoRepository
}

func (r *repoManager) TransactionRepository() TransactionRepository {
	return r.transactionRepository
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
