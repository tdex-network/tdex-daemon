package domain

import "context"

type WithdrawalRepository interface {
	AddWithdrawal(ctx context.Context, withdrawal Withdrawal) error
	ListWithdrawalsForAccountId(
		ctx context.Context, accountIndex int, page *Page,
	) ([]Withdrawal, error)
	ListAllWithdrawals(ctx context.Context, page *Page) ([]Withdrawal, error)
}
