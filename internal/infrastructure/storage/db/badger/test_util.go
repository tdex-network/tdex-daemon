package dbbadger

import (
	"context"
	"github.com/dgraph-io/badger"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
	"os"
)

var ctx context.Context
var marketRepository domain.MarketRepository
var unspentRepository domain.UnspentRepository
var vaultRepository domain.VaultRepository
var dbManager *DbManager
var testDbDir = "testdb"

func before() {
	var err error

	dbManager, err = NewDbManager(testDbDir)
	if err != nil {
		panic(err)
	}
	marketRepository = NewMarketRepositoryImpl(dbManager)
	unspentRepository = NewUnspentRepositoryImpl(dbManager)
	vaultRepository = NewVaultRepositoryImpl(dbManager)
	tx := dbManager.Store.Badger().NewTransaction(true)

	if err = insertMarkets(tx, dbManager); err != nil {
		panic(err)
	}
	if err = insertUnspents(tx, dbManager); err != nil {
		panic(err)
	}

	tmpCtx := context.WithValue(
		context.Background(),
		"tx",
		tx,
	)
	if err = insertVault(tmpCtx); err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	ctx = context.WithValue(
		context.Background(),
		"tx",
		dbManager.Store.Badger().NewTransaction(true))

}

func after() {
	tx := ctx.Value("tx").(*badger.Txn)
	if err := tx.Commit(); err != nil {
		panic(err)
	}
	tx.Discard()
	dbManager.Store.Close()

	err := os.RemoveAll(testDbDir)
	if err != nil {
		panic(err)
	}
}

func insertMarkets(tx *badger.Txn, db *DbManager) error {
	markets := []domain.Market{
		{
			AccountIndex: 5,
			BaseAsset:    "ah5",
			QuoteAsset:   "qh5",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     true,
			Strategy:     mm.NewStrategyFromFormula(formula.BalancedReserves{}),
			BasePrice:    domain.PriceByTime{},
			QuotePrice:   domain.PriceByTime{},
		},
		{
			AccountIndex: 6,
			BaseAsset:    "ah6",
			QuoteAsset:   "qh6",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     true,
			Strategy:     mm.NewStrategyFromFormula(formula.BalancedReserves{}),
			BasePrice:    domain.PriceByTime{},
			QuotePrice:   domain.PriceByTime{},
		},
		{
			AccountIndex: 7,
			BaseAsset:    "ah7",
			QuoteAsset:   "qh7",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     false,
			Strategy:     mm.NewStrategyFromFormula(formula.BalancedReserves{}),
			BasePrice:    domain.PriceByTime{},
			QuotePrice:   domain.PriceByTime{},
		},
		{
			AccountIndex: 8,
			BaseAsset:    "ah8",
			QuoteAsset:   "qh8",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     false,
			Strategy:     mm.NewStrategyFromFormula(formula.BalancedReserves{}),
			BasePrice:    domain.PriceByTime{},
			QuotePrice:   domain.PriceByTime{},
		},
		{
			AccountIndex: 9,
			BaseAsset:    "ah9",
			QuoteAsset:   "qh9",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     false,
			Strategy:     mm.NewStrategyFromFormula(formula.BalancedReserves{}),
			BasePrice:    domain.PriceByTime{},
			QuotePrice:   domain.PriceByTime{},
		},
	}
	for _, v := range markets {
		err := db.Store.TxInsert(tx, v.AccountIndex, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func insertUnspents(tx *badger.Txn, db *DbManager) error {
	unspents := []domain.Unspent{
		{
			TxID:         "1",
			VOut:         0,
			Value:        4,
			AssetHash:    "ah",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    true,
		},
		{
			TxID:         "1",
			VOut:         1,
			Value:        2,
			AssetHash:    "ah",
			Address:      "adr",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    true,
		},
		{
			TxID:         "2",
			VOut:         1,
			Value:        4,
			AssetHash:    "ah",
			Address:      "adre",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
		{
			TxID:         "2",
			VOut:         2,
			Value:        9,
			AssetHash:    "ah",
			Address:      "adra",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
		{
			TxID:         "3",
			VOut:         1,
			Value:        4,
			AssetHash:    "ah",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
		{
			TxID:         "3",
			VOut:         0,
			Value:        2,
			AssetHash:    "ah",
			Address:      "a",
			Spent:        false,
			Locked:       false,
			ScriptPubKey: nil,
			LockedBy:     nil,
			Confirmed:    false,
		},
	}
	for _, v := range unspents {
		if err := db.Store.TxInsert(
			tx,
			v.Key(),
			&v,
		); err != nil {
			return err
		}
	}

	return nil
}

func insertVault(ctx context.Context) error {

	pass := "pass"
	mnemonic := []string{
		"addict",
		"able",
		"about",
		"above",
		"absent",
		"absorb",
		"abstract",
		"absurd",
		"abuse",
		"access",
		"accident",
		"account",
	}

	_, err := vaultRepository.GetOrCreateVault(ctx, mnemonic, pass)
	return err
}
