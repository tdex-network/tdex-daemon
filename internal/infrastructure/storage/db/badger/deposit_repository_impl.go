package dbbadger

import (
	"context"

	"github.com/dgraph-io/badger/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v2"
)

type depositRepositoryImpl struct {
	store *badgerhold.Store
}

// NewDepositRepositoryImpl initialize a badger implementation of the domain.StatsRepository
func NewDepositRepositoryImpl(store *badgerhold.Store) domain.DepositRepository {
	return depositRepositoryImpl{store}
}

func (d depositRepositoryImpl) AddDeposit(
	ctx context.Context,
	deposit domain.Deposit,
) error {
	return d.insertDeposit(ctx, deposit)
}

func (d depositRepositoryImpl) ListDepositsForAccountIdAndPage(
	ctx context.Context,
	accountIndex int,
	page domain.Page,
) ([]domain.Deposit, error) {
	query := badgerhold.Where("AccountIndex").Eq(accountIndex)
	var deposits []domain.Deposit

	from := page.Number*page.Size - page.Size

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		if err := d.store.TxFind(
			tx, &deposits,
			query.Skip(from).Limit(page.Size),
		); err != nil {
			return nil, err
		}
	} else {
		if err := d.store.Find(
			&deposits,
			query.Skip(from).Limit(page.Size),
		); err != nil {
			return nil, err
		}
	}

	return deposits, nil
}

func (d depositRepositoryImpl) insertDeposit(
	ctx context.Context,
	deposit domain.Deposit,
) error {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = d.store.TxInsert(tx, deposit.Key(), &deposit)
	} else {
		err = d.store.Insert(deposit.Key(), &deposit)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}

func (d depositRepositoryImpl) ListAllDeposits(
	ctx context.Context,
) ([]domain.Deposit, error) {
	return d.listAll(ctx), nil
}

func (d depositRepositoryImpl) listAll(ctx context.Context) []domain.Deposit {
	deposits, _ := d.findDeposits(ctx, nil)
	return deposits
}

func (d depositRepositoryImpl) findDeposits(
	ctx context.Context,
	query *badgerhold.Query,
) ([]domain.Deposit, error) {
	var deposits []domain.Deposit
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = d.store.TxFind(tx, &deposits, query)
	} else {
		err = d.store.Find(&deposits, query)
	}
	if err != nil {
		return nil, err
	}

	return deposits, nil
}
