package transactionutil

import (
	"encoding/hex"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/transaction"
)

type UnblindedResult struct {
	AssetHash string
	Value     uint64
}

func UnblindOutput(
	utxo *transaction.TxOutput,
	blindKey []byte,
) (*UnblindedResult, bool) {
	// ignore the following errors because this function is called only if
	// asset and value commitments are defined. However, if a bad (nil) nonce
	// is passed to the UnblindOutput function, this will not be able to reveal
	// secrets of the output.
	nonce, _ := confidential.NonceHash(utxo.Nonce, blindKey)
	arg := confidential.UnblindOutputArg{
		Nonce:           nonce,
		AssetCommitment: utxo.Asset,
		ValueCommitment: utxo.Value,
		ScriptPubkey:    utxo.Script,
		Rangeproof:      utxo.RangeProof,
	}
	revealed, err := confidential.UnblindOutput(arg)
	if err == nil {
		return &UnblindedResult{
			AssetHash: hex.EncodeToString(bufferutil.ReverseBytes(revealed.Asset)),
			Value:     revealed.Value,
		}, true
	}

	return nil, false
}
