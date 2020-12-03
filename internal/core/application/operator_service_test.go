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

	marketList, err := operatorService.ListMarket(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(marketList))

	t.Cleanup(close)
}

func TestDepositMarket(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)

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

	t.Cleanup(close)
}

func TestFailingDepositMarket(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		marketRepoIsEmpty,
		tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)

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
			domain.ErrInvalidBaseAsset,
		},
		{
			"ldjbwjkbfjksdbjkvcsbdjkbcdsjkb",
			market.QuoteAsset,
			domain.ErrInvalidBaseAsset,
		},
		{
			market.BaseAsset,
			"",
			domain.ErrInvalidQuoteAsset,
		},
		{
			market.BaseAsset,
			"ldjbwjkbfjksdbjkvcsbdjkbcdsjkb",
			domain.ErrInvalidQuoteAsset,
		},
	}

	for _, tt := range tests {
		_, err := operatorService.DepositMarket(ctx, tt.baseAsset, tt.quoteAsset, 0)
		assert.Equal(t, tt.expectedError, err)
	}

	t.Cleanup(close)
}

func TestUpdateMarketPrice(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)

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

	t.Cleanup(close)
}

func TestFailingUpdateMarketPrice(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)

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
			domain.ErrInvalidBasePrice,
		},
		{
			0,
			10000,
			domain.ErrInvalidBasePrice,
		},
		{
			2099999997690000 + 1,
			10000,
			domain.ErrInvalidBasePrice,
		},
		{
			1,
			-1,
			domain.ErrInvalidQuotePrice,
		},
		{
			1,
			0,
			domain.ErrInvalidQuotePrice,
		},
		{
			1,
			2099999997690000 + 1,
			domain.ErrInvalidQuotePrice,
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

	t.Cleanup(close)
}
func TestListSwap(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		!tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)

	swapInfos, err := operatorService.ListSwaps(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(swapInfos))
	t.Cleanup(close)
}

func TestWithdrawMarket(t *testing.T) {
	operatorService, _, _, ctx, close, _ := newMockServices(
		!marketRepoIsEmpty,
		!tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
		!unspentRepoIsEmpty,
		false,
	)

	rawTx, err := operatorService.WithdrawMarketFunds(ctx, WithdrawMarketReq{
		Market: Market{
			BaseAsset:  marketUnspents[0].AssetHash,
			QuoteAsset: marketUnspents[1].AssetHash,
		},
		BalanceToWithdraw: Balance{
			BaseAmount:  4200,
			QuoteAmount: 2300,
		},
		MillisatPerByte: 20,
		Address:         "el1qq22f83p6asdy7jsp4tuke0d9emvxhcenqee5umsn88fsn8gggzlrx0md4hp38rnwcnu9lusmzhmktlt3h5q0gecfpfvx6uac2",
		Push:            false,
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, len(rawTx) > 0)

	t.Cleanup(close)
}

func TestFailingWithdrawMarket(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		!tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)

	tests := []struct {
		args          WithdrawMarketReq
		expectedError error
	}{
		{
			args: WithdrawMarketReq{
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
			},
			expectedError: domain.ErrInvalidBaseAsset,
		},
		{
			args: WithdrawMarketReq{
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
			},
			expectedError: domain.ErrMarketNotExist,
		},
		{
			args: WithdrawMarketReq{
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
			},
			expectedError: domain.ErrMarketNotExist,
		},
	}

	for _, tt := range tests {
		_, err := operatorService.WithdrawMarketFunds(ctx, tt.args)
		assert.Equal(t, tt.expectedError, err)
	}

	t.Cleanup(close)
}

func TestBalanceFeeAccount(t *testing.T) {
	operatorService, _, _, ctx, close, _ := newMockServices(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
		!unspentRepoIsEmpty,
		false,
	)

	balance, err := operatorService.FeeAccountBalance(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(100000000), balance)

	t.Cleanup(close)
}

func TestGetCollectedMarketFee(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		!tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	fee, err := operatorService.GetCollectedMarketFee(ctx, market)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, len(fee.CollectedFees))

	t.Cleanup(close)
}

func TestListMarketExternalAddresses(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	addresses, err := operatorService.ListMarketExternalAddresses(ctx, market)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(addresses))

	t.Cleanup(close)
}

func TestFailingListMarketExternalAddresses(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)

	tests := []struct {
		market        Market
		expectedError error
	}{
		{
			Market{
				"",
				marketUnspents[1].AssetHash,
			},
			domain.ErrInvalidBaseAsset,
		},
		{
			Market{
				marketUnspents[0].AssetHash,
				"",
			},
			domain.ErrInvalidQuoteAsset,
		},
		{
			Market{
				marketUnspents[0].AssetHash,
				marketUnspents[0].AssetHash,
			},
			domain.ErrMarketNotExist,
		},
	}

	for _, tt := range tests {
		_, err := operatorService.ListMarketExternalAddresses(ctx, tt.market)
		assert.Equal(t, tt.expectedError, err)
	}

	t.Cleanup(close)
}

func TestOpenMarket(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		vaultRepoIsEmpty,
	)

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

	t.Cleanup(close)
}

func TestFailingOpenMarket(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)

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
			domain.ErrInvalidBaseAsset,
		},
		{
			"invalidasset",
			market.QuoteAsset,
			domain.ErrInvalidBaseAsset,
		},
		{
			market.BaseAsset,
			"",
			domain.ErrInvalidQuoteAsset,
		},
		{
			market.BaseAsset,
			"invalidasset",
			domain.ErrInvalidQuoteAsset,
		},
		{
			market.BaseAsset,
			"0ddfa690c7b2ba3b8ecee8200da2420fc502f57f8312c83d466b6f8dced70441",
			domain.ErrMarketNotExist,
		},
	}

	for _, tt := range tests {
		err := operatorService.OpenMarket(ctx, tt.baseAsset, tt.quoteAsset)
		assert.Equal(t, tt.expectedError, err)
	}

	t.Cleanup(close)
}

func TestUpdateMarketStrategy(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)

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
	assert.Equal(t, domain.ErrMarketMustBeClose, err)

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

	t.Cleanup(close)
}

func TestFailingUpdateMarketStratergy(t *testing.T) {
	operatorService, ctx, close := newTestOperator(
		!marketRepoIsEmpty,
		tradeRepoIsEmpty,
		!vaultRepoIsEmpty,
	)

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
			domain.ErrInvalidBaseAsset,
		},
		{
			"invalidasset",
			market.QuoteAsset,
			domain.StrategyTypePluggable,
			domain.ErrInvalidBaseAsset,
		},
		{
			market.QuoteAsset,
			"",
			domain.StrategyTypePluggable,
			domain.ErrInvalidQuoteAsset,
		},
		{
			market.QuoteAsset,
			"invalidasset",
			domain.StrategyTypePluggable,
			domain.ErrInvalidQuoteAsset,
		},
		{
			market.QuoteAsset,
			"0ddfa690c7b2ba3b8ecee8200da2420fc502f57f8312c83d466b6f8dced8a441",
			domain.StrategyTypePluggable,
			domain.ErrMarketNotExist,
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

	t.Cleanup(close)
}
