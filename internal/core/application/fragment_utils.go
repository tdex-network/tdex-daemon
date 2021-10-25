package application

import (
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
)

type pair struct {
	baseAsset  string
	quoteAsset string
	baseValue  uint64
	quoteValue uint64
}

func (p pair) BaseAsset() string {
	return p.baseAsset
}

func (p pair) QuoteAsset() string {
	return p.quoteAsset
}

func (p pair) BaseValue() uint64 {
	return p.baseValue
}

func (p pair) QuoteValue() uint64 {
	return p.quoteValue
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

func createOutputs(
	baseFragments, quoteFragments []uint64,
	feeAmount uint64,
	addresses []string,
	assetValuePair pair,
) []ports.TxOut {
	outsLen := len(baseFragments) + len(quoteFragments)
	outputs := make([]ports.TxOut, 0, outsLen)

	index := 0
	for _, v := range baseFragments {
		outputs = append(outputs, TxOut{
			asset:   assetValuePair.baseAsset,
			value:   int64(v),
			address: addresses[index],
		})
		index++
	}
	for _, v := range quoteFragments {
		outputs = append(outputs, TxOut{
			asset:   assetValuePair.quoteAsset,
			value:   int64(v),
			address: addresses[index],
		})
		index++
	}

	return outputs
}
