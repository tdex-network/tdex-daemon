package v0domain

import (
	"fmt"
	"path/filepath"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/timshannon/badgerhold/v2"
)

const (
	vaultKey = "vault"
)

type TdexRepoManager interface {
	GetVaultRepository() VaultRepository
	MarketRepository() MarketRepository
	GetTradeRepository() TradeRepository
	GetDepositRepository() DepositRepository
	GetWithdrawalRepository() WithdrawalRepository
	GetMarketRepository() MarketRepository
	GetUnspentRepository() UnspentRepository
}

type repoManager struct {
	vaultRepository      VaultRepository
	marketRepository     MarketRepository
	tradeRepository      TradeRepository
	depositRepository    DepositRepository
	withdrawalRepository WithdrawalRepository
	unspentRepository    UnspentRepository
}

func NewTdexRepoManager(
	dbDir string, logger badger.Logger,
) (TdexRepoManager, error) {
	mainDb, err := createDb(filepath.Join(dbDir, "main"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening main db: %w", err)
	}

	pricesDb, err := createDb(filepath.Join(dbDir, "prices"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening prices db: %w", err)
	}

	unspentDb, err := createDb(filepath.Join(dbDir, "unspents"), logger)
	if err != nil {
		return nil, fmt.Errorf("opening unspents db: %w", err)
	}

	return &repoManager{
		vaultRepository:      NewVaultRepositoryImpl(mainDb),
		marketRepository:     NewMarketRepositoryImpl(mainDb, pricesDb),
		tradeRepository:      NewTradeRepositoryImpl(mainDb),
		depositRepository:    NewDepositRepositoryImpl(mainDb),
		withdrawalRepository: NewWithdrawalRepositoryImpl(mainDb),
		unspentRepository:    NewUnspentRepositoryImpl(unspentDb, mainDb),
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

func (r *repoManager) GetDepositRepository() DepositRepository {
	return r.depositRepository
}

func (r *repoManager) GetWithdrawalRepository() WithdrawalRepository {
	return r.withdrawalRepository
}

func (r *repoManager) GetMarketRepository() MarketRepository {
	return r.marketRepository
}

func (r *repoManager) GetUnspentRepository() UnspentRepository {
	return r.unspentRepository
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
