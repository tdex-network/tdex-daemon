package application

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

const marketRepoIsEmpty = true

var baseAsset = config.GetString(config.BaseAssetKey)

func TestListMarket(t *testing.T) {
	t.Run("ListMarket should return an empty list and a nil error if market repository is empty", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty)
		marketInfos, err := operatorService.ListMarket(ctx)
		close()
		assert.Equal(t, nil, err)
		assert.Equal(t, 0, len(marketInfos))
	})

	t.Run("ListMarket should return the number of markets in the market repository", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(!marketRepoIsEmpty)
		marketInfos, err := operatorService.ListMarket(ctx)
		close()
		assert.Equal(t, nil, err)
		assert.Equal(t, 2, len(marketInfos))
	})

}

func TestDepositMarket(t *testing.T) {

	t.Run("DepositMarket with new market", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty)

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
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty)

		emptyAddress, err := operatorService.DepositMarket(ctx, "", "validQuoteAsset")
		assert.Equal(t, domain.ErrInvalidBaseAsset, err)
		assert.Equal(
			t,
			"",
			emptyAddress,
		)

		close()
	})

	t.Run("DepositMarket with valid base asset and empty quote asset", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty)

		emptyAddress, err := operatorService.DepositMarket(ctx, baseAsset, "")
		assert.Equal(t, domain.ErrMarketNotExist, err)
		assert.Equal(
			t,
			"",
			emptyAddress,
		)

		close()
	})

	t.Run("DepositMarket with valid base asset and invalid quote asset", func(t *testing.T) {
		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty)

		emptyAddress, err := operatorService.DepositMarket(ctx, baseAsset, "ldjbwjkbfjksdbjkvcsbdjkbcdsjkb")
		assert.Equal(t, domain.ErrMarketNotExist, err)
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

		operatorService, ctx, close := newTestOperator(marketRepoIsEmpty)

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
