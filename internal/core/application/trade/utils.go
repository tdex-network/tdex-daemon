package trade

import (
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

// isValidPrice checks that the amounts of the trade are valid by
// making a preview of each counter amounts of the swap given the
// current price of the market.
// Since the price is variable in time, the predicted amounts are not compared
// against those of the swap, but rather they are used to create a range in
// which the swap amounts must be included to be considered valid.
func isValidTradePrice(
	market domain.Market, balance map[string]ports.Balance,
	tradeType ports.TradeType, swapRequest ports.SwapRequest,
	slippage decimal.Decimal,
) bool {
	amount := swapRequest.GetAmountR()
	if tradeType.IsSell() {
		amount = swapRequest.GetAmountP()
	}

	preview, _ := tradePreview(
		market, balance,
		tradeType, market.BaseAsset, amount,
	)

	if preview != nil {
		if isPriceInRange(
			tradeType, swapRequest, preview.GetAmount(), true, slippage,
		) {
			return true
		}
	}

	amount = swapRequest.GetAmountP()
	if tradeType.IsSell() {
		amount = swapRequest.GetAmountR()
	}

	preview, _ = tradePreview(
		market, balance,
		tradeType, market.QuoteAsset, amount,
	)

	if preview == nil {
		return false
	}

	return isPriceInRange(
		tradeType, swapRequest, preview.GetAmount(), false, slippage,
	)
}

func isPriceInRange(
	tradeType ports.TradeType, swapRequest ports.SwapRequest,
	previewAmount uint64, isPreviewForQuoteAsset bool,
	slippage decimal.Decimal,
) bool {
	amountToCheck := decimal.NewFromInt(int64(swapRequest.GetAmountP()))
	if tradeType.IsSell() {
		if isPreviewForQuoteAsset {
			amountToCheck = decimal.NewFromInt(int64(swapRequest.GetAmountR()))
		}
	} else {
		if !isPreviewForQuoteAsset {
			amountToCheck = decimal.NewFromInt(int64(swapRequest.GetAmountR()))
		}
	}

	expectedAmount := decimal.NewFromInt(int64(previewAmount))
	lowerBound := expectedAmount.Mul(decimal.NewFromInt(1).Sub(slippage))
	upperBound := expectedAmount.Mul(decimal.NewFromInt(1).Add(slippage))

	return amountToCheck.GreaterThanOrEqual(lowerBound) && amountToCheck.LessThanOrEqual(upperBound)
}

func tradePreview(
	mkt domain.Market, balance map[string]ports.Balance,
	tradeType ports.TradeType, asset string, amount uint64,
) (ports.TradePreview, error) {
	var baseBalance, quoteBalance uint64
	if balance != nil {
		if b, ok := balance[mkt.BaseAsset]; ok {
			baseBalance = b.GetConfirmedBalance()
		}
		if b, ok := balance[mkt.QuoteAsset]; ok {
			quoteBalance = b.GetConfirmedBalance()
		}
	}
	isBaseAsset := asset == mkt.BaseAsset

	preview, err := mkt.Preview(
		baseBalance, quoteBalance, amount, isBaseAsset, tradeType.IsBuy(),
	)
	if err != nil {
		return nil, err
	}
	return previewInfo{mkt, *preview, balance}, nil
}
