package postgresdb

import (
	"context"

	"github.com/jackc/pgconn"

	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/pg/sqlc/queries"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

const (
	depositTypeKey = "deposit"
)

type depositRepositoryImpl struct {
	querier *queries.Queries
	execTx  func(
		ctx context.Context,
		txBody func(*queries.Queries) error,
	) error
}

func NewDepositRepositoryImpl(
	querier *queries.Queries,
	execTx func(ctx context.Context, txBody func(*queries.Queries) error) error,
) domain.DepositRepository {
	return &depositRepositoryImpl{
		querier: querier,
		execTx:  execTx,
	}
}

func (d *depositRepositoryImpl) AddDeposits(
	ctx context.Context, deposits []domain.Deposit,
) (int, error) {
	count := 0
	for _, v := range deposits {
		txBody := func(q *queries.Queries) error {
			tx, err := q.InsertTransaction(
				ctx,
				queries.InsertTransactionParams{
					Type:        depositTypeKey,
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

		if err := d.execTx(ctx, txBody); err != nil {
			if err, ok := err.(*pgconn.PgError); ok && err.Code == uniqueViolation {
				continue
			}

			return -1, err
		}

		count++
	}

	return count, nil
}

func (d *depositRepositoryImpl) GetDepositsForAccount(
	ctx context.Context, accountName string, page domain.Page,
) ([]domain.Deposit, error) {
	limit, offset := parsePage(page)
	depositsRows, err := d.querier.GetAllTransactionsForAccountNameAndPage(
		ctx,
		queries.GetAllTransactionsForAccountNameAndPageParams{
			Type:        depositTypeKey,
			AccountName: accountName,
			Limit:       limit,
			Offset:      offset,
		})
	if err != nil {
		return nil, err
	}

	rows := make([]txRow, 0, len(depositsRows))
	for _, v := range depositsRows {
		rows = append(rows, AllTransactionsForAccountNameAndPageRow{v})
	}

	txs := convertTxsRowsToTxs(rows)
	deposits := make([]domain.Deposit, 0, len(txs))
	for _, v := range txs {
		deposits = append(deposits, domain.Deposit{
			AccountName:       v.accountName,
			TxID:              v.txid,
			TotAmountPerAsset: v.totAmountPerAsset,
			Timestamp:         v.timestamp,
		})
	}

	return deposits, nil
}

func (d *depositRepositoryImpl) GetAllDeposits(
	ctx context.Context, page domain.Page,
) ([]domain.Deposit, error) {
	limit, offset := parsePage(page)
	depositsRows, err := d.querier.GetAllTransactions(
		ctx,
		queries.GetAllTransactionsParams{
			Type:   depositTypeKey,
			Limit:  limit,
			Offset: offset,
		})
	if err != nil {
		return nil, err
	}

	rows := make([]txRow, 0, len(depositsRows))
	for _, v := range depositsRows {
		rows = append(rows, AllTransactionsRow{v})
	}

	txs := convertTxsRowsToTxs(rows)
	deposits := make([]domain.Deposit, 0, len(txs))
	for _, v := range txs {
		deposits = append(deposits, domain.Deposit{
			AccountName:       v.accountName,
			TxID:              v.txid,
			TotAmountPerAsset: v.totAmountPerAsset,
			Timestamp:         v.timestamp,
		})
	}

	return deposits, nil
}
