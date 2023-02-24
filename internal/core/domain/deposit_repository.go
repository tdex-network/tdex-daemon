package domain

import "context"

// DepositRepository is the abstraction for any kind of database intended to
// persist Deposits.
type DepositRepository interface {
	// AddDeposits adds the provided deposits to the repository. Those already
	// existing won't be re-added.
	AddDeposits(ctx context.Context, deposits []Deposit) (int, error)
	// GetDepositsForAccount returns the deposits related to the given account.
	GetDepositsForAccount(
		ctx context.Context, accountName string, page Page,
	) ([]Deposit, error)
	// GetAllDeposits returns all deposits related to all markets.
	GetAllDeposits(ctx context.Context, page Page) ([]Deposit, error)
}
