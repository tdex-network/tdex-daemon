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
	slippage decimal.Decimal, feesToAdd bool,
) bool {
	feeAsset := swapRequest.GetFeeAsset()
	amount, asset := swapRequest.GetAmountP(), swapRequest.GetAssetP()
	otherAmount, otherAsset := swapRequest.GetAmountR(), swapRequest.GetAssetR()
	if feeAsset == swapRequest.GetAssetR() {
		amount, asset = swapRequest.GetAmountR(), swapRequest.GetAssetR()
		otherAmount, otherAsset = swapRequest.GetAmountP(), swapRequest.GetAssetP()
	}

	// Charge fees to the amount of the swap request message.
	if feesToAdd {
		amount += swapRequest.GetFeeAmount()
	} else {
		amount -= swapRequest.GetFeeAmount()
	}

	preview, _ := tradePreview(
		market, balance, tradeType, feeAsset, asset, amount,
	)

	if preview != nil {
		if isPriceInRange(tradeType, swapRequest, preview, slippage, feesToAdd) {
			return true
		}
	}

	preview, _ = tradePreview(
		market, balance, tradeType, feeAsset, otherAsset, otherAmount,
	)

	if preview == nil {
		return false
	}

	return isPriceInRange(tradeType, swapRequest, preview, slippage, feesToAdd)
}

func isPriceInRange(
	tradeType ports.TradeType, swapRequest ports.SwapRequest,
	preview ports.TradePreview, slippage decimal.Decimal, feesToAdd bool,
) bool {
	amount, asset := swapRequest.GetAmountP(), swapRequest.GetAssetP()
	if preview.GetAsset() == swapRequest.GetAssetP() {
		amount, asset = swapRequest.GetAmountR(), swapRequest.GetAssetR()
	}

	amountToCheck := decimal.NewFromInt(int64(amount))
	feeAmount := decimal.NewFromInt(int64(preview.GetFeeAmount()))
	if preview.GetFeeAsset() != asset {
		amountToCheck = decimal.NewFromInt(int64(preview.GetAmount()))
	}
	if feesToAdd {
		amountToCheck.Add(feeAmount)
	} else {
		amountToCheck.Sub(feeAmount)
	}

	expectedAmount := decimal.NewFromInt(int64(amount))
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
