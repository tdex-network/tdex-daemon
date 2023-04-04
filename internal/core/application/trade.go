package application

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/application/pubsub"
	"github.com/tdex-network/tdex-daemon/internal/core/application/trade"
	"github.com/tdex-network/tdex-daemon/internal/core/application/wallet"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type TradeService interface {
	GetTradableMarkets(ctx context.Context) ([]ports.MarketInfo, error)
	GetMarketPrice(
		ctx context.Context, market ports.Market,
	) (decimal.Decimal, uint64, error)
	TradePreview(
		ctx context.Context, market ports.Market,
		tradeType ports.TradeType, amount uint64, asset, feeAsset string,
	) (ports.TradePreview, error)
	TradePropose(
		ctx context.Context, market ports.Market,
		tradeType ports.TradeType, swapRequest ports.SwapRequest,
	) (ports.SwapAccept, ports.SwapFail, int64, error)
	TradeComplete(
		ctx context.Context,
		swapComplete ports.SwapComplete, swapFail ports.SwapFail,
	) (string, ports.SwapFail, error)
	GetMarketBalance(
		ctx context.Context, market ports.Market,
	) (ports.MarketInfo, error)
}

func NewTradeService(
	walletSvc WalletService, pubsubSvc PubSubService,
	repoManager ports.RepoManager,
	priceSlippage, satsPerByte decimal.Decimal,
) (TradeService, error) {
	w := walletSvc.(*wallet.Service)
	p := pubsubSvc.(*pubsub.Service)
	return trade.NewService(w, p, repoManager, priceSlippage, satsPerByte)
}
