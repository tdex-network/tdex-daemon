package domain

import "context"

type Page struct {
	Number int
	Size   int
}

type StatsRepository interface {
	AddWithdrawal(ctx context.Context, withdrawal Withdrawal) error
	ListWithdrawalsForAccountIdAndPage(
		ctx context.Context,
		accountIndex int,
		page Page,
	) ([]Withdrawal, error)
	AddDeposit(ctx context.Context, deposit Deposit) error
	ListDepositsForAccountIdAndPage(
		ctx context.Context,
		accountIndex int,
		page Page,
	) ([]Deposit, error)
}
