package v091domain

import (
	"fmt"
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
	MarketRepository() MarketRepository
	GetTradeRepository() TradeRepository
}

type repoManager struct {
	vaultRepository  VaultRepository
	marketRepository MarketRepository
	tradeRepository  TradeRepository
}

func NewRepositoryImpl(
	dbDir string, logger badger.Logger,
) (Repository, error) {
	mainDb, err := createDb(filepath.Join(dbDir, "main"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening main db: %w", err)
	}

	pricesDb, err := createDb(filepath.Join(dbDir, "prices"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening prices db: %w", err)
	}

	return &repoManager{
		vaultRepository:  NewVaultRepositoryImpl(mainDb),
		marketRepository: NewMarketRepositoryImpl(mainDb, pricesDb),
		tradeRepository:  NewTradeRepositoryImpl(mainDb),
	}, nil
}

func (r *repoManager) GetVaultRepository() VaultRepository {
	return r.vaultRepository
}

func (r *repoManager) MarketRepository() MarketRepository {
	return r.marketRepository
}

func (r *repoManager) GetTradeRepository() TradeRepository {
	return r.tradeRepository
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
