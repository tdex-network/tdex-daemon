package application

import (
	"sort"

	"github.com/shopspring/decimal"
)

type pair struct {
	baseAsset  string
	quoteAsset string
	baseValue  uint64
	quoteValue uint64
}

func feeFragmentAmount(
	valueToBeFragmented, fragmentValue uint64, maxNumOfFragments uint32,
) []uint64 {
	res := make([]uint64, 0)
	for i := uint32(0); valueToBeFragmented >= fragmentValue && i < maxNumOfFragments; i++ {
		res = append(res, fragmentValue)
		valueToBeFragmented -= fragmentValue
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

func marketFragmentAmount(
	assetValuePair pair, fragmentationMap map[int]int,
) (baseFragments, quoteFragments []uint64, err error) {
	baseSum := uint64(0)
	quoteSum := uint64(0)
	for numOfUtxo, percentage := range fragmentationMap {
		for ; numOfUtxo > 0; numOfUtxo-- {
			if assetValuePair.baseValue > 0 {
				baseAssetPart := percent(int(assetValuePair.baseValue), percentage)
				baseSum += baseAssetPart
				baseFragments = append(baseFragments, baseAssetPart)
			}

			if assetValuePair.quoteValue > 0 {
				quoteAssetPart := percent(int(assetValuePair.quoteValue), percentage)
				quoteSum += quoteAssetPart
				quoteFragments = append(quoteFragments, quoteAssetPart)
			}
		}
	}

	sort.Slice(baseFragments, func(i, j int) bool {
		return baseFragments[i] < baseFragments[j]
	})

	sort.Slice(quoteFragments, func(i, j int) bool {
		return quoteFragments[i] < quoteFragments[j]
	})

	// if there is rest, created when calculating percentage,
	// add it to last fragment
	if baseSum != assetValuePair.baseValue {
		baseRest := assetValuePair.baseValue - baseSum
		if baseRest > 0 {
			baseFragments[len(baseFragments)-1] =
				baseFragments[len(baseFragments)-1] + baseRest
		}
	}

	// if there is rest, created when calculating percentage,
	// add it to last fragment
	if quoteSum != assetValuePair.quoteValue {
		quoteRest := assetValuePair.quoteValue - quoteSum
		if quoteRest > 0 {
			quoteFragments[len(quoteFragments)-1] =
				quoteFragments[len(quoteFragments)-1] + quoteRest
		}
	}

	return
}

func percent(num, percent int) uint64 {
	return decimal.NewFromInt(int64(num)).
		Mul(decimal.NewFromInt(int64(percent))).
		Div(decimal.NewFromInt(100)).BigInt().Uint64()
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

func createOutputs(
	baseFragments, quoteFragments []uint64,
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
