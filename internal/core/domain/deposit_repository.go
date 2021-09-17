package domain

import "context"

// DepositRepository is the abstraction to which all concrete implementations
// must sitck with to persist deposits.
type DepositRepository interface {
	// AddDeposits adds the provided deposits to the repository. Those already
	// existing won't be re-added.
	AddDeposits(ctx context.Context, deposits []Deposit) (int, error)
	// ListDepositsForAccount returns the list of deposits related to the given
	// wallet account id.
	ListDepositsForAccount(
		ctx context.Context, accountIndex int,
	) ([]Deposit, error)
	// ListDepositsForAccountAndPage returns a page containing a subset of the
	// list of deposits related to the given wallet account id.
	ListDepositsForAccountAndPage(
		ctx context.Context, accountIndex int, page Page,
	) ([]Deposit, error)
	// ListAllDeposits returns all deposits related to all wallet accounts stored
	// in the repository.
	ListAllDeposits(ctx context.Context) ([]Deposit, error)
	// ListAllDepositsForPage returns a page containing a subset of deposits
	// related to all wallet accounts stored in the repository.
	ListAllDepositsForPage(ctx context.Context, page Page) ([]Deposit, error)
}
