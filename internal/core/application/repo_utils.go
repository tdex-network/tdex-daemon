package application

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

// getUnlockedBalanceForMarket helper to get the available balance of a market
// from the repositories of a repoManager.
func getUnlockedBalanceForMarket(
	repoManager ports.RepoManager, ctx context.Context, mkt *domain.Market,
) (*Balance, error) {
	return getBalanceForMarket(repoManager, ctx, mkt, true)
}

// getUnlockedBalanceForFee helper to get the available balance of the fee
// account from the repositories of a repoManager.
func getUnlockedBalanceForFee(
	repoManager ports.RepoManager, ctx context.Context, lbtcAsset string,
) (uint64, error) {
	return getBalanceForFee(repoManager, ctx, lbtcAsset, true)
}

// getBalanceForMarket helper to get the balance or "unlocked" balance of a
// market account.
func getBalanceForMarket(
	repoManager ports.RepoManager,
	ctx context.Context, mkt *domain.Market, wantUnlocked bool,
) (*Balance, error) {
	info, err := repoManager.VaultRepository().
		GetAllDerivedAddressesInfoForAccount(ctx, mkt.AccountIndex)
	if err != nil {
		return nil, err
	}
	addresses := info.Addresses()

	getBalance := repoManager.UnspentRepository().GetBalance
	if wantUnlocked {
		getBalance = repoManager.UnspentRepository().GetUnlockedBalance
	}

	baseBalance, err := getBalance(ctx, addresses, mkt.BaseAsset)
	if err != nil {
		return nil, err
	}

	quoteBalance, err := getBalance(ctx, addresses, mkt.QuoteAsset)
	if err != nil {
		return nil, err
	}

	return &Balance{
		BaseAmount:  baseBalance,
		QuoteAmount: quoteBalance,
	}, nil
}

// getBalanceForFee helper to get the balance or "unlocked" balance of the fee
// account.
func getBalanceForFee(
	repoManager ports.RepoManager,
	ctx context.Context, lbtcAsset string, wantUnlocked bool,
) (uint64, error) {
	info, err := repoManager.VaultRepository().
		GetAllDerivedAddressesInfoForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return 0, err
	}
	addresses := info.Addresses()

	getBalance := repoManager.UnspentRepository().GetBalance
	if wantUnlocked {
		getBalance = repoManager.UnspentRepository().GetUnlockedBalance
	}

	return getBalance(ctx, addresses, lbtcAsset)
}
