package swap

import (
	"encoding/hex"

	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/transaction"
)

// reverseBytes returns a copy of the given byte slice with elems in reverse order.
func reverseBytes(buf []byte) []byte {
	if len(buf) < 1 {
		return buf
	}
	tmp := make([]byte, len(buf))
	copy(tmp, buf)
	for i := len(tmp)/2 - 1; i >= 0; i-- {
		j := len(tmp) - 1 - i
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}
	return tmp
}

func assetHashFromBytes(buffer []byte) string {
	// We remove the first byte from the buffer array that represents if confidential or unconfidential
	return hex.EncodeToString(reverseBytes(buffer[1:]))
}

func valueFromBytes(buffer []byte) uint64 {
	var elementsValue [9]byte
	copy(elementsValue[:], buffer[0:9])
	value, _ := confidential.ElementsToSatoshiValue(elementsValue)
	return value
}

type unblindedResult struct {
	assetHash string
	value     uint64
}

func unblindPrevOut(utxo *transaction.TxOutput, blindKey []byte) (*unblindedResult, bool) {
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
		return &unblindedResult{
			assetHash: hex.EncodeToString(reverseBytes(revealed.Asset)),
			value:     revealed.Value,
		}, true
	}

	return nil, false
}

func unblindOut(out *transaction.TxOutput, blindKey []byte) (*unblindedResult, bool) {
	nonce, _ := confidential.NonceHash(out.Nonce, blindKey)
	arg := confidential.UnblindOutputArg{
		Nonce:           nonce,
		AssetCommitment: out.Asset,
		ValueCommitment: out.Value,
		ScriptPubkey:    out.Script,
		Rangeproof:      out.RangeProof,
	}
	revealed, err := confidential.UnblindOutput(arg)
	if err == nil {
		return &unblindedResult{
			assetHash: hex.EncodeToString(reverseBytes(revealed.Asset)),
			value:     revealed.Value,
		}, true
	}

	return nil, false
}
