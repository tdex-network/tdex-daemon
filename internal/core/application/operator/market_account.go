package operator

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

const startYear = 2021

func (s *service) NewMarket(
	ctx context.Context,
	market ports.Market, marketName string,
	basePercentageFee, quotePercentageFee, baseFixedFee, quoteFixedFee uint64,
	baseAssetPrecision, quoteAssetPrecision, strategyType uint,
) (ports.MarketInfo, error) {
	mkt, _ := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if mkt != nil {
		return nil, fmt.Errorf("market already exists")
	}

	newMarket, err := domain.NewMarket(
		market.GetBaseAsset(), market.GetQuoteAsset(), marketName,
		basePercentageFee, quotePercentageFee, baseFixedFee, quoteFixedFee,
		baseAssetPrecision, quoteAssetPrecision, strategyType,
	)
	if err != nil {
		return nil, err
	}

	accountInfo, err := s.wallet.Account().CreateAccount(
		ctx, newMarket.Name, true,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repoManager.MarketRepository().AddMarket(
		ctx, newMarket,
	); err != nil {
		go func() {
			if err := s.wallet.Account().DeleteAccount(
				ctx, newMarket.Name,
			); err != nil {
				log.WithError(err).Warn(
					"failed to delete wallet account, please do it manually",
				)
			}
		}()
		return nil, err
	}

	s.accounts.add(accountInfo.GetNamespace(), accountInfo.GetLabel())

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
	ctx context.Context, market ports.Market,
	timeRange ports.TimeRange, timeFrame int,
) (ports.MarketReport, error) {
	rangeStart, rangeEnd, err := timeRangeToDates(timeRange)
	if err != nil {
		return nil, err
	}

	if timeFrame == 0 {
		timeFrame = getDefaultTimeFrameForRange(rangeStart, rangeEnd)
	}

	if int(rangeEnd.Sub(rangeStart).Hours()) < timeFrame {
		return nil, fmt.Errorf("time range must be larger than time frame")
	}

	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return nil, err
	}

	trades, err := s.repoManager.TradeRepository().GetCompletedTradesByMarket(
		ctx, mkt.Name, nil,
	)
	if err != nil {
		return nil, err
	}

	// Sort trades by swap request timestamp, desc order.
	sort.SliceStable(trades, func(i, j int) bool {
		return trades[i].SwapRequest.Timestamp > trades[j].SwapRequest.Timestamp
	})

	subVolumes := splitTimeRange(rangeStart, rangeEnd, timeFrame)
	tradesFee := make([]tradeFeeInfo, 0)
	for _, trade := range trades {
		if isInTimeRange(trade.SwapRequest.Timestamp, rangeStart, rangeEnd) {
			swapRequest := trade.SwapRequestMessage()
			marketPrice := trade.MarketPrice.BasePrice
			if trade.FeeAsset == mkt.BaseAsset {
				marketPrice = trade.MarketPrice.QuotePrice
			}

			tradesFee = append(tradesFee, tradeFeeInfo{trade, marketPrice})

			for i, v := range subVolumes {
				if isInTimeRange(trade.SwapRequest.Timestamp, v.start, v.end) {
					baseAmount := swapRequest.GetAmountR()
					quoteAmount := swapRequest.GetAmountP()
					if swapRequest.GetAssetR() == mkt.QuoteAsset {
						baseAmount, quoteAmount = quoteAmount, baseAmount
					}

					subVolumes[i].baseVolume += baseAmount
					subVolumes[i].quoteVolume += quoteAmount
				}
			}
		}
	}

	var totBaseVolume, totQuoteVolume uint64
	for _, v := range subVolumes {
		totBaseVolume += v.baseVolume
		totQuoteVolume += v.quoteVolume
	}
	totVolume := marketVolumeInfo{
		rangeStart, rangeEnd, totBaseVolume, totQuoteVolume,
	}

	var totBaseFee, totQuoteFee uint64
	for _, v := range tradesFee {
		if v.GetFeeAsset() == v.Trade.MarketBaseAsset {
			totBaseFee += v.GetFeeAmount()
		} else {
			totQuoteFee += v.GetFeeAmount()
		}
	}

	feeReport := marketFeeReportInfo{
		rangeStart, rangeEnd, totBaseFee, totQuoteFee, tradesFee,
	}

	return marketReportInfo{
		*mkt, feeReport, totVolume, subVolumes,
	}, nil
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
	ctx context.Context, password string,
	market ports.Market, outputs []ports.TxOutput, millisatsPerByte uint64,
) (string, error) {
	ok, err := s.wallet.Wallet().Auth(ctx, password)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("invalid password")
	}

	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return "", err
	}

	return s.wallet.SendToMany(mkt.Name, outputs, millisatsPerByte)
}

func (s *service) UpdateMarketPercentageFee(
	ctx context.Context,
	market ports.Market, basePercentageFee, quotePercentageFee int64,
) (ports.MarketInfo, error) {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return nil, err
	}

	if err := mkt.ChangePercentageFee(
		basePercentageFee, quotePercentageFee,
	); err != nil {
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

func (s *service) UpdateMarketAssetsPrecision(
	ctx context.Context, market ports.Market, basePrecision, quotePrecision int,
) error {
	mkt, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.GetBaseAsset(), market.GetQuoteAsset(),
	)
	if err != nil {
		return err
	}

	if err := mkt.ChangeAssetPrecision(
		basePrecision, quotePrecision,
	); err != nil {
		return err
	}

	return s.repoManager.MarketRepository().UpdateMarket(
		ctx, mkt.Name, func(_ *domain.Market) (*domain.Market, error) {
			return mkt, nil
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

func timeRangeToDates(tr ports.TimeRange) (startTime, endTime time.Time, err error) {
	now := time.Now()
	endTime = now

	if p := tr.GetCustomPeriod(); p != nil {
		startTime, err = time.Parse(time.RFC3339, p.GetStartDate())
		if err != nil {
			return
		}

		if p.GetEndDate() != "" {
			endTime, err = time.Parse(time.RFC3339, p.GetEndDate())
			if err != nil {
				return
			}
		}
		return
	}

	p := tr.GetPredefinedPeriod()
	if p.IsLastHour() {
		startTime = now.Add(time.Duration(-60) * time.Minute)
	}
	if p.IsLastDay() {
		startTime = now.AddDate(0, 0, -1)
	}
	if p.IsLastWeek() {
		startTime = now.AddDate(0, 0, -7)
	}
	if p.IsLastMonth() {
		startTime = now.AddDate(0, -1, 0)
	}
	if p.IsLastThreeMonths() {
		startTime = now.AddDate(0, -3, 0)
	}
	if p.IsYearToDate() {
		y, _, _ := now.Date()
		startTime = time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC)
	}
	if p.IsLastYear() {
		y, _, _ := now.Date()
		startTime = time.Date(y-1, time.January, 1, 0, 0, 0, 0, time.UTC)
		endTime = time.Date(y-1, time.December, 31, 23, 59, 59, 0, time.UTC)
	}
	if p.IsAll() {
		startTime = time.Date(startYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	return
}

// splitTimeRange splits the given time range (start, end) into a list of
// sub-ranges of frame hours, ordered from end to start.
// Example:
// in: 2009-11-10 19:00:00 (start), 2009-11-11 00:00:00 (end), 1 (frame)
// out: [
//
//	{end: 2009-11-11 00:00:00, start: 2009-11-10 22:00:01},
//	{end: 2009-11-11 22:00:00, start: 2009-11-10 20:00:01},
//	{end: 2009-11-10 20:00:00, start: 2009-11-10 19:00:00},
//
// ]
func splitTimeRange(
	start, end time.Time, frame int,
) marketVolumeInfoList {
	subRanges := make([]*marketVolumeInfo, 0)
	for {
		if end.Equal(start) || end.Before(start) {
			return subRanges
		} else {
			nextEnd := end.Add(-time.Hour * time.Duration(frame))
			nextStart := start
			if nextEnd.Sub(start).Seconds() > 0 {
				nextStart = nextEnd.Add(time.Second)
			}
			subRanges = append(subRanges, &marketVolumeInfo{
				start: nextStart, end: end,
			})
			end = nextEnd
		}
	}
}

func isInTimeRange(t int64, start, end time.Time) bool {
	tt := time.Unix(t, 0)
	return !tt.Before(start) && !tt.After(end)
}

// getDefaultTimeFrameForRange returns the appropriate time frame (tf)
// based on the delta in days between start and end.
// - delta >= 5y -> tf = 1m
// - delta >= 1y -> tf = 7d
// - delta >= 3m -> tf = 1d
// - delta >= 7d -> tf = 4h
// - otherwise tf = 1h
func getDefaultTimeFrameForRange(start, end time.Time) int {
	delta := int64(end.Sub(start).Hours() / 24)

	if delta >= 365*5 {
		return 24 * 30
	}
	if delta >= 365 {
		return 24 * 7
	}
	if delta >= 88 {
		return 24
	}
	if delta >= 7 {
		return 4
	}
	return 1
}
