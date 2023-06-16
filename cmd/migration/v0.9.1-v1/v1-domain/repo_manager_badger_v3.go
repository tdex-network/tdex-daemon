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
	GetTradeRepository() TradeRepository
	GetDepositsRepository() DepositRepository
}

type repoManager struct {
	walletRepository   WalletRepository
	tradeRepository    TradeRepository
	depositsRepository DepositRepository
}

func NewRepositoryImpl(
	oceanDbDir, tdexdDbDir string, logger badger.Logger,
) (Repository, error) {
	walletDb, err := createDb(filepath.Join(oceanDbDir, "wallet"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening wallet db: %w", err)
	}

	tradeDb, err := createDb(filepath.Join(tdexdDbDir, "trades"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening unspents db: %w", err)
	}

	txDb, err := createDb(filepath.Join(tdexdDbDir, "transactions"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening unspents db: %w", err)
	}

	return &repoManager{
		walletRepository:   NewWalletRepositoryImpl(walletDb),
		tradeRepository:    NewTradeRepositoryImpl(tradeDb),
		depositsRepository: NewDepositRepositoryImpl(txDb),
	}, nil
}

func (r *repoManager) GetWalletRepository() WalletRepository {
	return r.walletRepository
}

func (r *repoManager) GetTradeRepository() TradeRepository {
	return r.tradeRepository
}

func (r *repoManager) GetDepositsRepository() DepositRepository {
	return r.depositsRepository
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
