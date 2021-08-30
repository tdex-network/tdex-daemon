package domain

import "context"

type WithdrawalRepository interface {
	AddWithdrawal(ctx context.Context, withdrawal Withdrawal) error
	ListWithdrawalsForAccountIdAndPage(
		ctx context.Context,
		accountIndex int,
		page Page,
	) ([]Withdrawal, error)
}
