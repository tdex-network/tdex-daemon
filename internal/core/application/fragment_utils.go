package application

import (
	"encoding/hex"
	"errors"
	"sort"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
)

var (
	FragmentationMap = map[int]int{
		1: 30,
		2: 15,
		3: 10,
		5: 2,
	}
)

func fragmentAmountForFeeAccount(
	valueToBeFragmented, minFragmentValue uint64, maxNumOfFragments uint32,
) []uint64 {
	res := make([]uint64, 0)
	for i := uint32(0); valueToBeFragmented >= minFragmentValue && i < maxNumOfFragments; i++ {
		res = append(res, minFragmentValue)
		valueToBeFragmented -= minFragmentValue
	}
	if valueToBeFragmented > 0 {
		if len(res) > 0 {
			res[len(res)-1] += valueToBeFragmented
		} else {
			res = append(res, valueToBeFragmented)
		}
	}

	return res
}

func fragmentAmountsForMarket(assetValuePair pair) ([]uint64, []uint64) {
	baseAssetFragments := make([]uint64, 0)
	quoteAssetFragments := make([]uint64, 0)

	baseSum := uint64(0)
	quoteSum := uint64(0)
	for numOfUtxo, percentage := range FragmentationMap {
		for ; numOfUtxo > 0; numOfUtxo-- {
			if assetValuePair.baseValue > 0 {
				baseAssetPart := percent(int(assetValuePair.baseValue), percentage)
				baseSum += baseAssetPart
				baseAssetFragments = append(baseAssetFragments, baseAssetPart)
			}

			if assetValuePair.quoteValue > 0 {
				quoteAssetPart := percent(int(assetValuePair.quoteValue), percentage)
				quoteSum += quoteAssetPart
				quoteAssetFragments = append(quoteAssetFragments, quoteAssetPart)
			}
		}
	}

	sort.Slice(baseAssetFragments, func(i, j int) bool {
		return baseAssetFragments[i] < baseAssetFragments[j]
	})

	sort.Slice(quoteAssetFragments, func(i, j int) bool {
		return quoteAssetFragments[i] < quoteAssetFragments[j]
	})

	//if there is rest, created when calculating percentage,
	//add it to last fragment
	if baseSum != assetValuePair.baseValue {
		baseRest := assetValuePair.baseValue - baseSum
		if baseRest > 0 {
			baseAssetFragments[len(baseAssetFragments)-1] =
				baseAssetFragments[len(baseAssetFragments)-1] + baseRest
		}
	}

	//if there is rest, created when calculating percentage,
	//add it to last fragment
	if quoteSum != assetValuePair.quoteValue {
		quoteRest := assetValuePair.quoteValue - quoteSum
		if quoteRest > 0 {
			quoteAssetFragments[len(quoteAssetFragments)-1] =
				quoteAssetFragments[len(quoteAssetFragments)-1] + quoteRest
		}
	}

	return baseAssetFragments, quoteAssetFragments
}

func percent(num, percent int) uint64 {
	return decimal.NewFromInt(int64(num)).
		Mul(decimal.NewFromInt(int64(percent))).
		Div(decimal.NewFromInt(100)).BigInt().Uint64()
}

func estimateFees(numIns, numOuts int) uint64 {
	ins := make([]int, 0, numIns)
	for i := 0; i < numIns; i++ {
		ins = append(ins, wallet.P2WPKH)
	}

	outs := make([]int, 0, numOuts)
	for i := 0; i < numOuts; i++ {
		outs = append(outs, wallet.P2WPKH)
	}

	size := wallet.EstimateTxSize(ins, nil, nil, outs, nil)
	return uint64(float64(size) * 0.1)
}

func deductFeeFromFragments(fragments []uint64, feeAmount uint64) []uint64 {
	f := make([]uint64, len(fragments))
	copy(f, fragments)

	amountToPay := int64(feeAmount)
	for amountToPay > 0 {
		fLen := len(f) - 1
		lastFragment := int64(f[fLen])
		if amountToPay >= lastFragment {
			f = f[:fLen]
		} else {
			f[fLen] -= uint64(amountToPay)
		}
		amountToPay -= lastFragment
	}
	return f
}

func createOutpoints(txid string, numOuts int) []TxOutpoint {
	outpoints := make([]TxOutpoint, 0, numOuts)
	for i := 0; i < numOuts; i++ {
		outpoints = append(outpoints, TxOutpoint{
			Hash:  txid,
			Index: i,
		})
	}
	return outpoints
}

func craftTransaction(
	ephWallet *trade.Wallet, utxos []explorer.Utxo,
	baseFragments, quoteFragments []uint64,
	addresses []string, feeAmount uint64, assetValuePair pair, lbtc string,
) (string, error) {
	outs := createOutputs(
		baseFragments,
		quoteFragments,
		feeAmount,
		addresses,
		assetValuePair,
	)

	return buildFinalizedTx(ephWallet, utxos, outs, feeAmount, lbtc)
}

type pair struct {
	baseValue  uint64
	baseAsset  string
	quoteValue uint64
	quoteAsset string
}

func createOutputs(
	baseFragments, quoteFragments []uint64,
	feeAmount uint64,
	addresses []string,
	assetValuePair pair,
) []TxOut {
	outsLen := len(baseFragments) + len(quoteFragments)
	outputs := make([]TxOut, 0, outsLen)

	index := 0
	for _, v := range baseFragments {
		outputs = append(outputs, TxOut{
			Asset:   assetValuePair.baseAsset,
			Value:   int64(v),
			Address: addresses[index],
		})
		index++
	}
	for _, v := range quoteFragments {
		outputs = append(outputs, TxOut{
			Asset:   assetValuePair.quoteAsset,
			Value:   int64(v),
			Address: addresses[index],
		})
		index++
	}

	return outputs
}

func buildFinalizedTx(
	ephWallet *trade.Wallet, utxos []explorer.Utxo,
	outs []TxOut, feeAmount uint64, lbtc string,
) (string, error) {
	outputs, outputsBlindingKeys, err := parseRequestOutputs(outs)
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
		return "", errors.New("invalid signatures")
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
