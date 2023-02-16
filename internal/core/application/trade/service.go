package trade

import (
	"context"
	"fmt"
	"math"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/application/pubsub"
	"github.com/tdex-network/tdex-daemon/internal/core/application/wallet"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/psetv2"
)

var (
	ErrServiceUnavailable = fmt.Errorf("service is unavailable, retry later")
	ErrMarketUnavailable  = fmt.Errorf("market is closed, retry later")

	minSatsPerByte = decimal.NewFromFloat(0.1)
	maxSatsPerByte = decimal.NewFromInt(10000)
)

type Service struct {
	wallet      *wallet.Service
	pubsub      *pubsub.Service
	repoManager ports.RepoManager

	priceSlippage    decimal.Decimal
	milliSatsPerByte uint64
}

func NewService(
	walletSvc *wallet.Service,
	pubsubSvc *pubsub.Service,
	repoManager ports.RepoManager,
	priceSlippage, satsPerByte decimal.Decimal,
) (*Service, error) {
	if walletSvc == nil {
		return nil, fmt.Errorf("missing wallet service")
	}
	if pubsubSvc == nil {
		return nil, fmt.Errorf("missing pubsub service")
	}
	if repoManager == nil {
		return nil, fmt.Errorf("missing repo manager")
	}
	if satsPerByte.LessThan(minSatsPerByte) ||
		satsPerByte.GreaterThan(maxSatsPerByte) {
		return nil, fmt.Errorf(
			"sats per byte ratio must be in range [%s, %s]",
			minSatsPerByte, maxSatsPerByte,
		)
	}
	msatsPerByte := satsPerByte.Mul(decimal.NewFromInt(1000)).BigInt().Uint64()

	svc := &Service{
		walletSvc, pubsubSvc, repoManager, priceSlippage, msatsPerByte,
	}

	go svc.checkForPendingTrades()
	return svc, nil
}

func (s *Service) GetTradableMarkets(ctx context.Context) ([]ports.MarketInfo, error) {
	markets, err := s.repoManager.MarketRepository().GetTradableMarkets(ctx)
	if err != nil {
		return nil, ErrServiceUnavailable
	}
	return marketList(markets).toPortableList(), nil
}

func (s *Service) GetMarketPrice(
	ctx context.Context, market ports.Market,
) (decimal.Decimal, uint64, error) {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return decimal.Zero, 0, ErrServiceUnavailable
	}

	balance, err := s.wallet.Account().GetBalance(ctx, mkt.Name)
	if err != nil {
		return decimal.Zero, 0, ErrServiceUnavailable
	}

	baseAssetBalance := balance[mkt.BaseAsset].GetConfirmedBalance()
	quoteAssetBalance := balance[mkt.QuoteAsset].GetConfirmedBalance()

	price, err := mkt.SpotPrice(baseAssetBalance, quoteAssetBalance)
	if err != nil {
		log.WithError(err).Debug("error while retrieving spot price")
		return decimal.Zero, 0, ErrServiceUnavailable
	}
	spotPrice, _ := decimal.NewFromString(price.QuotePrice)

	// 1 sat of base asset * quote price is the ideal min tradable amount but
	// there are max 8 decimal in blockchain, so if the value is < 1, the min
	// tradable amount is 1 / value.
	// For example, if 1 sat of base asset * quote price = 0.001, the min
	// tradable amount is 1 / 0.001 = 100 sats of base asset (fees excluded).
	minAmount := decimal.NewFromFloat(math.Pow10(-int(mkt.BaseAssetPrecision))).Mul(spotPrice)
	if one := decimal.NewFromFloat(1); minAmount.LessThan(one) {
		minAmount = one.Div(minAmount)
	}
	minTradableAmount := minAmount.BigInt().Uint64()
	if mkt.FixedFee.BaseFee > 0 {
		minTradableAmount += uint64(mkt.FixedFee.BaseFee)
	}

	return spotPrice, minTradableAmount, nil
}

func (s *Service) GetMarketBalance(
	ctx context.Context, market ports.Market,
) (ports.MarketInfo, error) {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return nil, err
	}

	balance, err := s.wallet.Account().GetBalance(ctx, mkt.Name)
	if err != nil {
		return nil, err
	}

	return marketInfo{*mkt, balance}, nil
}

func (s *Service) checkForPendingTrades() {
	ctx := context.Background()
	trades, _ := s.repoManager.TradeRepository().GetAllTrades(ctx, nil)
	expiredTrades := make([]*domain.Trade, 0)
	for _, t := range trades {
		trade := &t
		if trade.IsAccepted() || trade.IsCompleted() {
			if ok, _ := trade.Expire(); ok {
				expiredTrades = append(expiredTrades, trade)
				continue
			}

			pset, _ := psetv2.NewPsetFromBase64(
				trade.SwapAcceptMessage().Transaction,
			)
			tradeUtxos := make([]ports.Utxo, 0, pset.Global.InputCount)
			for _, in := range pset.Inputs {
				prevTxid := elementsutil.TxIDFromBytes(in.PreviousTxid)
				tradeUtxos = append(tradeUtxos, utxo{prevTxid, in.PreviousTxIndex})
			}
			s.makeTradeSettledOrExpired(trade.Id, tradeUtxos)
		}
	}

	for _, t := range expiredTrades {
		if err := s.repoManager.TradeRepository().UpdateTrade(
			ctx, t.Id, func(_ *domain.Trade) (*domain.Trade, error) {
				return t, nil
			},
		); err != nil {
			log.WithError(err).Warnf("failed to expire trade with id %s", t.Id)
			continue
		}
		log.Debugf("expired trade with id %s", t.Id)
	}
}
