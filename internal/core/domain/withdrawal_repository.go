package domain

import "context"

// WithdrawalRepository is the abstraction for any kind of database intended to
// persist Withdrawals.
type WithdrawalRepository interface {
	// AddWithdrawals adds the provided withdrawals to the repository. Those already
	// existing won't be re-added.
	AddWithdrawals(ctx context.Context, withdrawals []Withdrawal) (int, error)
	// GetWithdrawalsForAccount returns the list with the withdrawals related to
	// the given market.
	GetWithdrawalsForAccount(
		ctx context.Context, accountName string, page Page,
	) ([]Withdrawal, error)
	// GetAllWithdrawals returns all withdrawals related to all markets.
	GetAllWithdrawals(ctx context.Context, page Page) ([]Withdrawal, error)
}
