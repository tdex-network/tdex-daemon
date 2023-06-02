package postgresdb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/pg/sqlc/queries"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

const (
	marketFixedFeeKey      = "fixed"
	marketPercentageFeeKey = "percentage"
)

type marketRepositoryImpl struct {
	querier *queries.Queries
	execTx  func(
		ctx context.Context,
		txBody func(*queries.Queries) error,
	) error
}

func NewMarketRepositoryImpl(
	querier *queries.Queries,
	execTx func(ctx context.Context, txBody func(*queries.Queries) error) error,
) domain.MarketRepository {
	return &marketRepositoryImpl{
		querier: querier,
		execTx:  execTx,
	}
}

func (m *marketRepositoryImpl) AddMarket(
	ctx context.Context, market *domain.Market,
) error {
	txBody := func(querierWithTx *queries.Queries) error {
		var basePrice float64 = 0
		if market.Price.BasePrice != "" {
			price, err := strconv.ParseFloat(market.Price.BasePrice, 64)
			if err != nil {
				return err
			}

			basePrice = price
		}

		var quotePrice float64 = 0
		if market.Price.QuotePrice != "" {
			price, err := strconv.ParseFloat(market.Price.QuotePrice, 64)
			if err != nil {
				return err
			}

			quotePrice = price
		}

		if _, err := querierWithTx.InsertMarket(ctx, queries.InsertMarketParams{
			Name:                market.Name,
			BaseAsset:           market.BaseAsset,
			QuoteAsset:          market.QuoteAsset,
			BaseAssetPrecision:  int32(market.BaseAssetPrecision),
			QuoteAssetPrecision: int32(market.QuoteAssetPrecision),
			Tradable:            market.Tradable,
			StrategyType:        int32(market.StrategyType),
			BasePrice:           basePrice,
			QuotePrice:          quotePrice,
			Active:              true,
		}); err != nil {
			return err
		}

		if market.FixedFee.BaseAsset > 0 || market.FixedFee.QuoteAsset > 0 {
			if _, err := querierWithTx.InsertMarketFee(ctx, queries.InsertMarketFeeParams{
				BaseAssetFee:  int64(market.FixedFee.BaseAsset),
				QuoteAssetFee: int64(market.FixedFee.QuoteAsset),
				Type:          marketFixedFeeKey,
				FkMarketName:  market.Name,
			}); err != nil {
				return err
			}
		}

		if market.PercentageFee.BaseAsset > 0 || market.PercentageFee.QuoteAsset > 0 {
			if _, err := querierWithTx.InsertMarketFee(ctx, queries.InsertMarketFeeParams{
				BaseAssetFee:  int64(market.PercentageFee.BaseAsset),
				QuoteAssetFee: int64(market.PercentageFee.QuoteAsset),
				Type:          marketPercentageFeeKey,
				FkMarketName:  market.Name,
			}); err != nil {
				return err
			}
		}

		return nil
	}

	return m.execTx(ctx, txBody)
}

func (m *marketRepositoryImpl) GetMarketByName(
	ctx context.Context, marketName string,
) (*domain.Market, error) {
	mktRows, err := m.querier.GetMarketByName(ctx, marketName)
	if err != nil {
		return nil, err
	}

	if len(mktRows) == 0 {
		return nil, fmt.Errorf("market with name %s not found", marketName)
	}

	rows := make([]marketRow, 0, len(mktRows))
	for _, v := range mktRows {
		rows = append(rows, MarketByNameRow{v})
	}

	mkt := convertMarketRowsToMarket(rows)

	return mkt, nil
}

func (m *marketRepositoryImpl) GetMarketByAssets(
	ctx context.Context, baseAsset, quoteAsset string,
) (*domain.Market, error) {
	mktRows, err := m.querier.GetMarketByBaseAndQuoteAsset(
		ctx,
		queries.GetMarketByBaseAndQuoteAssetParams{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(mktRows) == 0 {
		return nil, fmt.Errorf(
			"market with assets %s %s not found", baseAsset, quoteAsset,
		)
	}

	rows := make([]marketRow, 0, len(mktRows))
	for _, v := range mktRows {
		rows = append(rows, MarketByBaseAndQuoteAssetRow{v})
	}

	mkt := convertMarketRowsToMarket(rows)

	return mkt, nil
}

func (m *marketRepositoryImpl) GetTradableMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	mktRows, err := m.querier.GetTradableMarkets(ctx)
	if err != nil {
		return nil, err
	}

	if len(mktRows) == 0 {
		return nil, nil
	}

	rows := make([]marketRow, 0, len(mktRows))
	for _, v := range mktRows {
		rows = append(rows, TradableMarketsRow{v})
	}

	mkt := convertMarketsRowsToMarkets(rows)

	return mkt, nil
}

func (m *marketRepositoryImpl) GetAllMarkets(
	ctx context.Context,
) ([]domain.Market, error) {
	mktRows, err := m.querier.GetAllMarkets(ctx)
	if err != nil {
		return nil, err
	}

	if len(mktRows) == 0 {
		return nil, nil
	}

	rows := make([]marketRow, 0, len(mktRows))
	for _, v := range mktRows {
		rows = append(rows, AllMarketsRow{v})
	}

	mkt := convertMarketsRowsToMarkets(rows)

	return mkt, nil
}

func (m *marketRepositoryImpl) UpdateMarket(
	ctx context.Context, marketName string,
	updateFn func(m *domain.Market) (*domain.Market, error),
) error {
	currentMarket, err := m.GetMarketByName(ctx, marketName)
	if err != nil {
		return err
	}
	if currentMarket == nil {
		return fmt.Errorf("market with name %s not found", marketName)
	}

	updatedMarket, err := updateFn(currentMarket)
	if err != nil {
		return err
	}

	return m.updateMarket(ctx, *updatedMarket)
}

func (m *marketRepositoryImpl) OpenMarket(
	ctx context.Context, marketName string,
) error {
	market, err := m.GetMarketByName(ctx, marketName)
	if err != nil {
		return err
	}

	if market.IsTradable() {
		return nil
	}

	err = market.MakeTradable()
	if err != nil {
		return err
	}

	return m.updateMarket(ctx, *market)
}

func (m *marketRepositoryImpl) CloseMarket(
	ctx context.Context, marketName string,
) error {
	market, err := m.GetMarketByName(ctx, marketName)
	if err != nil {
		return err
	}

	if !market.IsTradable() {
		return nil
	}

	market.MakeNotTradable()

	return m.updateMarket(ctx, *market)
}

func (m *marketRepositoryImpl) DeleteMarket(
	ctx context.Context, marketName string,
) error {
	market, err := m.GetMarketByName(ctx, marketName)
	if err != nil {
		return err
	}

	if market == nil {
		return fmt.Errorf("market with name %s not found", marketName)
	}

	return m.querier.InactivateMarket(ctx, marketName)
}

func (m *marketRepositoryImpl) UpdateMarketPrice(
	ctx context.Context, marketName string, price domain.MarketPrice,
) error {
	var basePrice float64 = 0
	if price.BasePrice != "" {
		price, err := strconv.ParseFloat(price.BasePrice, 64)
		if err != nil {
			return err
		}

		basePrice = price
	}

	var quotePrice float64 = 0
	if price.QuotePrice != "" {
		price, err := strconv.ParseFloat(price.QuotePrice, 64)
		if err != nil {
			return err
		}

		quotePrice = price
	}

	if _, err := m.querier.UpdateMarketPrice(ctx, queries.UpdateMarketPriceParams{
		BasePrice:  basePrice,
		QuotePrice: quotePrice,
		Name:       marketName,
	}); err != nil {
		return err
	}

	return nil
}

func (m *marketRepositoryImpl) updateMarket(
	ctx context.Context, market domain.Market,
) error {
	txBody := func(querierWithTx *queries.Queries) error {
		var basePrice float64 = 0
		if market.Price.BasePrice != "" {
			price, err := strconv.ParseFloat(market.Price.BasePrice, 64)
			if err != nil {
				return err
			}

			basePrice = price
		}

		var quotePrice float64 = 0
		if market.Price.QuotePrice != "" {
			price, err := strconv.ParseFloat(market.Price.QuotePrice, 64)
			if err != nil {
				return err
			}

			quotePrice = price
		}

		if _, err := querierWithTx.UpdateMarket(ctx, queries.UpdateMarketParams{
			BaseAssetPrecision:  int32(market.BaseAssetPrecision),
			QuoteAssetPrecision: int32(market.QuoteAssetPrecision),
			Tradable:            market.Tradable,
			StrategyType:        int32(market.StrategyType),
			BasePrice:           basePrice,
			QuotePrice:          quotePrice,
			Name:                market.Name,
		}); err != nil {
			return err
		}

		if market.FixedFee.BaseAsset > 0 || market.FixedFee.QuoteAsset > 0 {
			if _, err := querierWithTx.UpdateMarketFee(ctx, queries.UpdateMarketFeeParams{
				BaseAssetFee:  int64(market.FixedFee.BaseAsset),
				QuoteAssetFee: int64(market.FixedFee.QuoteAsset),
				FkMarketName:  market.Name,
				Type:          marketFixedFeeKey,
			}); err != nil {
				return err
			}
		}

		if market.PercentageFee.BaseAsset > 0 || market.PercentageFee.QuoteAsset > 0 {
			if _, err := querierWithTx.UpdateMarketFee(ctx, queries.UpdateMarketFeeParams{
				BaseAssetFee:  int64(market.PercentageFee.BaseAsset),
				QuoteAssetFee: int64(market.PercentageFee.QuoteAsset),
				FkMarketName:  market.Name,
				Type:          marketPercentageFeeKey,
			}); err != nil {
				return err
			}
		}

		return nil
	}

	return m.execTx(ctx, txBody)
}
