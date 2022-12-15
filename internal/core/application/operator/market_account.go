package operator

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/tdex-network/tdex-daemon/pkg/mathutil"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

func (s *service) NewMarket(
	ctx context.Context, market ports.Market,
) (ports.MarketInfo, error) {
	mkt, _ := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if mkt != nil {
		return nil, fmt.Errorf("market already exists")
	}

	newMarket, err := domain.NewMarket(
		market.GetBaseAsset(), market.GetQuoteAsset(), s.marketPercentageFee,
	)
	if err != nil {
		return nil, err
	}

	if _, err := s.wallet.Account().CreateAccount(
		ctx, newMarket.Name,
	); err != nil {
		return nil, err
	}

	if err := s.repoManager.MarketRepository().AddMarket(ctx, newMarket); err != nil {
		go func() {
			if err := s.wallet.Account().DeleteAccount(ctx, newMarket.Name); err != nil {
				log.WithError(err).Warn("failed to delete wallet account, please do it manually")
			}
		}()
		return nil, err
	}

	return marketInfo{*newMarket, nil}, nil
}

func (s *service) GetMarketInfo(
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
		log.WithError(err).Warnf("failed to fetch balance for market %s", mkt.Name)
	}

	return marketInfo{*mkt, balance}, nil
}

func (s *service) DeriveMarketAddresses(
	ctx context.Context, market ports.Market, num int,
) ([]string, error) {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return nil, err
	}

	return s.wallet.Account().DeriveAddresses(ctx, mkt.Name, num)
}

func (s *service) ListMarketExternalAddresses(
	ctx context.Context, market ports.Market,
) ([]string, error) {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return nil, err
	}

	return s.wallet.Account().ListAddresses(ctx, mkt.Name)
}

func (s *service) GetMarketReport(
	ctx context.Context, market Market,
	timeRange TimeRange, groupByHours int,
) (*MarketReport, error) {
	startTime, endTime, err := timeRange.getStartAndEndTime(time.Now())
	if err != nil {
		return nil, err
	}

	if int(endTime.Sub(startTime).Hours()) <= groupByHours {
		return nil, ErrInvalidTimeFrame
	}

	m, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.BaseAsset, market.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, ErrMarketNotExist
	}

	trades, err := s.repoManager.TradeRepository().GetCompletedTradesByMarket(
		ctx, market.QuoteAsset, nil,
	)
	if err != nil {
		return nil, err
	}

	//sort desc
	sort.SliceStable(trades, func(i, j int) bool {
		return trades[i].SwapRequest.Timestamp > trades[j].SwapRequest.Timestamp
	})

	groupedVolume := initGroupedVolume(startTime, endTime, groupByHours)

	totalFees := make(map[string]int64)
	volume := make(map[string]int64)
	tradeFeeInfo := make([]TradeFeeInfo, 0, len(trades))
	for _, trade := range trades {
		if (time.Unix(trade.SwapRequest.Timestamp, 0).After(startTime) ||
			time.Unix(trade.SwapRequest.Timestamp, 0).Equal(startTime)) &&
			(time.Unix(trade.SwapRequest.Timestamp, 0).Before(endTime) ||
				time.Unix(trade.SwapRequest.Timestamp, 0).Equal(endTime)) {
			feeBasisPoint := trade.MarketPercentageFee
			swapRequest := trade.SwapRequestMessage()
			feeAsset := swapRequest.GetAssetP()
			amountP := swapRequest.GetAmountP()
			_, percentageFeeAmount := mathutil.LessFee(amountP, uint64(feeBasisPoint))

			marketPrice := trade.MarketPrice.QuotePrice
			fixedFeeAmount := trade.MarketFixedQuoteFee
			if feeAsset == m.BaseAsset {
				marketPrice = trade.MarketPrice.BasePrice
				fixedFeeAmount = trade.MarketFixedBaseFee
			}
			totalFees[feeAsset] += int64(percentageFeeAmount) + int64(fixedFeeAmount)

			volume[swapRequest.GetAssetR()] += int64(swapRequest.GetAmountR())
			volume[swapRequest.GetAssetP()] += int64(swapRequest.GetAmountP())

			for i, v := range groupedVolume {
				//find time slot to which trade belongs to and calculate volume for that slot
				if (time.Unix(trade.SwapRequest.Timestamp, 0).After(v.StartTime) ||
					time.Unix(trade.SwapRequest.Timestamp, 0).Equal(v.StartTime)) &&
					(time.Unix(trade.SwapRequest.Timestamp, 0).Before(v.EndTime) ||
						time.Unix(trade.SwapRequest.Timestamp, 0).Equal(v.EndTime)) {

					//assume AmountR is base asset, AmountP(FeeAsset) is quote asset
					volumeBaseAmount := swapRequest.GetAmountR() + v.BaseVolume
					volumeQuoteAmount := swapRequest.GetAmountP() + v.QuoteVolume
					if swapRequest.GetAssetR() == market.QuoteAsset {
						volumeBaseAmount = swapRequest.GetAmountP() + v.BaseVolume
						volumeQuoteAmount = swapRequest.GetAmountR() + v.QuoteVolume
					}

					groupedVolume[i] = MarketVolume{
						BaseVolume:  volumeBaseAmount,
						QuoteVolume: volumeQuoteAmount,
						StartTime:   v.StartTime,
						EndTime:     v.EndTime,
					}
				}
			}

			mp, err := decimal.NewFromString(marketPrice)
			if err != nil {
				return nil, err
			}

			tradeFeeInfo = append(tradeFeeInfo, TradeFeeInfo{
				TradeID:             trade.Id,
				PercentageFee:       uint64(feeBasisPoint),
				FeeAsset:            feeAsset,
				PercentageFeeAmount: percentageFeeAmount,
				FixedFeeAmount:      fixedFeeAmount,
				MarketPrice:         mp,
			})
		}
	}

	return &MarketReport{
		Market: Market{
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
		},
		CollectedFees: MarketCollectedFees{
			BaseAmount:   uint64(totalFees[market.BaseAsset]),
			QuoteAmount:  uint64(totalFees[market.QuoteAsset]),
			TradeFeeInfo: tradeFeeInfo,
			StartTime:    startTime,
			EndTime:      endTime,
		},
		TotalVolume: MarketVolume{
			BaseVolume:  uint64(volume[market.BaseAsset]),
			QuoteVolume: uint64(volume[market.QuoteAsset]),
			StartTime:   startTime,
			EndTime:     endTime,
		},
		VolumesPerFrame: groupedVolume,
	}, nil
}

// initGroupedVolume splits the given time range (start, end) into a list of
// MarketVolume, ie. smaller consecutive time ranges of numHours hours in descending order.
// Example:
// in: 2009-11-10 19:00:00 (start), 2009-11-11 00:00:00 (end), 2 (numHours)
// out: [
//
//	{end: 2009-11-11 00:00:00, start: 2009-11-10 22:00:01},
//	{end: 2009-11-11 22:00:00, start: 2009-11-10 20:00:01},
//	{end: 2009-11-10 20:00:00, start: 2009-11-10 19:00:00},
//
// ]
func initGroupedVolume(start, end time.Time, groupByHours int) []MarketVolume {
	groupedVolume := make([]MarketVolume, 0)
	for {
		if end.Equal(start) || end.Before(start) {
			return groupedVolume
		} else {
			nextEnd := end.Add(-time.Hour * time.Duration(groupByHours))
			nextStart := start
			if nextEnd.Sub(start).Seconds() > 0 {
				nextStart = nextEnd.Add(time.Second)
			}
			groupedVolume = append(groupedVolume, MarketVolume{
				StartTime: nextStart,
				EndTime:   end,
			})
			end = nextEnd
		}
	}
}

func (s *service) OpenMarket(ctx context.Context, market ports.Market) error {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return err
	}

	return s.repoManager.MarketRepository().OpenMarket(ctx, mkt.Name)
}

func (s *service) CloseMarket(ctx context.Context, market ports.Market) error {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return err
	}

	return s.repoManager.MarketRepository().CloseMarket(ctx, mkt.Name)
}

func (s *service) DropMarket(ctx context.Context, market ports.Market) error {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return err
	}

	if err := s.wallet.Account().DeleteAccount(ctx, mkt.Name); err != nil {
		return err
	}

	return s.repoManager.MarketRepository().DeleteMarket(ctx, mkt.Name)
}

func (s *service) WithdrawMarketFunds(
	ctx context.Context,
	market ports.Market, outputs []ports.TxOutput, millisatsPerByte uint64,
) (string, error) {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return "", err
	}

	return s.wallet.SendToMany(mkt.Name, outputs, millisatsPerByte)
}

func (s *service) UpdateMarketPercentageFee(
	ctx context.Context, market ports.Market, basisPoint uint32,
) (ports.MarketInfo, error) {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return nil, err
	}

	if err := mkt.ChangePercentageFee(basisPoint); err != nil {
		return nil, err
	}

	if err := s.repoManager.MarketRepository().UpdateMarket(
		ctx, mkt.Name, func(_ *domain.Market) (*domain.Market, error) {
			return mkt, nil
		},
	); err != nil {
		return nil, err
	}

	return marketInfo{*mkt, nil}, nil
}

func (s *service) UpdateMarketFixedFee(
	ctx context.Context,
	market ports.Market, baseFixedFee, quoteFixedFee int64,
) (ports.MarketInfo, error) {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return nil, err
	}

	if err := mkt.ChangeFixedFee(baseFixedFee, quoteFixedFee); err != nil {
		return nil, err
	}

	if err := s.repoManager.MarketRepository().UpdateMarket(
		ctx, mkt.Name, func(_ *domain.Market) (*domain.Market, error) {
			return mkt, nil
		},
	); err != nil {
		return nil, err
	}

	return marketInfo{*mkt, nil}, nil
}

func (s *service) UpdateMarketPrice(
	ctx context.Context,
	market ports.Market, basePrice, quotePrice decimal.Decimal,
) error {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return err
	}

	return s.repoManager.MarketRepository().UpdateMarketPrice(
		ctx, mkt.Name, domain.MarketPrice{
			BasePrice:  basePrice.String(),
			QuotePrice: quotePrice.String(),
		},
	)
}

func (s *service) UpdateMarketStrategy(
	ctx context.Context, market ports.Market, strategyType int,
) error {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return err
	}

	switch strategyType {
	case domain.StrategyTypeBalanced:
		if err := mkt.MakeStrategyBalanced(); err != nil {
			return err
		}
	case domain.StrategyTypePluggable:
		if err := mkt.MakeStrategyPluggable(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown strategy type")
	}

	return s.repoManager.MarketRepository().UpdateMarket(
		ctx, mkt.Name, func(_ *domain.Market) (*domain.Market, error) {
			return mkt, nil
		},
	)
}
