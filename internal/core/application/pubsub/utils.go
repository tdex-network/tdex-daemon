package pubsub

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

func topicForEvent(event ports.WebhookEvent) string {
	switch {
	case event.IsTradeSettled():
		return eventTradeSettled
	case event.IsAccountLowBalance():
		return eventAccountLowBalance
	case event.IsAccountWithdraw():
		return eventAccountWithdraw
	case event.IsAccountDeposit():
		return eventAccountDeposit
	case event.IsAny():
		return ports.AnyTopic
	case event.IsUnspecified():
		fallthrough
	default:
		return ports.UnspecifiedTopic
	}
}

func getAccountPayload(
	accountName string, market ports.Market,
) map[string]interface{} {
	account := make(map[string]interface{})
	if accountName == domain.FeeAccount {
		account["type"] = "fee"
		account["name"] = domain.FeeAccount
	} else {
		account["type"] = "market"
		account["base_asset"] = market.GetBaseAsset()
		account["quote_asset"] = market.GetQuoteAsset()
	}
	return account
}

func getBalancePayload(
	accountBalance map[string]ports.Balance,
) map[string]interface{} {
	balance := make(map[string]interface{})
	for asset, bal := range accountBalance {
		var unconfBalance, confBalance, lockBalance uint64
		if bal != nil {
			unconfBalance = bal.GetUnconfirmedBalance()
			confBalance = bal.GetConfirmedBalance()
			lockBalance = bal.GetLockedBalance()
		}
		balance[asset] = map[string]uint64{
			"unconfirmed": unconfBalance,
			"confirmed":   confBalance,
			"locked":      lockBalance,
		}
	}
	return balance
}

func getSwapPayload(sr *domain.SwapRequest) map[string]interface{} {
	return map[string]interface{}{
		"amount_p": sr.GetAmountP(),
		"asset_p":  sr.GetAssetP(),
		"amount_r": sr.GetAmountR(),
		"asset_r":  sr.GetAssetR(),
	}
}

func getMarketPayload(trade domain.Trade) map[string]interface{} {
	return map[string]interface{}{
		"base_asset":  trade.MarketBaseAsset,
		"quote_asset": trade.MarketQuoteAsset,
	}
}
