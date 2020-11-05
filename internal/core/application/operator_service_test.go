package application

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"

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
			"el1qqfzjp0y057j60avxqgmj9aycqhlq7ke20v20c8dkml68jjs0fu09u9sn55uduay46yyt25tcny0rfqejly5x6dgjw44uk9p8r",
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

func TestListMarketExternalAddresses(t *testing.T) {
	const (
		validQuoteAsset = "d090c403610fe8a9e31967355929833bc8a8fe08429e630162d1ecbf29fdf28b"
		validBaseAsset = "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"
		validQuoteAssetWithNoMarket = "0ddfa690c7b2ba3b8ecee8200da2420fc502f57f8312c83d466b6f8dced70441"
		invalidAsset = "aaa001zzzDL"
	)

	listMarketExternalRequest := func(
		baseAsset string, 
		quoteAsset string,
	) ([]string, error) {
		operatorService, ctx, close := newTestOperator(!marketRepoIsEmpty, tradeRepoIsEmpty)
		defer close()
		market := Market{
			QuoteAsset: quoteAsset,
			BaseAsset: baseAsset,
		}
		return operatorService.ListMarketExternalAddresses(ctx, market)
	}


	t.Run("should return error if baseAsset is an invalid asset string", func(t *testing.T) {
		_, err := listMarketExternalRequest(invalidAsset, validQuoteAsset)
		assert.NotEqual(t, nil, err)
	})

	t.Run("should return error if quoteAsset is an invalid asset string", func(t *testing.T) {
		_, err := listMarketExternalRequest(validBaseAsset, invalidAsset)
		assert.NotEqual(t, nil, err)
	})

	t.Run("should return error if market is not found for the given quoteAsset", func(t *testing.T) {
		_, err := listMarketExternalRequest(validBaseAsset, validQuoteAssetWithNoMarket)
		assert.NotEqual(t, nil, err)
	})

	t.Run("should return a list of addresses and a nil error if the market argument is valid", func(t *testing.T) {
		addresses, err := listMarketExternalRequest(validBaseAsset, validQuoteAsset)
		assert.Equal(t, nil, err)
		assert.NotEqual(t, nil, addresses)
		assert.Equal(t, 1, len(addresses))
	})
}
