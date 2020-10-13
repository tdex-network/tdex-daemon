package application

import (
	"context"
	"encoding/hex"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
)

func TestGetPriceAndPreviewForMarket(t *testing.T) {
	t.Run("with market default strategy", func(t *testing.T) {
		defaultStrategy := true
		tt, err := mocksForPriceAndPreview(defaultStrategy)
		if err != nil {
			t.Fatal(err)
		}

		price, previewAmount, err := getPriceAndPreviewForMarket(
			context.Background(),
			tt.vaultRepo,
			tt.unspentRepo,
			tt.market,
			TradeBuy,
			tt.lbtcAmount,
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int(tt.expectedBuyAmount), int(previewAmount))
		assert.Equal(t, tt.expectedPrice.BasePrice.String(), price.BasePrice.String())
		assert.Equal(t, tt.expectedPrice.QuotePrice.String(), price.QuotePrice.String())

		_, previewAmount, err = getPriceAndPreviewForMarket(
			context.Background(),
			tt.vaultRepo,
			tt.unspentRepo,
			tt.market,
			TradeSell,
			tt.lbtcAmount,
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int(tt.expectedSellAmount), int(previewAmount))
	})

	t.Run("with market pluggable strategy", func(t *testing.T) {
		defaultStrategy := false
		tt, err := mocksForPriceAndPreview(defaultStrategy)
		if err != nil {
			t.Fatal(err)
		}

		price, previewAmount, err := getPriceAndPreviewForMarket(
			context.Background(),
			tt.vaultRepo,
			tt.unspentRepo,
			tt.market,
			TradeBuy,
			tt.lbtcAmount,
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int(tt.expectedBuyAmount), int(previewAmount))
		assert.Equal(t, tt.expectedPrice.BasePrice.String(), price.BasePrice.String())
		assert.Equal(t, tt.expectedPrice.QuotePrice.String(), price.QuotePrice.String())

		_, previewAmount, err = getPriceAndPreviewForMarket(
			context.Background(),
			tt.vaultRepo,
			tt.unspentRepo,
			tt.market,
			TradeSell,
			tt.lbtcAmount,
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int(tt.expectedSellAmount), int(previewAmount))
	})

}

type priceAndPreviewTestData struct {
	vaultRepo          domain.VaultRepository
	unspentRepo        domain.UnspentRepository
	market             *domain.Market
	lbtcAmount         uint64
	expectedBuyAmount  uint64
	expectedSellAmount uint64
	expectedPrice      Price
}

func mocksForPriceAndPreview(withDefaultStrategy bool) (*priceAndPreviewTestData, error) {
	// create market
	market, _ := domain.NewMarket(domain.MarketAccountStart)

	// derive addresses for funding market
	mnemonic := []string{"curtain", "summer", "juice", "thought", "release", "velvet", "dress", "fantasy", "price", "hard", "core", "friend", "reopen", "myth", "giant", "consider", "seminar", "ladder", "thought", "spell", "state", "home", "diamond", "gold"}
	passphrase := "Sup3rS3cr3tP4ssw0rd!"
	var addr string
	var script []byte
	closure := func(v *domain.Vault) (*domain.Vault, error) {
		a, s, _, err := v.DeriveNextExternalAddressForAccount(domain.MarketAccountStart)
		if err != nil {
			return nil, err
		}
		addr = a
		script, _ = hex.DecodeString(s)
		return v, nil
	}
	vaultRepo := inmemory.NewVaultRepositoryImpl()
	if err := vaultRepo.UpdateVault(context.Background(), mnemonic, passphrase, closure); err != nil {
		return nil, err
	}

	// persist unspents and fund market
	dbManager, err := dbbadger.NewDbManager("testdb")
	if err != nil {
		panic(err)
	}
	unspentRepo := dbbadger.NewUnspentRepositoryImpl(dbManager)
	unspentRepo.AddUnspents(context.Background(), []domain.Unspent{
		// 1 LBTC
		{
			TxID:         "0000000000000000000000000000000000000000000000000000000000000000",
			VOut:         0,
			Value:        0100000000,
			AssetHash:    config.GetNetwork().AssetID,
			Address:      addr,
			Spent:        false,
			Locked:       false,
			ScriptPubKey: script,
			LockedBy:     nil,
			Confirmed:    true,
		},
		// 6500 ASS
		{
			TxID:         "0000000000000000000000000000000000000000000000000000000000000000",
			VOut:         1,
			Value:        650000000000,
			AssetHash:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Address:      addr,
			Spent:        false,
			Locked:       false,
			ScriptPubKey: script,
			LockedBy:     nil,
			Confirmed:    true,
		},
	})

	market.FundMarket([]domain.OutpointWithAsset{
		// LBTC
		domain.OutpointWithAsset{
			Asset: config.GetNetwork().AssetID,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  0,
		},
		// ASS
		domain.OutpointWithAsset{
			Asset: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  1,
		},
	})

	bp, _ := decimal.NewFromString("0.00015385")
	qp, _ := decimal.NewFromString("6500")
	price := Price{
		BasePrice:  bp,
		QuotePrice: qp,
	}

	if withDefaultStrategy {
		market.MakeTradable()

		return &priceAndPreviewTestData{
			vaultRepo:          vaultRepo,
			unspentRepo:        unspentRepo,
			market:             market,
			lbtcAmount:         10000, // 0.0001 LBTC
			expectedBuyAmount:  65169016,
			expectedSellAmount: 65155984,
			expectedPrice:      price,
		}, nil
	}

	market.MakeStrategyPluggable()
	market.ChangeBasePrice(bp)
	market.ChangeQuotePrice(qp)

	return &priceAndPreviewTestData{
		vaultRepo:          vaultRepo,
		unspentRepo:        unspentRepo,
		market:             market,
		lbtcAmount:         10000, // 0.0001 LBTC
		expectedBuyAmount:  81250000,
		expectedSellAmount: 48750000,
		expectedPrice:      price,
	}, nil
}
