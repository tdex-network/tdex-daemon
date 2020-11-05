package application

import (
	"context"
	"fmt"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/vulpemventures/go-elements/network"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

const marketRepoIsEmpty = true
const tradeRepoIsEmpty = true

var baseAsset = config.GetString(config.BaseAssetKey)

func TestListMarket(t *testing.T) {
	t.Run("ListMarket should return an empty list and a nil error if market repository is empty", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)
		marketInfos, err := operatorService.ListMarket(ctx)
		close()
		assert.Equal(t, nil, err)
		assert.Equal(t, 0, len(marketInfos))
	})

	t.Run("ListMarket should return the number of markets in the market repository", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(!marketRepoIsEmpty, tradeRepoIsEmpty)
		marketInfos, err := operatorService.ListMarket(ctx)
		close()
		assert.Equal(t, nil, err)
		assert.Equal(t, 2, len(marketInfos))
	})

}

func TestDepositMarket(t *testing.T) {

	t.Run("DepositMarket with new market", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)

		address, err := operatorService.DepositMarket(ctx, "", "")
		assert.Equal(t, nil, err)

		assert.Equal(
			t,
			"el1qqvead5fpxkjyyl3zwukr7twqrnag40ls0y052s547smxdyeus209ppkmtdyemgkz4rjn8ss8fhjrzc3q9evt7atrgtpff2thf",
			address,
		)

		close()
	})

	t.Run("DepositMarket with invalid base asset", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)

		validQuoteAsset := "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"
		emptyAddress, err := operatorService.DepositMarket(ctx, "", validQuoteAsset)
		assert.Equal(t, domain.ErrInvalidBaseAsset, err)
		assert.Equal(
			t,
			"",
			emptyAddress,
		)

		close()
	})

	t.Run("DepositMarket with valid base asset and empty quote asset", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)

		emptyAddress, err := operatorService.DepositMarket(ctx, baseAsset, "")
		assert.Equal(t, domain.ErrInvalidQuoteAsset, err)
		assert.Equal(
			t,
			"",
			emptyAddress,
		)

		close()
	})

	t.Run("DepositMarket with valid base asset and invalid quote asset", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)

		emptyAddress, err := operatorService.DepositMarket(ctx, baseAsset, "ldjbwjkbfjksdbjkvcsbdjkbcdsjkb")
		assert.Equal(t, domain.ErrInvalidQuoteAsset, err)
		assert.Equal(
			t,
			"",
			emptyAddress,
		)

		close()
	})
}

func TestDepositMarketWithCrawler(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("Get address to deposit, fund market and get next address for the market", func(t *testing.T) {

		startNigiriAndWait()

		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)

		address, err := operatorService.DepositMarket(ctx, "", "")
		assert.Equal(t, nil, err)

		assert.Equal(t, nil, err)
		assert.Equal(
			t,
			"el1qqvead5fpxkjyyl3zwukr7twqrnag40ls0y052s547smxdyeus209ppkmtdyemgkz4rjn8ss8fhjrzc3q9evt7atrgtpff2thf",
			address,
		)

		// Let's depsoit both assets on the same address
		explorerSvc := explorer.NewService(RegtestExplorerAPI)
		_, err = explorerSvc.Faucet(address)
		assert.Equal(t, nil, err)
		time.Sleep(1500 * time.Millisecond)

		_, quoteAsset, err := explorerSvc.Mint(address, 5)
		assert.Equal(t, nil, err)
		time.Sleep(1500 * time.Millisecond)

		// we try to get a child address for the quote asset. Since is not being expicitly initialized, should return ErrMarketNotExist
		failToGetChildAddress, err := operatorService.DepositMarket(ctx, baseAsset, quoteAsset)
		assert.Equal(t, domain.ErrMarketNotExist, err)
		assert.Equal(
			t,
			"",
			failToGetChildAddress,
		)

		// Now we try to intialize (ie. fund) the market by opening it
		err = operatorService.OpenMarket(ctx, baseAsset, quoteAsset)
		assert.Equal(t, nil, err)

		// Now we can derive a childAddress
		childAddress, err := operatorService.DepositMarket(ctx, baseAsset, quoteAsset)
		assert.Equal(t, nil, err)
		assert.Equal(
			t,
			"el1qqfzjp0y057j60avxqgmj9aycqhlq7ke20v20c8dkml68jjs0fu09u9sn55uduay46yyt25tcny0rfqejly5x6dgjw44uk9p8r",
			childAddress,
		)

		close()
		stopNigiri()
	})
}

func TestListSwap(t *testing.T) {
	t.Run("ListSwap should return an empty list and a nil error if there is not trades in TradeRepository", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty, tradeRepoIsEmpty)
		defer close()

		swapInfos, err := operatorService.ListSwaps(ctx)
		assert.Equal(t, nil, err)
		assert.Equal(t, 0, len(swapInfos))
	})

	t.Run("ListSwap should return the SwapInfo according to the number of trades in the TradeRepository", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(!marketRepoIsEmpty, !tradeRepoIsEmpty)
		defer close()

		swapInfos, err := operatorService.ListSwaps(ctx)
		assert.Equal(t, nil, err)
		assert.Equal(t, 1, len(swapInfos))
	})
}

func TestWithdrawMarket(t *testing.T) {
	dbManager, err := mockDb()
	if err != nil {
		t.Fatal(err)
	}

	vaultRepo := dbbadger.NewVaultRepositoryImpl(dbManager)
	unspentRepo := dbbadger.NewUnspentRepositoryImpl(dbManager)
	crawlerSvc := crawler.NewService(crawler.Opts{
		ExplorerSvc:            nil,
		Observables:            []crawler.Observable{},
		ErrorHandler:           func(err error) { fmt.Println(err) },
		IntervalInMilliseconds: 100,
	})
	marketRepo := dbbadger.NewMarketRepositoryImpl(dbManager)

	operatorService := NewOperatorService(
		marketRepo,
		vaultRepo,
		nil,
		unspentRepo,
		nil,
		crawlerSvc,
	)

	t.Run(
		"WithdrawMarketFunds should return raw transaction",
		func(t *testing.T) {
			tx := dbManager.NewTransaction()
			ctx := context.WithValue(context.Background(), "tx", tx)
			rawTx, err := operatorService.WithdrawMarketFunds(ctx, WithdrawMarketReq{
				Market: Market{
					BaseAsset:  "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
					QuoteAsset: "d73f5cd0954c1bf325f85d7a7ff43a6eb3ea3b516fd57064b85306d43bc1c9ff",
				},
				BalanceToWithdraw: Balance{
					BaseAmount:  4200,
					QuoteAmount: 2300,
				},
				MillisatPerByte: 20,
				Address:         "el1qq22f83p6asdy7jsp4tuke0d9emvxhcenqee5umsn88fsn8gggzlrx0md4hp38rnwcnu9lusmzhmktlt3h5q0gecfpfvx6uac2",
				Push:            false,
			})
			assert.NoError(t, err)
			assert.Equal(t, true, len(rawTx) > 0)
		},
	)

	t.Run(
		"WithdrawMarketFunds should return error for wrong base asset",
		func(t *testing.T) {
			tx := dbManager.NewTransaction()
			ctx := context.WithValue(context.Background(), "tx", tx)
			_, err := operatorService.WithdrawMarketFunds(ctx,
				WithdrawMarketReq{
					Market: Market{
						BaseAsset:  "4144",
						QuoteAsset: "d73f5cd0954c1bf325f85d7a7ff43a6eb3ea3b516fd57064b85306d43bc1c9ff",
					},
					BalanceToWithdraw: Balance{
						BaseAmount:  4200,
						QuoteAmount: 2300,
					},
					MillisatPerByte: 20,
					Address:         "el1qq22f83p6asdy7jsp4tuke0d9emvxhcenqee5umsn88fsn8gggzlrx0md4hp38rnwcnu9lusmzhmktlt3h5q0gecfpfvx6uac2",
					Push:            false,
				})
			assert.Error(t, err)
			assert.Equal(t, err, domain.ErrInvalidBaseAsset)
		},
	)

	t.Run(
		"WithdrawMarketFunds should return error for wrong qoute asset",
		func(t *testing.T) {
			tx := dbManager.NewTransaction()
			ctx := context.WithValue(context.Background(), "tx", tx)
			_, err := operatorService.WithdrawMarketFunds(ctx,
				WithdrawMarketReq{
					Market: Market{
						BaseAsset:  "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
						QuoteAsset: "eqwqw",
					},
					BalanceToWithdraw: Balance{
						BaseAmount:  4200,
						QuoteAmount: 2300,
					},
					MillisatPerByte: 20,
					Address:         "el1qq22f83p6asdy7jsp4tuke0d9emvxhcenqee5umsn88fsn8gggzlrx0md4hp38rnwcnu9lusmzhmktlt3h5q0gecfpfvx6uac2",
					Push:            false,
				})
			assert.Error(t, err)
			assert.Equal(t, err, domain.ErrMarketNotExist)
		},
	)

	t.Run(
		"WithdrawMarketFunds should return error, not enough money",
		func(t *testing.T) {
			tx := dbManager.NewTransaction()
			ctx := context.WithValue(context.Background(), "tx", tx)
			_, err := operatorService.WithdrawMarketFunds(ctx,
				WithdrawMarketReq{
					Market: Market{
						BaseAsset:  "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
						QuoteAsset: "d73f5cd0954c1bf325f85d7a7ff43a6eb3ea3b516fd57064b85306d43bc1c9ff",
					},
					BalanceToWithdraw: Balance{
						BaseAmount:  1000000000000,
						QuoteAmount: 2300,
					},
					MillisatPerByte: 20,
					Address:         "el1qq22f83p6asdy7jsp4tuke0d9emvxhcenqee5umsn88fsn8gggzlrx0md4hp38rnwcnu9lusmzhmktlt3h5q0gecfpfvx6uac2",
					Push:            false,
				})
			assert.Error(t, err)
		},
	)

	dbManager.Store.Close()
	dbManager.UnspentStore.Close()
	os.RemoveAll(testDir)
}

func TestBalanceFeeAccount(t *testing.T) {
	dbManager, err := mockDb()
	if err != nil {
		t.Fatal(err)
	}

	vaultRepo := dbbadger.NewVaultRepositoryImpl(dbManager)
	unspentRepo := dbbadger.NewUnspentRepositoryImpl(dbManager)
	crawlerSvc := crawler.NewService(crawler.Opts{
		ExplorerSvc:            nil,
		Observables:            []crawler.Observable{},
		ErrorHandler:           func(err error) { fmt.Println(err) },
		IntervalInMilliseconds: 100,
	})
	marketRepo := dbbadger.NewMarketRepositoryImpl(dbManager)

	operatorService := NewOperatorService(
		marketRepo,
		vaultRepo,
		nil,
		unspentRepo,
		nil,
		crawlerSvc,
	)

	t.Run(
		"FeeAccountBalance should return fee account balance",
		func(t *testing.T) {
			tx := dbManager.NewTransaction()
			ctx := context.WithValue(context.Background(), "tx", tx)
			balance, err := operatorService.FeeAccountBalance(ctx)

			assert.NoError(t, err)
			assert.Equal(t, balance, int64(100000000))
		},
	)

	dbManager.Store.Close()
	dbManager.UnspentStore.Close()
	os.RemoveAll(testDir)
}

func TestGetCollectedMarketFee(t *testing.T) {

	operatorService, ctx, closeOperator := newTestOperator(
		marketRepoIsEmpty,
		tradeRepoIsEmpty,
	)

	defer closeOperator()

	traderSvc, ctx, closeTrader := newTestTrader()
	defer closeTrader()

	markets, err := traderSvc.GetTradableMarkets(ctx)
	if err != nil {
		t.Fatal(err)
	}

	market := markets[0].Market
	preview, err := traderSvc.GetMarketPrice(ctx, market, TradeSell, 30000000)
	if err != nil {
		t.Fatal(err)
	}

	proposerWallet, err := trade.NewRandomWallet(&network.Regtest)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("GetCollectedMarketFee", func(t *testing.T) {
		swapRequest, err := newSwapRequest(
			proposerWallet,
			market.BaseAsset, 30000000,
			market.QuoteAsset, preview.Amount,
		)
		if err != nil {
			t.Fatal(err)
		}

		_, swapFail, _, err := traderSvc.TradePropose(
			ctx,
			market,
			TradeSell,
			swapRequest,
		)
		if err != nil {
			t.Fatal(err)
		}
		if swapFail != nil {
			t.Fatal(swapFail.GetFailureMessage())
		}

		fee, err := operatorService.GetCollectedMarketFee(ctx, market)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(fee.CollectedFees))

		fee, err = operatorService.GetCollectedMarketFee(ctx, market)
		if err != nil {
			t.Fatal(err)
		}

		tradeRepo := dbbadger.NewTradeRepositoryImpl(dbManager)
		trades, err := tradeRepo.GetAllTradesByMarket(ctx, market.QuoteAsset)
		if err != nil {
			t.Fatal(err)
		}

		tr := trades[0]
		err = tradeRepo.UpdateTrade(
			ctx,
			&tr.ID,
			func(trade *domain.Trade) (*domain.Trade, error) {
				trade.Status = domain.CompletedStatus
				return trade, nil
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		fee, err = operatorService.GetCollectedMarketFee(ctx, market)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(fee.CollectedFees))
		assert.Equal(
			t,
			int64(25),
			fee.TotalCollectedFeesPerAsset[network.Regtest.AssetID],
		)
	})

}
