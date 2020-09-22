package walletservice

import (
	"encoding/hex"
	"fmt"

	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/transaction"
)

func parseConfidentialAddress(addr string) ([]byte, []byte, error) {
	script, err := address.ToOutputScript(addr, *config.GetNetwork())
	if err != nil {
		return nil, nil, err
	}
	blindingKey, err := extractBlindingKey(addr, script)
	if err != nil {
		return nil, nil, err
	}
	return script, blindingKey, nil
}

func extractBlindingKey(addr string, script []byte) ([]byte, error) {
	addrType, _ := address.DecodeType(addr, *config.GetNetwork())
	switch addrType {
	case address.ConfidentialP2Pkh, address.ConfidentialP2Sh:
		decoded, _ := address.FromBase58(addr)
		return decoded.Data[1:34], nil
	case address.ConfidentialP2Wpkh, address.ConfidentialP2Wsh:
		decoded, _ := address.FromBlech32(addr)
		return decoded.PublicKey, nil
	default:
		return nil, fmt.Errorf("failed to extract blinding key from address '%s'", addr)
	}
}

func getAssetsOfOutputs(outputs []*transaction.TxOutput) []string {
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

func sendToMany(
	mnemonic []string,
	account *vault.Account,
	unspents []explorer.Utxo,
	outputs []*transaction.TxOutput,
	outputsBlindingKeys [][]byte,
	milliSatsPerBytes int,
	changePathsByAsset map[string]string,
) (string, error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	})
	if err != nil {
		return "", err
	}

	newPset, err := w.CreateTx()
	if err != nil {
		return "", err
	}
	updateResult, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64:         newPset,
		Unspents:           unspents,
		Outputs:            outputs,
		ChangePathsByAsset: changePathsByAsset,
		MilliSatsPerBytes:  milliSatsPerBytes,
		Network:            config.GetNetwork(),
	})
	if err != nil {
		return "", err
	}

	changeOutputsBlindingKeys := make([][]byte, 0, len(updateResult.ChangeOutputsBlindingKeys))
	for _, v := range updateResult.ChangeOutputsBlindingKeys {
		changeOutputsBlindingKeys = append(changeOutputsBlindingKeys, v)
	}
	outputsPlusChangesBlindingKeys := append(
		outputsBlindingKeys,
		changeOutputsBlindingKeys...,
	)
	blindedPset, err := w.BlindTransaction(wallet.BlindTransactionOpts{
		PsetBase64:         updateResult.PsetBase64,
		OutputBlindingKeys: outputsPlusChangesBlindingKeys,
	})
	if err != nil {
		return "", err
	}
	blindedPlusFees, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64: blindedPset,
		Outputs:    transactionutil.NewFeeOutput(updateResult.FeeAmount),
		Network:    config.GetNetwork(),
	})
	if err != nil {
		return "", err
	}

	inputPathsByScript := getDerivationPathsForUnspents(account, unspents)
	signedPset, err := w.SignTransaction(wallet.SignTransactionOpts{
		PsetBase64:        blindedPlusFees.PsetBase64,
		DerivationPathMap: inputPathsByScript,
	})
	if err != nil {
		return "", err
	}

	return wallet.FinalizeAndExtractTransaction(
		wallet.FinalizeAndExtractTransactionOpts{
			PsetBase64: signedPset,
		},
	)
}

func getDerivationPathsForUnspents(
	account *vault.Account,
	unspents []explorer.Utxo,
) map[string]string {
	paths := map[string]string{}
	for _, unspent := range unspents {
		script := hex.EncodeToString(unspent.Script())
		derivationPath, _ := account.DerivationPathByScript(script)
		paths[script] = derivationPath
	}
	return paths
}
