package postgresdb

import (
	"context"
	"database/sql"
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
		baseAssetPrecision := sql.NullInt32{}
		if market.BaseAssetPrecision > 0 {
			baseAssetPrecision.Int32 = int32(market.BaseAssetPrecision)
			baseAssetPrecision.Valid = true
		}

		quoteAssetPrecision := sql.NullInt32{}
		if market.QuoteAssetPrecision > 0 {
			quoteAssetPrecision.Int32 = int32(market.QuoteAssetPrecision)
			quoteAssetPrecision.Valid = true
		}

		basePrice := sql.NullFloat64{}
		if market.Price.BasePrice != "" {
			price, err := strconv.ParseFloat(market.Price.BasePrice, 64)
			if err != nil {
				return err
			}

			basePrice.Float64 = price
			basePrice.Valid = true
		}

		quotePrice := sql.NullFloat64{}
		if market.Price.QuotePrice != "" {
			price, err := strconv.ParseFloat(market.Price.QuotePrice, 64)
			if err != nil {
				return err
			}

			quotePrice.Float64 = price
			quotePrice.Valid = true
		}

		if _, err := querierWithTx.InsertMarket(ctx, queries.InsertMarketParams{
			Name:                market.Name,
			BaseAsset:           market.BaseAsset,
			QuoteAsset:          market.QuoteAsset,
			BaseAssetPrecision:  baseAssetPrecision,
			QuoteAssetPrecision: quoteAssetPrecision,
			Tradable: sql.NullBool{
				Bool:  market.Tradable,
				Valid: true,
			},
			StrategyType: sql.NullInt32{
				Int32: int32(market.StrategyType),
				Valid: true,
			},
			BasePrice:  basePrice,
			QuotePrice: quotePrice,
		}); err != nil {
			return err
		}

		if market.FixedFee.BaseAsset > 0 || market.FixedFee.QuoteAsset > 0 {
			if _, err := querierWithTx.InsertFee(ctx, queries.InsertFeeParams{
				BaseAssetFee:  int64(market.FixedFee.BaseAsset),
				QuoteAssetFee: int64(market.FixedFee.QuoteAsset),
				Type:          marketFixedFeeKey,
				FkMarketName:  market.Name,
			}); err != nil {
				return err
			}
		}

		if market.PercentageFee.BaseAsset > 0 || market.PercentageFee.QuoteAsset > 0 {
			if _, err := querierWithTx.InsertFee(ctx, queries.InsertFeeParams{
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

	txBody := func(querierWithTx *queries.Queries) error {
		if err := querierWithTx.DeleteFeeForMarket(ctx, marketName); err != nil {
			return err
		}

		return querierWithTx.DeleteMarket(ctx, marketName)
	}

	return m.execTx(ctx, txBody)
}

func (m *marketRepositoryImpl) UpdateMarketPrice(
	ctx context.Context, marketName string, price domain.MarketPrice,
) error {
	basePrice := sql.NullFloat64{}
	if price.BasePrice != "" {
		p, err := strconv.ParseFloat(price.BasePrice, 64)
		if err != nil {
			return err
		}

		basePrice = sql.NullFloat64{
			Float64: p,
			Valid:   true,
		}
	}

	quotePrice := sql.NullFloat64{}
	if price.QuotePrice != "" {
		p, err := strconv.ParseFloat(price.QuotePrice, 64)
		if err != nil {
			return err
		}

		quotePrice = sql.NullFloat64{
			Float64: p,
			Valid:   true,
		}
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
		baseAssetPrecision := sql.NullInt32{}
		if market.BaseAssetPrecision > 0 {
			baseAssetPrecision = sql.NullInt32{
				Int32: int32(market.BaseAssetPrecision),
				Valid: true,
			}
		}

		quoteAssetPrecision := sql.NullInt32{}
		if market.QuoteAssetPrecision > 0 {
			quoteAssetPrecision = sql.NullInt32{
				Int32: int32(market.QuoteAssetPrecision),
				Valid: true,
			}
		}

		tradable := sql.NullBool{}
		if market.Tradable {
			tradable = sql.NullBool{
				Bool:  market.Tradable,
				Valid: true,
			}
		}

		strategyType := sql.NullInt32{}
		if market.StrategyType > 0 {
			strategyType = sql.NullInt32{
				Int32: int32(market.StrategyType),
				Valid: true,
			}
		}

		basePrice := sql.NullFloat64{}
		if market.Price.BasePrice != "" {
			price, err := strconv.ParseFloat(market.Price.BasePrice, 64)
			if err != nil {
				return err
			}

			basePrice = sql.NullFloat64{
				Float64: price,
				Valid:   true,
			}
		}

		quotePrice := sql.NullFloat64{}
		if market.Price.QuotePrice != "" {
			price, err := strconv.ParseFloat(market.Price.QuotePrice, 64)
			if err != nil {
				return err
			}

			quotePrice = sql.NullFloat64{
				Float64: price,
				Valid:   true,
			}
		}

		if _, err := querierWithTx.UpdateMarket(ctx, queries.UpdateMarketParams{
			BaseAssetPrecision:  baseAssetPrecision,
			QuoteAssetPrecision: quoteAssetPrecision,
			Tradable:            tradable,
			StrategyType:        strategyType,
			BasePrice:           basePrice,
			QuotePrice:          quotePrice,
			Name:                market.Name,
		}); err != nil {
			return err
		}

		if market.FixedFee.BaseAsset > 0 || market.FixedFee.QuoteAsset > 0 {
			if _, err := querierWithTx.UpdateFee(ctx, queries.UpdateFeeParams{
				BaseAssetFee:  int64(market.FixedFee.BaseAsset),
				QuoteAssetFee: int64(market.FixedFee.QuoteAsset),
				FkMarketName:  market.Name,
				Type:          marketFixedFeeKey,
			}); err != nil {
				return err
			}
		}

		if market.PercentageFee.BaseAsset > 0 || market.PercentageFee.QuoteAsset > 0 {
			if _, err := querierWithTx.UpdateFee(ctx, queries.UpdateFeeParams{
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
