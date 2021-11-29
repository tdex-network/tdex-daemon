package application

import (
	"context"
	"strings"
	"time"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/circuitbreaker"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

func getAccountBalanceFromExplorer(
	repoManager ports.RepoManager, explorerSvc explorer.Service,
	ctx context.Context, accountIndex int,
) (map[string]BalanceInfo, error) {
	utxos, err := getAccountUtxosFromExplorer(
		repoManager, explorerSvc, ctx, accountIndex,
	)
	if err != nil {
		return nil, err
	}

	return getBalancesByAsset(utxos), nil
}

func getAccountUtxosFromExplorer(
	repoManager ports.RepoManager, explorerSvc explorer.Service,
	ctx context.Context, accountIndex int,
) ([]explorer.Utxo, error) {
	allInfo, err := repoManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(
		ctx, accountIndex,
	)
	if err != nil {
		return nil, err
	}

	addresses, keys := allInfo.AddressesAndKeys()

	cb := circuitbreaker.NewCircuitBreaker()
	iUtxos, err := cb.Execute(func() (interface{}, error) {
		return explorerSvc.GetUnspentsForAddresses(addresses, keys)
	})
	if err != nil {
		return nil, err
	}
	return iUtxos.([]explorer.Utxo), nil
}

func getBalancesByAsset(unspents []explorer.Utxo) map[string]BalanceInfo {
	balances := map[string]BalanceInfo{}
	for _, unspent := range unspents {
		if _, ok := balances[unspent.Asset()]; !ok {
			balances[unspent.Asset()] = BalanceInfo{}
		}

		balance := balances[unspent.Asset()]
		balance.TotalBalance += unspent.Value()
		if unspent.IsConfirmed() {
			balance.ConfirmedBalance += unspent.Value()
		} else {
			balance.UnconfirmedBalance += unspent.Value()
		}
		balances[unspent.Asset()] = balance
	}
	return balances
}

func waitForTx(explorerSvc explorer.Service, txid string) error {
	waitingTime := 1 * time.Second
	for {
		_, err := explorerSvc.GetTransaction(txid)
		if err != nil {
			if strings.Contains(err.Error(), "Transaction not found") {
				time.Sleep(waitingTime)
				continue
			}
			return err
		}
		return nil
	}
}
