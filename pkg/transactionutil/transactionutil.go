package transactionutil

import (
	"encoding/hex"

	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/transaction"
)

type UnblindedResult struct {
	AssetHash    string
	Value        uint64
	AssetBlinder []byte
	ValueBlinder []byte
}

func UnblindOutput(
	utxo *transaction.TxOutput,
	blindKey []byte,
) (*UnblindedResult, bool) {
	revealed, err := confidential.UnblindOutputWithKey(utxo, blindKey)
	if err != nil {
		return nil, false
	}
	return &UnblindedResult{
		AssetHash:    hex.EncodeToString(elementsutil.ReverseBytes(revealed.Asset)),
		Value:        revealed.Value,
		AssetBlinder: revealed.AssetBlindingFactor,
		ValueBlinder: revealed.ValueBlindingFactor,
	}, true
}

func NewFeeOutput(feeAmount uint64) []*transaction.TxOutput {
	feeAsset, _ := bufferutil.AssetHashToBytes(config.GetNetwork().AssetID)
	feeValue, _ := bufferutil.ValueToBytes(feeAmount)
	feeScript := make([]byte, 0)
	return []*transaction.TxOutput{
		transaction.NewTxOutput(feeAsset, feeValue, feeScript),
	}
}
