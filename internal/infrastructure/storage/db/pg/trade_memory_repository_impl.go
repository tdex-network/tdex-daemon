package postgresdb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/pg/sqlc/queries"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

const (
	swapAcceptKey   = "accept"
	swapCompleteKey = "complete"
	swapFailKey     = "fail"
	swapRequestKey  = "request"
)

type tradeRepositoryImpl struct {
	querier *queries.Queries
	execTx  func(
		ctx context.Context,
		txBody func(*queries.Queries) error,
	) error
}

func NewTradeRepositoryImpl(
	querier *queries.Queries,
	execTx func(ctx context.Context, txBody func(*queries.Queries) error) error,
) domain.TradeRepository {
	return &tradeRepositoryImpl{
		querier: querier,
		execTx:  execTx,
	}
}

func (t *tradeRepositoryImpl) AddTrade(
	ctx context.Context, trade *domain.Trade,
) error {
	txBody := func(querierWithTx *queries.Queries) error {
		statusCode := sql.NullInt32{}
		if trade.Status.Code > 0 {
			statusCode.Int32 = int32(trade.Status.Code)
			statusCode.Valid = true
		}

		txId := sql.NullString{}
		if trade.TxId != "" {
			txId.String = trade.TxId
			txId.Valid = true
		}

		expiryTime := sql.NullInt64{}
		if trade.ExpiryTime > 0 {
			expiryTime.Int64 = trade.ExpiryTime
			expiryTime.Valid = true
		}

		settlementTime := sql.NullInt64{}
		if trade.SettlementTime > 0 {
			settlementTime.Int64 = trade.SettlementTime
			settlementTime.Valid = true
		}

		if _, err := querierWithTx.InsertTrade(ctx, queries.InsertTradeParams{
			ID:           trade.Id,
			Type:         int32(trade.Type),
			FeeAsset:     trade.FeeAsset,
			FeeAmount:    int64(trade.FeeAmount),
			TraderPubkey: trade.TraderPubkey,
			StatusCode:   statusCode,
			StatusFailed: sql.NullBool{
				Bool:  trade.Status.Failed,
				Valid: true,
			},
			PsetBase64:     trade.PsetBase64,
			TxID:           txId,
			TxHex:          trade.TxHex,
			ExpiryTime:     expiryTime,
			SettlementTime: settlementTime,
			FkMarketName:   trade.MarketName,
		}); err != nil {
			return err
		}

		return t.insertTradeSwaps(ctx, querierWithTx, *trade)
	}

	return t.execTx(ctx, txBody)
}

func (t *tradeRepositoryImpl) insertTradeSwaps(
	ctx context.Context, querierWithTx *queries.Queries, trade domain.Trade,
) error {

	if trade.SwapAccept != nil {
		if _, err := querierWithTx.InsertSwap(ctx, queries.InsertSwapParams{
			ID:        trade.SwapAccept.Id,
			Message:   trade.SwapAccept.Message,
			Timestamp: trade.SwapAccept.Timestamp,
			Type:      swapAcceptKey,
			FkTradeID: trade.Id,
		}); err != nil {
			return err
		}
	}

	if trade.SwapComplete != nil {
		if _, err := querierWithTx.InsertSwap(ctx, queries.InsertSwapParams{
			ID:        trade.SwapComplete.Id,
			Message:   trade.SwapComplete.Message,
			Timestamp: trade.SwapComplete.Timestamp,
			Type:      swapCompleteKey,
			FkTradeID: trade.Id,
		}); err != nil {
			return err
		}
	}

	if trade.SwapRequest != nil {
		if _, err := querierWithTx.InsertSwap(ctx, queries.InsertSwapParams{
			ID:        trade.SwapRequest.Id,
			Message:   trade.SwapRequest.Message,
			Timestamp: trade.SwapRequest.Timestamp,
			Type:      swapRequestKey,
			FkTradeID: trade.Id,
		}); err != nil {
			return err
		}
	}

	if trade.SwapFail != nil {
		if _, err := querierWithTx.InsertSwap(ctx, queries.InsertSwapParams{
			ID:        trade.SwapFail.Id,
			Message:   trade.SwapFail.Message,
			Timestamp: trade.SwapFail.Timestamp,
			Type:      swapFailKey,
			FkTradeID: trade.Id,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (t *tradeRepositoryImpl) GetTradeById(
	ctx context.Context, id string,
) (*domain.Trade, error) {
	tradeRows, err := t.querier.GetTradeById(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(tradeRows) == 0 {
		return nil, fmt.Errorf("trade with id %s not found", id)
	}

	rows := make([]tradeRow, 0, len(tradeRows))
	for _, v := range tradeRows {
		rows = append(rows, TradeByIdRow{v})
	}

	trade, err := convertTradeRowsToTrade(rows)
	if err != nil {
		return nil, err
	}

	return trade, nil

}

func (t *tradeRepositoryImpl) GetAllTrades(
	ctx context.Context, page domain.Page,
) ([]domain.Trade, error) {
	limit, offset := parsePage(page)
	tradesRows, err := t.querier.GetAllTrades(ctx, queries.GetAllTradesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	if len(tradesRows) == 0 {
		return nil, nil
	}

	rows := make([]tradeRow, 0, len(tradesRows))
	for _, v := range tradesRows {
		rows = append(rows, AllTradesRow{v})
	}

	trades, err := convertTradesRowsToTrades(rows)
	if err != nil {
		return nil, err
	}

	return trades, nil
}

func (t *tradeRepositoryImpl) GetAllTradesByMarket(
	ctx context.Context, marketName string, page domain.Page,
) ([]domain.Trade, error) {
	limit, offset := parsePage(page)
	tradesRows, err := t.querier.GetAllTradesByMarket(
		ctx,
		queries.GetAllTradesByMarketParams{
			Name:   marketName,
			Limit:  limit,
			Offset: offset,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(tradesRows) <= 0 {
		return nil, nil
	}

	rows := make([]tradeRow, 0, len(tradesRows))
	for _, v := range tradesRows {
		rows = append(rows, AllTradesByMarketRow{v})
	}

	trades, err := convertTradesRowsToTrades(rows)
	if err != nil {
		return nil, err
	}

	return trades, nil
}

func (t *tradeRepositoryImpl) GetCompletedTradesByMarket(
	ctx context.Context, marketName string, page domain.Page,
) ([]domain.Trade, error) {
	limit, offset := parsePage(page)
	tradesRows, err := t.querier.GetTradesByMarketAndStatus(
		ctx,
		queries.GetTradesByMarketAndStatusParams{
			Name: marketName,
			StatusCode: sql.NullInt32{
				Int32: int32(domain.TradeStatusCodeCompleted),
				Valid: true,
			},
			StatusFailed: sql.NullBool{
				Bool:  false,
				Valid: true,
			},
			Limit:  limit,
			Offset: offset,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(tradesRows) <= 0 {
		return nil, nil
	}

	rows := make([]tradeRow, 0, len(tradesRows))
	for _, v := range tradesRows {
		rows = append(rows, TradesByMarketAndStatusRow{v})
	}

	trades, err := convertTradesRowsToTrades(rows)
	if err != nil {
		return nil, err
	}

	return trades, nil
}

func (t *tradeRepositoryImpl) GetTradeBySwapAcceptId(
	ctx context.Context, swapAcceptId string,
) (*domain.Trade, error) {
	tradeRows, err := t.querier.GetTradeBySwapAcceptId(ctx, swapAcceptId)
	if err != nil {
		return nil, err
	}

	if len(tradeRows) == 0 {
		return nil, fmt.Errorf(
			"trade with swap accept id %s not found", swapAcceptId,
		)
	}

	rows := make([]tradeRow, 0, len(tradeRows))
	for _, v := range tradeRows {
		rows = append(rows, GetTradeBySwapAcceptIdRow{v})
	}

	trade, err := convertTradeRowsToTrade(rows)
	if err != nil {
		return nil, err
	}

	return trade, nil
}

func (t *tradeRepositoryImpl) GetTradeByTxId(
	ctx context.Context, txid string,
) (*domain.Trade, error) {
	txId := sql.NullString{
		String: txid,
		Valid:  true,
	}
	tradeRows, err := t.querier.GetTradeByTxId(ctx, txId)
	if err != nil {
		return nil, err
	}

	if len(tradeRows) == 0 {
		return nil, nil
	}

	rows := make([]tradeRow, 0, len(tradeRows))
	for _, v := range tradeRows {
		rows = append(rows, GetTradeByTxIdRow{v})
	}

	trade, err := convertTradeRowsToTrade(rows)
	if err != nil {
		return nil, err
	}

	return trade, nil
}

func (t *tradeRepositoryImpl) UpdateTrade(
	ctx context.Context, tradeId string,
	updateFn func(t *domain.Trade) (*domain.Trade, error),
) error {
	currentTrade, err := t.GetTradeById(ctx, tradeId)
	if err != nil {
		return err
	}

	updatedTrade, err := updateFn(currentTrade)
	if err != nil {
		return err
	}

	txBody := func(querierWithTx *queries.Queries) error {
		if _, err := querierWithTx.UpdateTrade(ctx, queries.UpdateTradeParams{
			Type:         int32(updatedTrade.Type),
			FeeAsset:     updatedTrade.FeeAsset,
			FeeAmount:    int64(updatedTrade.FeeAmount),
			TraderPubkey: updatedTrade.TraderPubkey,
			StatusCode: sql.NullInt32{
				Int32: int32(updatedTrade.Status.Code),
				Valid: true,
			},
			StatusFailed: sql.NullBool{
				Bool:  updatedTrade.Status.Failed,
				Valid: true,
			},
			PsetBase64: updatedTrade.PsetBase64,
			TxID: sql.NullString{
				String: updatedTrade.TxId,
				Valid:  true,
			},
			TxHex: updatedTrade.TxHex,
			ExpiryTime: sql.NullInt64{
				Int64: updatedTrade.ExpiryTime,
				Valid: true,
			},
			SettlementTime: sql.NullInt64{
				Int64: updatedTrade.SettlementTime,
				Valid: true,
			},
			ID: updatedTrade.Id,
		}); err != nil {
			return err
		}

		if err := querierWithTx.DeleteSwapsByTradeId(ctx, updatedTrade.Id); err != nil {
			return err
		}

		return t.insertTradeSwaps(ctx, querierWithTx, *updatedTrade)
	}

	return t.execTx(ctx, txBody)
}
