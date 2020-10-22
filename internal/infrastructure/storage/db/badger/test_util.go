package dbbadger

import (
	"context"
	"github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
	"os"
)

var ctx context.Context
var marketRepository domain.MarketRepository
var unspentRepository domain.UnspentRepository
var vaultRepository domain.VaultRepository
var tradeRepository domain.TradeRepository
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
	tradeRepository = NewTradeRepositoryImpl(dbManager)
	tx := dbManager.Store.Badger().NewTransaction(true)

	if err = insertMarkets(tx, dbManager); err != nil {
		panic(err)
	}
	if err = insertUnspents(tx, dbManager); err != nil {
		panic(err)
	}
	if err = insertTrades(tx, dbManager); err != nil {
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

func insertTrades(tx *badger.Txn, db *DbManager) error {
	tradeID1, err := uuid.Parse("cc913d4e-174e-449c-82b4-e848d57cbf2e")
	if err != nil {
		return err
	}
	tradeID2, err := uuid.Parse("5440a53e-58d2-4e3d-8380-20410e687589")
	if err != nil {
		return err
	}
	tradeID3, err := uuid.Parse("2a12e2a0-d99c-4bd3-ad99-03dd926ae080")
	if err != nil {
		return err
	}
	trades := []domain.Trade{
		{
			ID:               tradeID1,
			MarketQuoteAsset: "mqa1",
			TraderPubkey:     nil,
			Status:           domain.Status{},
			PsetBase64:       "",
			TxID:             "",
			Price:            0,
			Timestamp:        domain.Timestamp{},
			SwapRequest: domain.Swap{
				ID:      "1",
				Message: nil,
			},
			SwapAccept: domain.Swap{
				ID:      "2",
				Message: nil,
			},
			SwapComplete: domain.Swap{
				ID:      "3",
				Message: nil,
			},
			SwapFail: domain.Swap{
				ID:      "4",
				Message: nil,
			},
		},
		{
			ID:               tradeID2,
			MarketQuoteAsset: "mqa2",
			TraderPubkey:     nil,
			Status:           domain.Status{},
			PsetBase64:       "",
			TxID:             "",
			Price:            0,
			Timestamp:        domain.Timestamp{},
			SwapRequest: domain.Swap{
				ID:      "11",
				Message: nil,
			},
			SwapAccept: domain.Swap{
				ID:      "21",
				Message: nil,
			},
			SwapComplete: domain.Swap{
				ID:      "31",
				Message: nil,
			},
			SwapFail: domain.Swap{
				ID:      "4",
				Message: nil,
			},
		},
		{
			ID:               tradeID3,
			MarketQuoteAsset: "mqa2",
			TraderPubkey:     nil,
			Status:           domain.Status{},
			PsetBase64:       "",
			TxID:             "424",
			Price:            0,
			Timestamp:        domain.Timestamp{},
			SwapRequest: domain.Swap{
				ID:      "12",
				Message: nil,
			},
			SwapAccept: domain.Swap{
				ID:      "22",
				Message: nil,
			},
			SwapComplete: domain.Swap{
				ID:      "32",
				Message: nil,
			},
			SwapFail: domain.Swap{
				ID:      "42",
				Message: nil,
			},
		},
	}

	for _, v := range trades {
		if err := db.Store.TxInsert(
			tx,
			v.ID,
			&v,
		); err != nil {
			return err
		}
	}

	return nil
}
