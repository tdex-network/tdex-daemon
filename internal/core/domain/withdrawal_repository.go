package domain

import "context"

// WithdrawalRepository is the abstraction to which all concrete implementations
// must sitck with to persist withdrawals.
type WithdrawalRepository interface {
	// AddWithdrawals adds the provided withdrawals to the repository. Those already
	// existing won't be re-added.
	AddWithdrawals(ctx context.Context, withdrawals []Withdrawal) (int, error)
	// ListWithdrawalsForAccount returns the list with the withdrawals related to
	// the given wallet account id.
	ListWithdrawalsForAccount(
		ctx context.Context, accountIndex int,
	) ([]Withdrawal, error)
	// ListWithdrawalsForAccountAndPage returns a page containing a subset of the
	// list with the withdrawals related to the given wallet account id.
	ListWithdrawalsForAccountAndPage(
		ctx context.Context, accountIndex int, page Page,
	) ([]Withdrawal, error)
	// ListAllWithdrawals returns all withdrawals related to all wallet accounts
	// stored in the repository.
	ListAllWithdrawals(ctx context.Context) ([]Withdrawal, error)
	// ListAllWithdrawalsForPage returns a page containing a subset of all
	// withdrawals related to all wallet accounts stored in the repository.
	ListAllWithdrawalsForPage(ctx context.Context, page Page) ([]Withdrawal, error)
}
