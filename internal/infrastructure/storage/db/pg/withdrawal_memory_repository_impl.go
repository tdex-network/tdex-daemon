package postgresdb

import (
	"context"

	"github.com/jackc/pgconn"

	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/pg/sqlc/queries"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

const (
	withdrawalTypeKey = "withdrawal"
)

type withdrawalRepositoryImpl struct {
	querier *queries.Queries
	execTx  func(
		ctx context.Context,
		txBody func(*queries.Queries) error,
	) error
}

func NewWithdrawalRepositoryImpl(
	querier *queries.Queries,
	execTx func(ctx context.Context, txBody func(*queries.Queries) error) error,
) domain.WithdrawalRepository {
	return &withdrawalRepositoryImpl{
		querier: querier,
		execTx:  execTx,
	}
}

func (w *withdrawalRepositoryImpl) AddWithdrawals(
	ctx context.Context, withdrawals []domain.Withdrawal,
) (int, error) {
	count := 0
	for _, v := range withdrawals {
		txBody := func(q *queries.Queries) error {
			tx, err := q.InsertTransaction(
				ctx,
				queries.InsertTransactionParams{
					Type:        withdrawalTypeKey,
					AccountName: v.AccountName,
					TxID:        v.TxID,
					Timestamp:   v.Timestamp,
				},
			)
			if err != nil {
				return err
			}

			for asset, amount := range v.TotAmountPerAsset {
				if _, err := q.InsertTransactionAssetAmount(
					ctx,
					queries.InsertTransactionAssetAmountParams{
						FkTransactionID: tx.ID,
						Asset:           asset,
						Amount:          int64(amount),
					},
				); err != nil {
					return err
				}
			}
			return nil
		}

		if err := w.execTx(ctx, txBody); err != nil {
			if err, ok := err.(*pgconn.PgError); ok && err.Code == uniqueViolation {
				continue
			}

			return -1, err
		}

		count++
	}

	return count, nil
}

func (w *withdrawalRepositoryImpl) GetWithdrawalsForAccount(
	ctx context.Context, accountName string, page domain.Page,
) ([]domain.Withdrawal, error) {
	limit, offset := parsePage(page)
	withdrawalsRows, err := w.querier.GetAllTransactionsForAccountNameAndPage(
		ctx,
		queries.GetAllTransactionsForAccountNameAndPageParams{
			Type:        withdrawalTypeKey,
			AccountName: accountName,
			Limit:       limit,
			Offset:      offset,
		})
	if err != nil {
		return nil, err
	}

	rows := make([]txRow, 0, len(withdrawalsRows))
	for _, v := range withdrawalsRows {
		rows = append(rows, AllTransactionsForAccountNameAndPageRow{v})
	}

	txs := convertTxsRowsToTxs(rows)
	withdrawals := make([]domain.Withdrawal, 0, len(txs))
	for _, v := range txs {
		withdrawals = append(withdrawals, domain.Withdrawal{
			AccountName:       v.accountName,
			TxID:              v.txid,
			TotAmountPerAsset: v.totAmountPerAsset,
			Timestamp:         v.timestamp,
		})
	}

	return withdrawals, nil
}

func (w *withdrawalRepositoryImpl) GetAllWithdrawals(
	ctx context.Context, page domain.Page,
) ([]domain.Withdrawal, error) {
	limit, offset := parsePage(page)
	withdrawalsRows, err := w.querier.GetAllTransactions(
		ctx,
		queries.GetAllTransactionsParams{
			Type:   withdrawalTypeKey,
			Limit:  limit,
			Offset: offset,
		})
	if err != nil {
		return nil, err
	}

	rows := make([]txRow, 0, len(withdrawalsRows))
	for _, v := range withdrawalsRows {
		rows = append(rows, AllTransactionsRow{v})
	}

	txs := convertTxsRowsToTxs(rows)
	withdrawals := make([]domain.Withdrawal, 0, len(txs))
	for _, v := range txs {
		withdrawals = append(withdrawals, domain.Withdrawal{
			AccountName:       v.accountName,
			TxID:              v.txid,
			TotAmountPerAsset: v.totAmountPerAsset,
			Timestamp:         v.timestamp,
		})
	}

	return withdrawals, nil
}
