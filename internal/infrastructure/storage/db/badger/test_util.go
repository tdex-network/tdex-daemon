package dbbadger

import (
	"context"
	"os"

	"github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
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
	tx := dbManager.NewTransaction()

	if err = insertMarkets(tx.(*badger.Txn), dbManager); err != nil {
		panic(err)
	}
	if err = insertUnspents(tx.(*badger.Txn), dbManager); err != nil {
		panic(err)
	}
	if err = insertTrades(tx.(*badger.Txn), dbManager); err != nil {
		panic(err)
	}

	if err = insertVault(tx.(*badger.Txn), dbManager); err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	ctx = context.WithValue(
		context.Background(),
		"tx",
		dbManager.NewTransaction(),
	)

}

func after() {
	tx := ctx.Value("tx").(*badger.Txn)
	if err := tx.Commit(); err != nil {
		panic(err)
	}
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

func insertVault(tx *badger.Txn, db *DbManager) error {
	vault := &domain.Vault{
		Mnemonic: []string{
			"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
			"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
			"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
		},
		EncryptedMnemonic:      "dVoBFte1oeRkPl8Vf8DzBP3PRnzPA3fxtyvDHXFGYAS9MP8V2Sc9nHcQW4PrMkQNnf2uGrDg81dFgBrwqv1n3frXxRBKhp83fSsTm4xqj8+jdwTI3nouFmi1W/O4UqpHdQ62EYoabJQtKpptWO11TFJzw8WF02pfS6git8YjLR4xrnfp2LkOEjSU9CI82ZasF46WZFKcpeUJTAsxU/03ONpAdwwEsC96f1KAvh8tqaO0yLDOcmPf8a5B82jefgncCRrt32kCpbpIE4YiCFrqqdUHXKH+",
		PassphraseHash:         []byte("pass"),
		Accounts:               map[int]*domain.Account{},
		AccountAndKeyByAddress: map[string]domain.AccountAndKey{},
	}

	return db.Store.TxInsert(tx, vaultKey, vault)
}

func insertTrades(tx *badger.Txn, db *DbManager) error {
	tradeID1, _ := uuid.Parse("cc913d4e-174e-449c-82b4-e848d57cbf2e")
	tradeID2, _ := uuid.Parse("5440a53e-58d2-4e3d-8380-20410e687589")
	tradeID3, _ := uuid.Parse("2a12e2a0-d99c-4bd3-ad99-03dd926ae080")

	trades := []domain.Trade{
		{
			ID:               tradeID1,
			MarketQuoteAsset: "mqa1",
			SwapRequest: domain.Swap{
				ID: "1",
			},
			SwapAccept: domain.Swap{
				ID: "2",
			},
			SwapComplete: domain.Swap{
				ID: "3",
			},
			SwapFail: domain.Swap{
				ID: "4",
			},
		},
		{
			ID:               tradeID2,
			MarketQuoteAsset: "mqa2",
			SwapRequest: domain.Swap{
				ID: "11",
			},
			SwapAccept: domain.Swap{
				ID: "21",
			},
			SwapComplete: domain.Swap{
				ID: "31",
			},
			SwapFail: domain.Swap{
				ID: "41",
			},
		},
		{
			ID:               tradeID3,
			MarketQuoteAsset: "mqa2",
			TxID:             "424",
			SwapRequest: domain.Swap{
				ID: "12",
			},
			SwapAccept: domain.Swap{
				ID: "22",
			},
			SwapComplete: domain.Swap{
				ID: "32",
			},
			SwapFail: domain.Swap{
				ID: "42",
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
