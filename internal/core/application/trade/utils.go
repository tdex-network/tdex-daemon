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
	feeAsset := swapRequest.GetFeeAsset()
	asset, amount := swapRequest.GetAssetP(), swapRequest.GetAmountP()
	preview, _ := tradePreview(
		market, balance, tradeType, feeAsset, asset, amount,
	)

	if preview != nil {
		if isPriceInRange(tradeType, market, swapRequest, preview, slippage) {
			return true
		}
	}

	asset, amount = swapRequest.GetAssetR(), swapRequest.GetAmountR()
	preview, _ = tradePreview(
		market, balance, tradeType, feeAsset, asset, amount,
	)

	if preview == nil {
		return false
	}

	return isPriceInRange(tradeType, market, swapRequest, preview, slippage)
}

func isPriceInRange(
	tradeType ports.TradeType, market domain.Market,
	swapRequest ports.SwapRequest, preview ports.TradePreview,
	slippage decimal.Decimal,
) bool {
	feeAmount := decimal.NewFromInt(int64(swapRequest.GetFeeAmount()))
	feeAsset := swapRequest.GetFeeAsset()
	expectedAmount := decimal.NewFromInt(int64(swapRequest.GetAmountP()))
	expectedAsset := swapRequest.GetAssetP()
	if preview.GetAsset() == swapRequest.GetAssetR() {
		expectedAmount = decimal.NewFromInt(int64(swapRequest.GetAmountR()))
		expectedAsset = swapRequest.GetAssetR()
	}

	amountToCheck := decimal.NewFromInt(int64(preview.GetAmount()))
	if feeAsset == expectedAsset {
		feesToAdd := (tradeType.IsBuy() && feeAsset == market.BaseAsset) ||
			(tradeType.IsBuy() && feeAsset == market.QuoteAsset)
		if feesToAdd {
			expectedAmount.Add(feeAmount)
			amountToCheck.Add(feeAmount)
		} else {
			expectedAmount.Sub(feeAmount)
			amountToCheck.Sub(feeAmount)
		}
	}

	lowerBound := expectedAmount.Mul(decimal.NewFromInt(1).Sub(slippage))
	upperBound := expectedAmount.Mul(decimal.NewFromInt(1).Add(slippage))

	return amountToCheck.GreaterThanOrEqual(lowerBound) &&
		amountToCheck.LessThanOrEqual(upperBound)
}

func tradePreview(
	mkt domain.Market, balance map[string]ports.Balance,
	tradeType ports.TradeType, feeAsset, asset string, amount uint64,
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

	preview, err := mkt.Preview(
		baseBalance, quoteBalance, amount, asset, feeAsset, tradeType.IsBuy(),
	)
	if err != nil {
		return nil, err
	}
	return previewInfo{mkt, *preview}, nil
}
