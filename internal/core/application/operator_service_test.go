package application

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

const (
	marketRepoIsEmpty  = true
	tradeRepoIsEmpty   = true
	vaultRepoIsEmpty   = true
	unspentRepoIsEmpty = true
	marketPluggable    = true
)

var baseAsset = config.GetString(config.BaseAssetKey)

func TestListMarket(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	marketList, err := operatorService.ListMarket(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(marketList))
}

func TestDepositMarket(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	tests := []struct {
		baseAsset      string
		quoteAsset     string
		numOfAddresses int
	}{
		{
			"",
			"",
			2,
		},
		{
			market.BaseAsset,
			market.QuoteAsset,
			0,
		},
	}

	for _, tt := range tests {
		addresses, err := operatorService.DepositMarket(
			ctx,
			tt.baseAsset,
			tt.quoteAsset,
			tt.numOfAddresses,
		)
		if err != nil {
			t.Fatal(err)
		}
		expectedLen := tt.numOfAddresses
		if tt.numOfAddresses == 0 {
			expectedLen = 1
		}
		assert.Equal(t, expectedLen, len(addresses))
	}
}

func TestFailingDepositMarket(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		marketRepoIsEmpty,
		tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	tests := []struct {
		baseAsset     string
		quoteAsset    string
		expectedError error
	}{
		{
			"",
			market.QuoteAsset,
			domain.ErrMarketInvalidBaseAsset,
		},
		{
			"ldjbwjkbfjksdbjkvcsbdjkbcdsjkb",
			market.QuoteAsset,
			domain.ErrMarketInvalidBaseAsset,
		},
		{
			market.BaseAsset,
			"",
			domain.ErrMarketInvalidQuoteAsset,
		},
		{
			market.BaseAsset,
			"ldjbwjkbfjksdbjkvcsbdjkbcdsjkb",
			domain.ErrMarketInvalidQuoteAsset,
		},
	}

	for _, tt := range tests {
		_, err := operatorService.DepositMarket(ctx, tt.baseAsset, tt.quoteAsset, 0)
		assert.Equal(t, tt.expectedError, err)
	}
}

func TestUpdateMarketPrice(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	args := MarketWithPrice{
		Market: market,
		Price: Price{
			BasePrice:  decimal.NewFromFloat(0.0001234),
			QuotePrice: decimal.NewFromFloat(9876.54321),
		},
	}

	// update the price
	if err := operatorService.UpdateMarketPrice(ctx, args); err != nil {
		t.Fatal(err)
	}
}

func TestFailingUpdateMarketPrice(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	tests := []struct {
		basePrice     float64
		quotePrice    float64
		expectedError error
	}{
		{
			-1,
			10000,
			domain.ErrMarketInvalidBasePrice,
		},
		{
			0,
			10000,
			domain.ErrMarketInvalidBasePrice,
		},
		{
			2099999997690000 + 1,
			10000,
			domain.ErrMarketInvalidBasePrice,
		},
		{
			1,
			-1,
			domain.ErrMarketInvalidQuotePrice,
		},
		{
			1,
			0,
			domain.ErrMarketInvalidQuotePrice,
		},
		{
			1,
			2099999997690000 + 1,
			domain.ErrMarketInvalidQuotePrice,
		},
	}

	for _, tt := range tests {
		args := MarketWithPrice{
			Market: market,
			Price: Price{
				BasePrice:  decimal.NewFromFloat(tt.basePrice),
				QuotePrice: decimal.NewFromFloat(tt.quotePrice),
			},
		}

		err := operatorService.UpdateMarketPrice(ctx, args)
		assert.Equal(t, tt.expectedError, err)
	}
}
func TestListSwap(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		!tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	swapInfos, err := operatorService.ListSwaps(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(swapInfos))
}

func TestBalanceFeeAccount(t *testing.T) {
	operatorService, _, _, ctx, close, _ := newMockServices(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
		!unspentRepoIsEmpty,
		false,
	)
	t.Cleanup(close)

	balance, err := operatorService.FeeAccountBalance(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(100000000), balance)
}

func TestGetCollectedMarketFee(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		!tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	fee, err := operatorService.GetCollectedMarketFee(ctx, market)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, len(fee.CollectedFees))
}

func TestListMarketExternalAddresses(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	addresses, err := operatorService.ListMarketExternalAddresses(ctx, market)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(addresses))
}

func TestFailingListMarketExternalAddresses(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	tests := []struct {
		market        Market
		expectedError error
	}{
		{
			Market{
				"",
				marketUnspents[1].AssetHash,
			},
			domain.ErrMarketInvalidBaseAsset,
		},
		{
			Market{
				marketUnspents[0].AssetHash,
				"",
			},
			domain.ErrMarketInvalidQuoteAsset,
		},
		{
			Market{
				marketUnspents[0].AssetHash,
				marketUnspents[0].AssetHash,
			},
			ErrMarketNotExist,
		},
	}

	for _, tt := range tests {
		_, err := operatorService.ListMarketExternalAddresses(ctx, tt.market)
		assert.Equal(t, tt.expectedError, err)
	}
}

func TestOpenMarket(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	err := operatorService.OpenMarket(ctx, market.BaseAsset, market.QuoteAsset)
	assert.Equal(t, ErrFeeAccountNotFunded, err)

	if _, err := operatorService.DepositFeeAccount(ctx, 1); err != nil {
		t.Fatal(err)
	}

	if err := operatorService.OpenMarket(
		ctx,
		market.BaseAsset,
		market.QuoteAsset,
	); err != nil {
		t.Fatal(err)
	}
}

func TestFailingOpenMarket(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	tests := []struct {
		baseAsset     string
		quoteAsset    string
		expectedError error
	}{
		{
			"",
			market.QuoteAsset,
			domain.ErrMarketInvalidBaseAsset,
		},
		{
			"invalidasset",
			market.QuoteAsset,
			domain.ErrMarketInvalidBaseAsset,
		},
		{
			market.BaseAsset,
			"",
			domain.ErrMarketInvalidQuoteAsset,
		},
		{
			market.BaseAsset,
			"invalidasset",
			domain.ErrMarketInvalidQuoteAsset,
		},
		{
			market.BaseAsset,
			"0ddfa690c7b2ba3b8ecee8200da2420fc502f57f8312c83d466b6f8dced70441",
			ErrMarketNotExist,
		},
	}

	for _, tt := range tests {
		err := operatorService.OpenMarket(ctx, tt.baseAsset, tt.quoteAsset)
		assert.Equal(t, tt.expectedError, err)
	}
}

func TestUpdateMarketStrategy(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	err := operatorService.UpdateMarketStrategy(
		ctx,
		MarketStrategy{
			Market:   market,
			Strategy: domain.StrategyTypePluggable,
		},
	)
	assert.Equal(t, domain.ErrMarketMustBeClosed, err)

	if err := operatorService.CloseMarket(
		ctx,
		market.BaseAsset,
		market.QuoteAsset,
	); err != nil {
		t.Fatal(err)
	}

	if err := operatorService.UpdateMarketStrategy(
		ctx,
		MarketStrategy{
			Market:   market,
			Strategy: domain.StrategyTypePluggable,
		},
	); err != nil {
		t.Fatal(err)
	}

	if err := operatorService.UpdateMarketStrategy(
		ctx,
		MarketStrategy{
			Market:   market,
			Strategy: domain.StrategyTypeBalanced,
		},
	); err != nil {
		t.Fatal(err)
	}
}

func TestFailingUpdateMarketStratergy(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)
	t.Cleanup(close)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	tests := []struct {
		baseAsset     string
		quoteAsset    string
		strategyType  domain.StrategyType
		expectedError error
	}{
		{
			"",
			market.QuoteAsset,
			domain.StrategyTypePluggable,
			domain.ErrMarketInvalidBaseAsset,
		},
		{
			"invalidasset",
			market.QuoteAsset,
			domain.StrategyTypePluggable,
			domain.ErrMarketInvalidBaseAsset,
		},
		{
			market.QuoteAsset,
			"",
			domain.StrategyTypePluggable,
			domain.ErrMarketInvalidQuoteAsset,
		},
		{
			market.QuoteAsset,
			"invalidasset",
			domain.StrategyTypePluggable,
			domain.ErrMarketInvalidQuoteAsset,
		},
		{
			market.QuoteAsset,
			"0ddfa690c7b2ba3b8ecee8200da2420fc502f57f8312c83d466b6f8dced8a441",
			domain.StrategyTypePluggable,
			ErrMarketNotExist,
		},
		{
			market.BaseAsset,
			market.QuoteAsset,
			domain.StrategyTypeUnbalanced,
			ErrUnknownStrategy,
		},
	}

	for _, tt := range tests {
		err := operatorService.UpdateMarketStrategy(
			ctx,
			MarketStrategy{
				Market: Market{
					tt.baseAsset,
					tt.quoteAsset,
				},
				Strategy: tt.strategyType,
			},
		)
		assert.Equal(t, tt.expectedError, err)
	}
}
