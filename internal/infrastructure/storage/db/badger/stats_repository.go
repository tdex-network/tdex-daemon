package dbbadger

import (
	"context"

	"github.com/dgraph-io/badger/v2"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v2"
)

type statsRepositoryImpl struct {
	store *badgerhold.Store
}

// NewStatsRepositoryImpl initialize a badger implementation of the domain.StatsRepository
func NewStatsRepositoryImpl(store *badgerhold.Store) domain.StatsRepository {
	return statsRepositoryImpl{store}
}

func (s statsRepositoryImpl) AddWithdrawal(
	ctx context.Context,
	withdrawal domain.Withdrawal,
) error {
	return s.insertWithdrawal(ctx, withdrawal)
}

func (s statsRepositoryImpl) ListWithdrawalsForAccountIdAndPage(
	ctx context.Context,
	accountIndex int,
	page domain.Page,
) ([]domain.Withdrawal, error) {
	query := badgerhold.Where("AccountIndex").Eq(accountIndex)
	var withdrawals []domain.Withdrawal

	from := page.Number*page.Size - page.Size + 1

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		if err := s.store.TxFind(
			tx, &withdrawals,
			query.Skip(from).Limit(page.Size),
		); err != nil {
			return nil, err
		}
	} else {
		if err := s.store.Find(
			&withdrawals,
			query.Skip(from).Limit(page.Size),
		); err != nil {
			return nil, err
		}
	}

	return withdrawals, nil
}

func (s statsRepositoryImpl) AddDeposit(
	ctx context.Context,
	deposit domain.Deposit,
) error {
	return s.insertDeposit(ctx, deposit)
}

func (s statsRepositoryImpl) ListDepositsForAccountIdAndPage(
	ctx context.Context,
	accountIndex int,
	page domain.Page,
) ([]domain.Deposit, error) {
	query := badgerhold.Where("AccountIndex").Eq(accountIndex)
	var deposits []domain.Deposit

	from := page.Number*page.Size - page.Size + 1

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		if err := s.store.TxFind(
			tx, &deposits,
			query.Skip(from).Limit(page.Size),
		); err != nil {
			return nil, err
		}
	} else {
		if err := s.store.Find(
			&deposits,
			query.Skip(from).Limit(page.Size),
		); err != nil {
			return nil, err
		}
	}

	return deposits, nil
}

func (s statsRepositoryImpl) insertWithdrawal(
	ctx context.Context,
	withdrawal domain.Withdrawal,
) error {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = s.store.TxInsert(tx, badgerhold.NextSequence(), &withdrawal)
	} else {
		err = s.store.Insert(badgerhold.NextSequence(), &withdrawal)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}

func (s statsRepositoryImpl) insertDeposit(
	ctx context.Context,
	deposit domain.Deposit,
) error {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = s.store.TxInsert(tx, deposit.Key(), &deposit)
	} else {
		err = s.store.Insert(deposit.Key(), &deposit)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}
