package domain

import "context"

type DepositRepository interface {
	AddDeposit(ctx context.Context, deposit Deposit) error
	ListDepositsForAccountId(
		ctx context.Context, accountIndex int, page *Page,
	) ([]Deposit, error)
	ListAllDeposits(ctx context.Context, page *Page) ([]Deposit, error)
}
