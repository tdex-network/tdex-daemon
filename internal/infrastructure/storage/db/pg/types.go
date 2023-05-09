package postgresdb

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/pg/sqlc/queries"
)

type marketRow interface {
	GetName() string
	GetBaseAsset() string
	GetQuoteAsset() string
	GetBaseAssetPrecision() sql.NullInt32
	GetQuoteAssetPrecision() sql.NullInt32
	GetTradable() sql.NullBool
	GetStrategyType() sql.NullInt32
	GetBasePrice() sql.NullFloat64
	GetQuotePrice() sql.NullFloat64
	GetID() int32
	GetBaseAssetFee() int64
	GetQuoteAssetFee() int64
	GetType() string
	GetFkMarketName() string
}

type MarketByNameRow struct {
	queries.GetMarketByNameRow
}

func (m MarketByNameRow) GetName() string {
	return m.Name
}

func (m MarketByNameRow) GetBaseAsset() string {
	return m.BaseAsset
}

func (m MarketByNameRow) GetQuoteAsset() string {
	return m.QuoteAsset
}

func (m MarketByNameRow) GetBaseAssetPrecision() sql.NullInt32 {
	return m.BaseAssetPrecision
}

func (m MarketByNameRow) GetQuoteAssetPrecision() sql.NullInt32 {
	return m.QuoteAssetPrecision
}

func (m MarketByNameRow) GetTradable() sql.NullBool {
	return m.Tradable
}

func (m MarketByNameRow) GetStrategyType() sql.NullInt32 {
	return m.StrategyType
}

func (m MarketByNameRow) GetBasePrice() sql.NullFloat64 {
	return m.BasePrice
}

func (m MarketByNameRow) GetQuotePrice() sql.NullFloat64 {
	return m.QuotePrice
}

func (m MarketByNameRow) GetID() int32 {
	return m.ID
}

func (m MarketByNameRow) GetBaseAssetFee() int64 {
	return m.BaseAssetFee
}

func (m MarketByNameRow) GetQuoteAssetFee() int64 {
	return m.QuoteAssetFee
}

func (m MarketByNameRow) GetType() string {
	return m.Type
}

func (m MarketByNameRow) GetFkMarketName() string {
	return m.FkMarketName
}

type TradableMarketsRow struct {
	queries.GetTradableMarketsRow
}

func (m TradableMarketsRow) GetName() string {
	return m.Name
}

func (m TradableMarketsRow) GetBaseAsset() string {
	return m.BaseAsset
}

func (m TradableMarketsRow) GetQuoteAsset() string {
	return m.QuoteAsset
}

func (m TradableMarketsRow) GetBaseAssetPrecision() sql.NullInt32 {
	return m.BaseAssetPrecision
}

func (m TradableMarketsRow) GetQuoteAssetPrecision() sql.NullInt32 {
	return m.QuoteAssetPrecision
}

func (m TradableMarketsRow) GetTradable() sql.NullBool {
	return m.Tradable
}

func (m TradableMarketsRow) GetStrategyType() sql.NullInt32 {
	return m.StrategyType
}

func (m TradableMarketsRow) GetBasePrice() sql.NullFloat64 {
	return m.BasePrice
}

func (m TradableMarketsRow) GetQuotePrice() sql.NullFloat64 {
	return m.QuotePrice
}

func (m TradableMarketsRow) GetID() int32 {
	return m.ID
}

func (m TradableMarketsRow) GetBaseAssetFee() int64 {
	return m.BaseAssetFee
}

func (m TradableMarketsRow) GetQuoteAssetFee() int64 {
	return m.QuoteAssetFee
}

func (m TradableMarketsRow) GetType() string {
	return m.Type
}

func (m TradableMarketsRow) GetFkMarketName() string {
	return m.FkMarketName
}

type MarketByBaseAndQuoteAssetRow struct {
	queries.GetMarketByBaseAndQuoteAssetRow
}

func (m MarketByBaseAndQuoteAssetRow) GetName() string {
	return m.Name
}

func (m MarketByBaseAndQuoteAssetRow) GetBaseAsset() string {
	return m.BaseAsset
}

func (m MarketByBaseAndQuoteAssetRow) GetQuoteAsset() string {
	return m.QuoteAsset
}

func (m MarketByBaseAndQuoteAssetRow) GetBaseAssetPrecision() sql.NullInt32 {
	return m.BaseAssetPrecision
}

func (m MarketByBaseAndQuoteAssetRow) GetQuoteAssetPrecision() sql.NullInt32 {
	return m.QuoteAssetPrecision
}

func (m MarketByBaseAndQuoteAssetRow) GetTradable() sql.NullBool {
	return m.Tradable
}

func (m MarketByBaseAndQuoteAssetRow) GetStrategyType() sql.NullInt32 {
	return m.StrategyType
}

func (m MarketByBaseAndQuoteAssetRow) GetBasePrice() sql.NullFloat64 {
	return m.BasePrice
}

func (m MarketByBaseAndQuoteAssetRow) GetQuotePrice() sql.NullFloat64 {
	return m.QuotePrice
}

func (m MarketByBaseAndQuoteAssetRow) GetID() int32 {
	return m.ID
}

func (m MarketByBaseAndQuoteAssetRow) GetBaseAssetFee() int64 {
	return m.BaseAssetFee
}

func (m MarketByBaseAndQuoteAssetRow) GetQuoteAssetFee() int64 {
	return m.QuoteAssetFee
}

func (m MarketByBaseAndQuoteAssetRow) GetType() string {
	return m.Type
}

func (m MarketByBaseAndQuoteAssetRow) GetFkMarketName() string {
	return m.FkMarketName
}

type AllMarketsRow struct {
	queries.GetAllMarketsRow
}

func (m AllMarketsRow) GetName() string {
	return m.Name
}

func (m AllMarketsRow) GetBaseAsset() string {
	return m.BaseAsset
}

func (m AllMarketsRow) GetQuoteAsset() string {
	return m.QuoteAsset
}

func (m AllMarketsRow) GetBaseAssetPrecision() sql.NullInt32 {
	return m.BaseAssetPrecision
}

func (m AllMarketsRow) GetQuoteAssetPrecision() sql.NullInt32 {
	return m.QuoteAssetPrecision
}

func (m AllMarketsRow) GetTradable() sql.NullBool {
	return m.Tradable
}

func (m AllMarketsRow) GetStrategyType() sql.NullInt32 {
	return m.StrategyType
}

func (m AllMarketsRow) GetBasePrice() sql.NullFloat64 {
	return m.BasePrice
}

func (m AllMarketsRow) GetQuotePrice() sql.NullFloat64 {
	return m.QuotePrice
}

func (m AllMarketsRow) GetID() int32 {
	return m.ID
}

func (m AllMarketsRow) GetBaseAssetFee() int64 {
	return m.BaseAssetFee
}

func (m AllMarketsRow) GetQuoteAssetFee() int64 {
	return m.QuoteAssetFee
}

func (m AllMarketsRow) GetType() string {
	return m.Type
}

func (m AllMarketsRow) GetFkMarketName() string {
	return m.FkMarketName
}

// convertMarketsRowsToMarkets takes a slice of marketRow instances (mktRows),
// which represents multiple markets, and converts it into a slice of domain.Market.
// This function groups market rows that belong to the same market and creates a domain.Market
// for each group.
func convertMarketsRowsToMarkets(mktRows []marketRow) []domain.Market {
	newMarketStartIndex := 0
	newMarketEndIndex := 0
	markets := make([]domain.Market, 0)
	//bellow identifies the start and end indices of each market in the mktRows
	//slice, and converting the corresponding sub-slice of marketRows into a domain.Market
	for i, mktRow := range mktRows {
		if i == 0 {
			newMarketStartIndex = 0
		} else if mktRow.GetName() != mktRows[i-1].GetName() {
			newMarketEndIndex = i
			markets = append(
				markets,
				*convertMarketRowsToMarket(mktRows[newMarketStartIndex:newMarketEndIndex]),
			)
			newMarketStartIndex = i
		}
	}

	// convert last sub-slice of marketRows into a domain.Market
	if len(mktRows) > 0 {
		markets = append(
			markets, *convertMarketRowsToMarket(mktRows[newMarketStartIndex:]),
		)
	}

	return markets
}

// convertMarketRowsToMarket takes a slice of marketRow instances (mktRows), which represents
// a single market (with multiple rows due to the join with the market_fee table), and
// converts it into a domain.Market instance. The function extracts the common properties
// of the market and combines the fees from different rows to create a single domain.Market.
func convertMarketRowsToMarket(mktRows []marketRow) *domain.Market {
	mktRow := mktRows[0]

	var baseAssetPrecision uint = 0
	if mktRow.GetBaseAssetPrecision().Valid {
		baseAssetPrecision = uint(mktRow.GetBaseAssetPrecision().Int32)
	}

	var quuoteAssetPrecision uint = 0
	if mktRow.GetQuoteAssetPrecision().Valid {
		quuoteAssetPrecision = uint(mktRow.GetQuoteAssetPrecision().Int32)
	}

	basePrice := ""
	if mktRow.GetBasePrice().Valid {
		basePrice = fmt.Sprintf("%f", mktRow.GetBasePrice().Float64)
	}

	quotePrice := ""
	if mktRow.GetQuotePrice().Valid {
		quotePrice = fmt.Sprintf("%f", mktRow.GetQuotePrice().Float64)
	}

	percentageFee := domain.MarketFee{}
	fixedFee := domain.MarketFee{}
	for _, v := range mktRows {
		if v.GetType() == marketPercentageFeeKey {
			percentageFee.BaseAsset = uint64(v.GetBaseAssetFee())
			percentageFee.QuoteAsset = uint64(v.GetQuoteAssetFee())
		} else if v.GetType() == marketFixedFeeKey {
			fixedFee.BaseAsset = uint64(v.GetBaseAssetFee())
			fixedFee.QuoteAsset = uint64(v.GetQuoteAssetFee())
		}
	}

	tradable := false
	if mktRow.GetTradable().Valid {
		tradable = mktRow.GetTradable().Bool
	}

	strategyType := 0
	if mktRow.GetStrategyType().Valid {
		strategyType = int(mktRow.GetStrategyType().Int32)
	}

	return &domain.Market{
		BaseAsset:           mktRow.GetBaseAsset(),
		QuoteAsset:          mktRow.GetQuoteAsset(),
		Name:                mktRow.GetName(),
		BaseAssetPrecision:  baseAssetPrecision,
		QuoteAssetPrecision: quuoteAssetPrecision,
		PercentageFee:       percentageFee,
		FixedFee:            fixedFee,
		Tradable:            tradable,
		StrategyType:        strategyType,
		Price: domain.MarketPrice{
			BasePrice:  basePrice,
			QuotePrice: quotePrice,
		},
	}
}

type tradeRow interface {
	GetID() string
	GetType() int32
	GetFeeAsset() string
	GetFeeAmount() int64
	GetTraderPubkey() []byte
	GetStatusCode() sql.NullInt32
	GetStatusFailed() sql.NullBool
	GetPsetBase64() string
	GetTxID() sql.NullString
	GetTxHex() string
	GetExpiryTime() sql.NullInt64
	GetSettlementTime() sql.NullInt64
	GetFkMarketName() string
	GetName() string
	GetBaseAsset() string
	GetQuoteAsset() string
	GetBasePrice() sql.NullFloat64
	GetQuotePrice() sql.NullFloat64
	GetSwapID() string
	GetMessage() []byte
	GetTimestamp() int64
	GetSwapType() string
	GetBaseAssetFee() int64
	GetQuoteAssetFee() int64
	GetFeeType() string
}

type TradeByIdRow struct {
	queries.GetTradeByIdRow
}

func (t TradeByIdRow) GetID() string {
	return t.ID
}

func (t TradeByIdRow) GetType() int32 {
	return t.Type
}

func (t TradeByIdRow) GetFeeAsset() string {
	return t.FeeAsset
}

func (t TradeByIdRow) GetFeeAmount() int64 {
	return t.FeeAmount
}

func (t TradeByIdRow) GetTraderPubkey() []byte {
	return t.TraderPubkey
}

func (t TradeByIdRow) GetStatusCode() sql.NullInt32 {
	return t.StatusCode
}

func (t TradeByIdRow) GetStatusFailed() sql.NullBool {
	return t.StatusFailed
}

func (t TradeByIdRow) GetPsetBase64() string {
	return t.PsetBase64
}

func (t TradeByIdRow) GetTxID() sql.NullString {
	return t.TxID
}

func (t TradeByIdRow) GetTxHex() string {
	return t.TxHex
}

func (t TradeByIdRow) GetExpiryTime() sql.NullInt64 {
	return t.ExpiryTime
}

func (t TradeByIdRow) GetSettlementTime() sql.NullInt64 {
	return t.SettlementTime
}

func (t TradeByIdRow) GetFkMarketName() string {
	return t.FkMarketName
}

func (t TradeByIdRow) GetName() string {
	return t.Name
}

func (t TradeByIdRow) GetBaseAsset() string {
	return t.BaseAsset
}

func (t TradeByIdRow) GetQuoteAsset() string {
	return t.QuoteAsset
}

func (t TradeByIdRow) GetBasePrice() sql.NullFloat64 {
	return t.BasePrice
}

func (t TradeByIdRow) GetQuotePrice() sql.NullFloat64 {
	return t.QuotePrice
}

func (t TradeByIdRow) GetSwapID() string {
	return t.SwapID
}

func (t TradeByIdRow) GetMessage() []byte {
	return t.Message
}

func (t TradeByIdRow) GetTimestamp() int64 {
	return t.Timestamp
}

func (t TradeByIdRow) GetSwapType() string {
	return t.SwapType
}

func (t TradeByIdRow) GetBaseAssetFee() int64 {
	return t.BaseAssetFee
}

func (t TradeByIdRow) GetQuoteAssetFee() int64 {
	return t.QuoteAssetFee
}

func (t TradeByIdRow) GetFeeType() string {
	return t.FeeType
}

type AllTradesRow struct {
	queries.GetAllTradesRow
}

func (t AllTradesRow) GetID() string {
	return t.ID
}

func (t AllTradesRow) GetType() int32 {
	return t.Type
}

func (t AllTradesRow) GetFeeAsset() string {
	return t.FeeAsset
}

func (t AllTradesRow) GetFeeAmount() int64 {
	return t.FeeAmount
}

func (t AllTradesRow) GetTraderPubkey() []byte {
	return t.TraderPubkey
}

func (t AllTradesRow) GetStatusCode() sql.NullInt32 {
	return t.StatusCode
}

func (t AllTradesRow) GetStatusFailed() sql.NullBool {
	return t.StatusFailed
}

func (t AllTradesRow) GetPsetBase64() string {
	return t.PsetBase64
}

func (t AllTradesRow) GetTxID() sql.NullString {
	return t.TxID
}

func (t AllTradesRow) GetTxHex() string {
	return t.TxHex
}

func (t AllTradesRow) GetExpiryTime() sql.NullInt64 {
	return t.ExpiryTime
}

func (t AllTradesRow) GetSettlementTime() sql.NullInt64 {
	return t.SettlementTime
}

func (t AllTradesRow) GetFkMarketName() string {
	return t.FkMarketName
}

func (t AllTradesRow) GetName() string {
	return t.Name
}

func (t AllTradesRow) GetBaseAsset() string {
	return t.BaseAsset
}

func (t AllTradesRow) GetQuoteAsset() string {
	return t.QuoteAsset
}

func (t AllTradesRow) GetBasePrice() sql.NullFloat64 {
	return t.BasePrice
}

func (t AllTradesRow) GetQuotePrice() sql.NullFloat64 {
	return t.QuotePrice
}

func (t AllTradesRow) GetSwapID() string {
	return t.SwapID
}

func (t AllTradesRow) GetMessage() []byte {
	return t.Message
}

func (t AllTradesRow) GetTimestamp() int64 {
	return t.Timestamp
}

func (t AllTradesRow) GetSwapType() string {
	return t.SwapType
}

func (t AllTradesRow) GetBaseAssetFee() int64 {
	return t.BaseAssetFee
}

func (t AllTradesRow) GetQuoteAssetFee() int64 {
	return t.QuoteAssetFee
}

func (t AllTradesRow) GetFeeType() string {
	return t.FeeType
}

type AllTradesByMarketRow struct {
	queries.GetAllTradesByMarketRow
}

func (t AllTradesByMarketRow) GetID() string {
	return t.ID
}

func (t AllTradesByMarketRow) GetType() int32 {
	return t.Type
}

func (t AllTradesByMarketRow) GetFeeAsset() string {
	return t.FeeAsset
}

func (t AllTradesByMarketRow) GetFeeAmount() int64 {
	return t.FeeAmount
}

func (t AllTradesByMarketRow) GetTraderPubkey() []byte {
	return t.TraderPubkey
}

func (t AllTradesByMarketRow) GetStatusCode() sql.NullInt32 {
	return t.StatusCode
}

func (t AllTradesByMarketRow) GetStatusFailed() sql.NullBool {
	return t.StatusFailed
}

func (t AllTradesByMarketRow) GetPsetBase64() string {
	return t.PsetBase64
}

func (t AllTradesByMarketRow) GetTxID() sql.NullString {
	return t.TxID
}

func (t AllTradesByMarketRow) GetTxHex() string {
	return t.TxHex
}

func (t AllTradesByMarketRow) GetExpiryTime() sql.NullInt64 {
	return t.ExpiryTime
}

func (t AllTradesByMarketRow) GetSettlementTime() sql.NullInt64 {
	return t.SettlementTime
}

func (t AllTradesByMarketRow) GetFkMarketName() string {
	return t.FkMarketName
}

func (t AllTradesByMarketRow) GetName() string {
	return t.Name
}

func (t AllTradesByMarketRow) GetBaseAsset() string {
	return t.BaseAsset
}

func (t AllTradesByMarketRow) GetQuoteAsset() string {
	return t.QuoteAsset
}

func (t AllTradesByMarketRow) GetBasePrice() sql.NullFloat64 {
	return t.BasePrice
}

func (t AllTradesByMarketRow) GetQuotePrice() sql.NullFloat64 {
	return t.QuotePrice
}

func (t AllTradesByMarketRow) GetSwapID() string {
	return t.SwapID
}

func (t AllTradesByMarketRow) GetMessage() []byte {
	return t.Message
}

func (t AllTradesByMarketRow) GetTimestamp() int64 {
	return t.Timestamp
}

func (t AllTradesByMarketRow) GetSwapType() string {
	return t.SwapType
}

func (t AllTradesByMarketRow) GetBaseAssetFee() int64 {
	return t.BaseAssetFee
}

func (t AllTradesByMarketRow) GetQuoteAssetFee() int64 {
	return t.QuoteAssetFee
}

func (t AllTradesByMarketRow) GetFeeType() string {
	return t.FeeType
}

type TradesByMarketAndStatusRow struct {
	queries.GetTradesByMarketAndStatusRow
}

func (t TradesByMarketAndStatusRow) GetID() string {
	return t.ID
}

func (t TradesByMarketAndStatusRow) GetType() int32 {
	return t.Type
}

func (t TradesByMarketAndStatusRow) GetFeeAsset() string {
	return t.FeeAsset
}

func (t TradesByMarketAndStatusRow) GetFeeAmount() int64 {
	return t.FeeAmount
}

func (t TradesByMarketAndStatusRow) GetTraderPubkey() []byte {
	return t.TraderPubkey
}

func (t TradesByMarketAndStatusRow) GetStatusCode() sql.NullInt32 {
	return t.StatusCode
}

func (t TradesByMarketAndStatusRow) GetStatusFailed() sql.NullBool {
	return t.StatusFailed
}

func (t TradesByMarketAndStatusRow) GetPsetBase64() string {
	return t.PsetBase64
}

func (t TradesByMarketAndStatusRow) GetTxID() sql.NullString {
	return t.TxID
}

func (t TradesByMarketAndStatusRow) GetTxHex() string {
	return t.TxHex
}

func (t TradesByMarketAndStatusRow) GetExpiryTime() sql.NullInt64 {
	return t.ExpiryTime
}

func (t TradesByMarketAndStatusRow) GetSettlementTime() sql.NullInt64 {
	return t.SettlementTime
}

func (t TradesByMarketAndStatusRow) GetFkMarketName() string {
	return t.FkMarketName
}

func (t TradesByMarketAndStatusRow) GetName() string {
	return t.Name
}

func (t TradesByMarketAndStatusRow) GetBaseAsset() string {
	return t.BaseAsset
}

func (t TradesByMarketAndStatusRow) GetQuoteAsset() string {
	return t.QuoteAsset
}

func (t TradesByMarketAndStatusRow) GetBasePrice() sql.NullFloat64 {
	return t.BasePrice
}

func (t TradesByMarketAndStatusRow) GetQuotePrice() sql.NullFloat64 {
	return t.QuotePrice
}

func (t TradesByMarketAndStatusRow) GetSwapID() string {
	return t.SwapID
}

func (t TradesByMarketAndStatusRow) GetMessage() []byte {
	return t.Message
}

func (t TradesByMarketAndStatusRow) GetTimestamp() int64 {
	return t.Timestamp
}

func (t TradesByMarketAndStatusRow) GetSwapType() string {
	return t.SwapType
}

func (t TradesByMarketAndStatusRow) GetBaseAssetFee() int64 {
	return t.BaseAssetFee
}

func (t TradesByMarketAndStatusRow) GetQuoteAssetFee() int64 {
	return t.QuoteAssetFee
}

func (t TradesByMarketAndStatusRow) GetFeeType() string {
	return t.FeeType
}

type GetTradeBySwapAcceptIdRow struct {
	queries.GetTradeBySwapAcceptIdRow
}

func (t GetTradeBySwapAcceptIdRow) GetID() string {
	return t.ID
}

func (t GetTradeBySwapAcceptIdRow) GetType() int32 {
	return t.Type
}

func (t GetTradeBySwapAcceptIdRow) GetFeeAsset() string {
	return t.FeeAsset
}

func (t GetTradeBySwapAcceptIdRow) GetFeeAmount() int64 {
	return t.FeeAmount
}

func (t GetTradeBySwapAcceptIdRow) GetTraderPubkey() []byte {
	return t.TraderPubkey
}

func (t GetTradeBySwapAcceptIdRow) GetStatusCode() sql.NullInt32 {
	return t.StatusCode
}

func (t GetTradeBySwapAcceptIdRow) GetStatusFailed() sql.NullBool {
	return t.StatusFailed
}

func (t GetTradeBySwapAcceptIdRow) GetPsetBase64() string {
	return t.PsetBase64
}

func (t GetTradeBySwapAcceptIdRow) GetTxID() sql.NullString {
	return t.TxID
}

func (t GetTradeBySwapAcceptIdRow) GetTxHex() string {
	return t.TxHex
}

func (t GetTradeBySwapAcceptIdRow) GetExpiryTime() sql.NullInt64 {
	return t.ExpiryTime
}

func (t GetTradeBySwapAcceptIdRow) GetSettlementTime() sql.NullInt64 {
	return t.SettlementTime
}

func (t GetTradeBySwapAcceptIdRow) GetFkMarketName() string {
	return t.FkMarketName
}

func (t GetTradeBySwapAcceptIdRow) GetName() string {
	return t.Name
}

func (t GetTradeBySwapAcceptIdRow) GetBaseAsset() string {
	return t.BaseAsset
}

func (t GetTradeBySwapAcceptIdRow) GetQuoteAsset() string {
	return t.QuoteAsset
}

func (t GetTradeBySwapAcceptIdRow) GetBasePrice() sql.NullFloat64 {
	return t.BasePrice
}

func (t GetTradeBySwapAcceptIdRow) GetQuotePrice() sql.NullFloat64 {
	return t.QuotePrice
}

func (t GetTradeBySwapAcceptIdRow) GetSwapID() string {
	return t.SwapID
}

func (t GetTradeBySwapAcceptIdRow) GetMessage() []byte {
	return t.Message
}

func (t GetTradeBySwapAcceptIdRow) GetTimestamp() int64 {
	return t.Timestamp
}

func (t GetTradeBySwapAcceptIdRow) GetSwapType() string {
	return t.SwapType
}

func (t GetTradeBySwapAcceptIdRow) GetBaseAssetFee() int64 {
	return t.BaseAssetFee
}

func (t GetTradeBySwapAcceptIdRow) GetQuoteAssetFee() int64 {
	return t.QuoteAssetFee
}

func (t GetTradeBySwapAcceptIdRow) GetFeeType() string {
	return t.FeeType
}

type GetTradeByTxIdRow struct {
	queries.GetTradeByTxIdRow
}

func (t GetTradeByTxIdRow) GetID() string {
	return t.ID
}

func (t GetTradeByTxIdRow) GetType() int32 {
	return t.Type
}

func (t GetTradeByTxIdRow) GetFeeAsset() string {
	return t.FeeAsset
}

func (t GetTradeByTxIdRow) GetFeeAmount() int64 {
	return t.FeeAmount
}

func (t GetTradeByTxIdRow) GetTraderPubkey() []byte {
	return t.TraderPubkey
}

func (t GetTradeByTxIdRow) GetStatusCode() sql.NullInt32 {
	return t.StatusCode
}

func (t GetTradeByTxIdRow) GetStatusFailed() sql.NullBool {
	return t.StatusFailed
}

func (t GetTradeByTxIdRow) GetPsetBase64() string {
	return t.PsetBase64
}

func (t GetTradeByTxIdRow) GetTxID() sql.NullString {
	return t.TxID
}

func (t GetTradeByTxIdRow) GetTxHex() string {
	return t.TxHex
}

func (t GetTradeByTxIdRow) GetExpiryTime() sql.NullInt64 {
	return t.ExpiryTime
}

func (t GetTradeByTxIdRow) GetSettlementTime() sql.NullInt64 {
	return t.SettlementTime
}

func (t GetTradeByTxIdRow) GetFkMarketName() string {
	return t.FkMarketName
}

func (t GetTradeByTxIdRow) GetName() string {
	return t.Name
}

func (t GetTradeByTxIdRow) GetBaseAsset() string {
	return t.BaseAsset
}

func (t GetTradeByTxIdRow) GetQuoteAsset() string {
	return t.QuoteAsset
}

func (t GetTradeByTxIdRow) GetBasePrice() sql.NullFloat64 {
	return t.BasePrice
}

func (t GetTradeByTxIdRow) GetQuotePrice() sql.NullFloat64 {
	return t.QuotePrice
}

func (t GetTradeByTxIdRow) GetSwapID() string {
	return t.SwapID
}

func (t GetTradeByTxIdRow) GetMessage() []byte {
	return t.Message
}

func (t GetTradeByTxIdRow) GetTimestamp() int64 {
	return t.Timestamp
}

func (t GetTradeByTxIdRow) GetSwapType() string {
	return t.SwapType
}

func (t GetTradeByTxIdRow) GetBaseAssetFee() int64 {
	return t.BaseAssetFee
}

func (t GetTradeByTxIdRow) GetQuoteAssetFee() int64 {
	return t.QuoteAssetFee
}

func (t GetTradeByTxIdRow) GetFeeType() string {
	return t.FeeType
}

func convertTradesRowsToTrades(tradesRows []tradeRow) ([]domain.Trade, error) {
	trades := make([]domain.Trade, 0)
	newTradeStartIndex := 0
	newTradeEndIndex := 0

	for i, tradeRow := range tradesRows {
		if i == 0 {
			newTradeStartIndex = 0
		} else if tradeRow.GetID() != tradesRows[i-1].GetID() {
			newTradeEndIndex = i

			trd, err := convertTradeRowsToTrade(
				tradesRows[newTradeStartIndex:newTradeEndIndex],
			)
			if err != nil {
				return nil, err
			}

			trades = append(trades, *trd)
			newTradeStartIndex = i
		}
	}

	if len(tradesRows) > 0 {
		trd, err := convertTradeRowsToTrade(
			tradesRows[newTradeStartIndex:],
		)
		if err != nil {
			return nil, err
		}

		trades = append(trades, *trd)
	}

	sort.SliceStable(trades, func(i, j int) bool {
		return trades[i].SwapRequest != nil && trades[j].SwapRequest != nil &&
			trades[i].SwapRequest.Timestamp > trades[j].SwapRequest.Timestamp
	})

	return trades, nil
}

func convertTradeRowsToTrade(tradeRows []tradeRow) (*domain.Trade, error) {
	tradeRow := tradeRows[0]

	basePrice := ""
	if tradeRow.GetBasePrice().Valid {
		basePrice = fmt.Sprintf("%f", tradeRow.GetBasePrice().Float64)
	}

	quotePrice := ""
	if tradeRow.GetQuotePrice().Valid {
		quotePrice = fmt.Sprintf("%f", tradeRow.GetQuotePrice().Float64)
	}

	statusCode := 0
	if tradeRow.GetStatusCode().Valid {
		statusCode = int(tradeRow.GetStatusCode().Int32)
	}

	statusFailed := false
	if tradeRow.GetStatusFailed().Valid {
		statusFailed = tradeRow.GetStatusFailed().Bool
	}

	txId := ""
	if tradeRow.GetTxID().Valid {
		txId = tradeRow.GetTxID().String
	}

	expiryTime := 0
	if tradeRow.GetExpiryTime().Valid {
		expiryTime = int(tradeRow.GetExpiryTime().Int64)
	}

	settlementTime := 0
	if tradeRow.GetSettlementTime().Valid {
		settlementTime = int(tradeRow.GetSettlementTime().Int64)
	}

	percentageFee := domain.MarketFee{}
	fixedFee := domain.MarketFee{}
	var swapAccept *domain.Swap
	var swapComplete *domain.Swap
	var swapFail *domain.Swap
	var swapRequest *domain.Swap
	for _, v := range tradeRows {
		if v.GetFeeType() == marketPercentageFeeKey {
			percentageFee = domain.MarketFee{
				BaseAsset:  uint64(v.GetBaseAssetFee()),
				QuoteAsset: uint64(v.GetQuoteAssetFee()),
			}
		}

		if v.GetFeeType() == marketFixedFeeKey {
			fixedFee = domain.MarketFee{
				BaseAsset:  uint64(v.GetBaseAssetFee()),
				QuoteAsset: uint64(v.GetQuoteAssetFee()),
			}
		}

		if v.GetSwapType() == swapAcceptKey {
			swapAccept = &domain.Swap{
				Id:        v.GetSwapID(),
				Message:   v.GetMessage(),
				Timestamp: v.GetTimestamp(),
			}
		}

		if v.GetSwapType() == swapCompleteKey {
			swapComplete = &domain.Swap{
				Id:        v.GetSwapID(),
				Message:   v.GetMessage(),
				Timestamp: v.GetTimestamp(),
			}
		}

		if v.GetSwapType() == swapFailKey {
			swapFail = &domain.Swap{
				Id:        v.GetSwapID(),
				Message:   v.GetMessage(),
				Timestamp: v.GetTimestamp(),
			}
		}

		if v.GetSwapType() == swapRequestKey {
			swapRequest = &domain.Swap{
				Id:        v.GetSwapID(),
				Message:   v.GetMessage(),
				Timestamp: v.GetTimestamp(),
			}
		}
	}

	return &domain.Trade{
		Id:               tradeRow.GetID(),
		Type:             domain.TradeType(tradeRow.GetType()),
		MarketName:       tradeRow.GetFkMarketName(),
		MarketBaseAsset:  tradeRow.GetBaseAsset(),
		MarketQuoteAsset: tradeRow.GetQuoteAsset(),
		MarketPrice: domain.MarketPrice{
			BasePrice:  basePrice,
			QuotePrice: quotePrice,
		},
		MarketPercentageFee: percentageFee,
		MarketFixedFee:      fixedFee,
		FeeAsset:            tradeRow.GetFeeAsset(),
		FeeAmount:           uint64(tradeRow.GetFeeAmount()),
		TraderPubkey:        tradeRow.GetTraderPubkey(),
		Status: domain.TradeStatus{
			Code:   statusCode,
			Failed: statusFailed,
		},
		PsetBase64:     tradeRow.GetPsetBase64(),
		TxId:           txId,
		TxHex:          tradeRow.GetTxHex(),
		ExpiryTime:     int64(expiryTime),
		SettlementTime: int64(settlementTime),
		SwapRequest:    swapRequest,
		SwapAccept:     swapAccept,
		SwapComplete:   swapComplete,
		SwapFail:       swapFail,
	}, nil
}

func parsePage(page domain.Page) (int32, int32) {
	limit := 500
	offset := 0
	if page != nil {
		limit = int(page.GetSize())
		offset = int(page.GetNumber()*page.GetSize() - page.GetSize())
	}

	return int32(limit), int32(offset)
}

type txRow interface {
	GetType() string
	GetAccountName() string
	GetTxID() string
	GetTimestamp() int64
	GetAsset() sql.NullString
	GetAmount() sql.NullInt64
}

type AllTransactionsForAccountNameAndPageRow struct {
	queries.GetAllTransactionsForAccountNameAndPageRow
}

func (a AllTransactionsForAccountNameAndPageRow) GetType() string {
	return a.Type
}

func (a AllTransactionsForAccountNameAndPageRow) GetAccountName() string {
	return a.AccountName
}

func (a AllTransactionsForAccountNameAndPageRow) GetTxID() string {
	return a.TxID
}

func (a AllTransactionsForAccountNameAndPageRow) GetTimestamp() int64 {
	return a.Timestamp
}

func (a AllTransactionsForAccountNameAndPageRow) GetAsset() sql.NullString {
	return a.Asset
}

func (a AllTransactionsForAccountNameAndPageRow) GetAmount() sql.NullInt64 {
	return a.Amount
}

type AllTransactionsRow struct {
	queries.GetAllTransactionsRow
}

func (a AllTransactionsRow) GetType() string {
	return a.Type
}

func (a AllTransactionsRow) GetAccountName() string {
	return a.AccountName
}

func (a AllTransactionsRow) GetTxID() string {
	return a.TxID
}

func (a AllTransactionsRow) GetTimestamp() int64 {
	return a.Timestamp
}

func (a AllTransactionsRow) GetAsset() sql.NullString {
	return a.Asset
}

func (a AllTransactionsRow) GetAmount() sql.NullInt64 {
	return a.Amount
}

type tx struct {
	accountName       string
	txid              string
	totAmountPerAsset map[string]uint64
	timestamp         int64
}

func convertTxsRowsToTxs(txsRows []txRow) []tx {
	txs := make([]tx, 0)
	newTxStartIndex := 0
	newTxEndIndex := 0

	for i, txRow := range txsRows {
		if i == 0 {
			newTxStartIndex = 0
		} else if txRow.GetTxID() != txsRows[i-1].GetTxID() {
			newTxEndIndex = i

			txs = append(
				txs, convertTxRowsToTx(txsRows[newTxStartIndex:newTxEndIndex]),
			)
			newTxStartIndex = i
		}
	}

	if len(txsRows) > 0 {
		txs = append(txs, convertTxRowsToTx(txsRows[newTxStartIndex:]))
	}

	return txs
}

func convertTxRowsToTx(txRows []txRow) tx {
	txRow := txRows[0]

	amountPerAsset := make(map[string]uint64)
	for _, v := range txRows {
		if v.GetAsset().Valid {
			amountPerAsset[v.GetAsset().String] = uint64(v.GetAmount().Int64)
		}
	}

	return tx{
		accountName:       txRow.GetAccountName(),
		txid:              txRow.GetTxID(),
		totAmountPerAsset: amountPerAsset,
		timestamp:         txRow.GetTimestamp(),
	}
}
