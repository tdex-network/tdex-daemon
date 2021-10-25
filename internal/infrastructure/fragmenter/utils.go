package fragmenter

import (
	"encoding/hex"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
)

func percent(num, percent int) uint64 {
	return decimal.NewFromInt(int64(num)).
		Mul(decimal.NewFromInt(int64(percent))).
		Div(decimal.NewFromInt(100)).BigInt().Uint64()
}

func buildFinalizedTx(
	ephWallet *trade.Wallet, utxos []explorer.Utxo,
	outs []ports.TxOut, feeAmount uint64, lbtc string,
) (string, error) {
	outputs, outputsBlindingKeys, err := parseOutputs(outs)
	if err != nil {
		return "", err
	}

	ptx, err := pset.New(
		make([]*transaction.TxInput, 0, len(utxos)),
		make([]*transaction.TxOutput, 0, len(outputs)),
		2,
		0,
	)
	if err != nil {
		return "", err
	}

	ptx, err = addInsAndOutsToPset(ptx, utxos, outputs)
	if err != nil {
		return "", err
	}

	dataLen := len(utxos)
	inBlindData := make([]pset.BlindingDataLike, 0, dataLen)
	for _, u := range utxos {
		asset, _ := hex.DecodeString(u.Asset())
		inBlindData = append(inBlindData, pset.BlindingData{
			Value:               u.Value(),
			Asset:               elementsutil.ReverseBytes(asset),
			ValueBlindingFactor: u.ValueBlinder(),
			AssetBlindingFactor: u.AssetBlinder(),
		})
	}

	outBlindingKeysByIndex := make(map[int][]byte)
	for i, key := range outputsBlindingKeys {
		outBlindingKeysByIndex[i] = key
	}

	blinder, err := pset.NewBlinder(
		ptx,
		inBlindData,
		outBlindingKeysByIndex,
		nil,
		nil,
	)
	if err != nil {
		return "", err
	}

	err = blinder.Blind()
	if err != nil {
		return "", err
	}

	updater, err := pset.NewUpdater(ptx)
	if err != nil {
		return "", err
	}

	feeValue, _ := elementsutil.SatoshiToElementsValue(feeAmount)
	lbtcAsset, err := bufferutil.AssetHashToBytes(lbtc)
	if err != nil {
		return "", err
	}
	feeOutput := transaction.NewTxOutput(lbtcAsset, feeValue[:], []byte{})
	updater.AddOutput(feeOutput)

	ptxBase64, err := ptx.ToBase64()
	if err != nil {
		return "", err
	}

	signedPtxBase64, err := ephWallet.Sign(ptxBase64)
	if err != nil {
		return "", err
	}

	signedPtx, err := pset.NewPsetFromBase64(signedPtxBase64)
	if err != nil {
		return "", err
	}

	valid, err := signedPtx.ValidateAllSignatures()
	if err != nil {
		return "", err
	}

	if !valid {
		return "", fmt.Errorf("invalid signatures")
	}

	err = pset.FinalizeAll(signedPtx)
	if err != nil {
		return "", err
	}

	finalTx, err := pset.Extract(signedPtx)
	if err != nil {
		return "", err
	}

	txHex, err := finalTx.ToHex()
	if err != nil {
		return "", err
	}

	return txHex, nil
}

func addInsAndOutsToPset(
	ptx *pset.Pset,
	inputsToAdd []explorer.Utxo,
	outputsToAdd []*transaction.TxOutput,
) (*pset.Pset, error) {
	updater, err := pset.NewUpdater(ptx)
	if err != nil {
		return nil, err
	}

	for _, in := range inputsToAdd {
		input, witnessUtxo, _ := in.Parse()
		updater.AddInput(input)
		err := updater.AddInWitnessUtxo(witnessUtxo, len(ptx.Inputs)-1)
		if err != nil {
			return nil, err
		}
	}

	for _, out := range outputsToAdd {
		updater.AddOutput(out)
	}

	return ptx, nil
}

func parseOutputs(
	outs []ports.TxOut,
) ([]*transaction.TxOutput, [][]byte, error) {
	outputs := make([]*transaction.TxOutput, 0, len(outs))
	blindingKeys := make([][]byte, 0, len(outs))

	for _, out := range outs {
		asset, err := bufferutil.AssetHashToBytes(out.Asset())
		if err != nil {
			return nil, nil, err
		}
		value, err := bufferutil.ValueToBytes(uint64(out.Value()))
		if err != nil {
			return nil, nil, err
		}
		script, blindingKey, err := parseConfidentialAddress(out.Address())
		if err != nil {
			return nil, nil, err
		}

		output := transaction.NewTxOutput(asset, value, script)
		outputs = append(outputs, output)
		blindingKeys = append(blindingKeys, blindingKey)
	}
	return outputs, blindingKeys, nil
}

func parseConfidentialAddress(addr string) ([]byte, []byte, error) {
	script, err := address.ToOutputScript(addr)
	if err != nil {
		return nil, nil, err
	}
	ctAddr, err := address.FromConfidential(addr)
	if err != nil {
		return nil, nil, err
	}
	return script, ctAddr.BlindingKey, nil
}
