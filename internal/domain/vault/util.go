package vault

import (
	"errors"
	"fmt"

	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/transaction"
)

func validateAccountIndex(accIndex int) error {
	if accIndex < 0 {
		return errors.New("Account index must be a positive integer number")
	}

	return nil
}

func getOutputsAssets(outputs []*transaction.TxOutput) []string {
	assets := make([]string, 0)
	for _, out := range outputs {
		asset := bufferutil.AssetHashFromBytes(out.Asset)
		if !containsAsset(assets, asset) {
			assets = append(assets, asset)
		}
	}
	return assets
}

func containsAsset(assets []string, asset string) bool {
	for _, a := range assets {
		if a == asset {
			return true
		}
	}
	return false
}

func createFeeOutput(feeAmount uint64) []*transaction.TxOutput {
	feeAsset, _ := bufferutil.AssetHashToBytes(config.GetNetwork().AssetID)
	feeValue, _ := bufferutil.ValueToBytes(feeAmount)
	feeScript := make([]byte, 0)
	return []*transaction.TxOutput{
		transaction.NewTxOutput(feeAsset, feeValue, feeScript),
	}
}

func deriveAddressesInRange(
	w *wallet.Wallet,
	accountIndex,
	chainIndex,
	firstAddressIndex,
	lastAddressIndex int,
) []string {
	addresses := make([]string, 0)
	for i := firstAddressIndex; i <= lastAddressIndex; i++ {
		derivationPath := fmt.Sprintf("%d'/%d/%d", accountIndex, chainIndex, i)
		addr, _, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
			DerivationPath: derivationPath,
			Network:        config.GetNetwork(),
		})
		addresses = append(addresses, addr)
	}
	return addresses
}
