package explorer

import (
	"errors"

	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/vulpemventures/go-elements/transaction"
)

type Utxo interface {
	Hash() string
	Index() uint32
	Value() uint64
	Asset() string
	ValueCommitment() string
	AssetCommitment() string
	Script() []byte
	Nonce() []byte
	RangeProof() []byte
	SurjectionProof() []byte
	IsConfidential() bool
	IsConfirmed() bool
	SetScript(script []byte)
	SetUnconfidential(asset string, value uint64)
	SetConfidential(nonce, rangeProof, surjectionProof []byte)
	Parse() (*transaction.TxInput, *transaction.TxOutput, error)
}

func NewUnconfidentialWitnessUtxo(
	hash string,
	index uint32,
	value uint64,
	asset string,
	script []byte,
) Utxo {
	return witnessUtxo{
		UHash:   hash,
		UIndex:  index,
		UValue:  value,
		UAsset:  asset,
		UScript: script,
	}
}

func NewConfidentialWitnessUtxo(
	hash string,
	index uint32,
	valueCommitment, assetCommitment string,
	script, nonce, rangeProof, surjectionProof []byte,
) Utxo {
	return witnessUtxo{
		UHash:            hash,
		UIndex:           index,
		UValueCommitment: valueCommitment,
		UAssetCommitment: assetCommitment,
		UScript:          script,
		UNonce:           nonce,
		URangeProof:      rangeProof,
		USurjectionProof: surjectionProof,
	}
}

func NewWitnessUtxo(
	hash string, index uint32,
	value uint64, asset string,
	valueCommitment, assetCommitment string,
	script, nonce, rangeProof, surjectionProof []byte,
	confirmed bool,
) Utxo {
	return witnessUtxo{
		UHash:            hash,
		UIndex:           index,
		UValue:           value,
		UAsset:           asset,
		UValueCommitment: valueCommitment,
		UAssetCommitment: assetCommitment,
		UScript:          script,
		UNonce:           nonce,
		URangeProof:      rangeProof,
		USurjectionProof: surjectionProof,
		UStatus:          status{Confirmed: confirmed},
	}
}

type status struct {
	Confirmed bool `json:"confirmed"`
}

type witnessUtxo struct {
	UHash            string `json:"txid"`
	UIndex           uint32 `json:"vout"`
	UValue           uint64 `json:"value"`
	UAsset           string `json:"asset"`
	UValueCommitment string `json:"valuecommitment"`
	UAssetCommitment string `json:"assetcommitment"`
	UStatus          status `json:"status"`
	UScript          []byte
	UNonce           []byte
	URangeProof      []byte
	USurjectionProof []byte
}

func (wu witnessUtxo) Hash() string {
	return wu.UHash
}

func (wu witnessUtxo) Index() uint32 {
	return wu.UIndex
}

func (wu witnessUtxo) Value() uint64 {
	return wu.UValue
}

func (wu witnessUtxo) Asset() string {
	return wu.UAsset
}

func (wu witnessUtxo) ValueCommitment() string {
	return wu.UValueCommitment
}

func (wu witnessUtxo) AssetCommitment() string {
	return wu.UAssetCommitment
}

func (wu witnessUtxo) Nonce() []byte {
	return wu.UNonce
}

func (wu witnessUtxo) Script() []byte {
	return wu.UScript
}

func (wu witnessUtxo) RangeProof() []byte {
	return wu.URangeProof
}

func (wu witnessUtxo) SurjectionProof() []byte {
	return wu.USurjectionProof
}

func (wu witnessUtxo) IsConfidential() bool {
	return len(wu.UValueCommitment) > 0 && len(wu.UAssetCommitment) > 0
}

func (wu witnessUtxo) IsConfirmed() bool {
	return wu.UStatus.Confirmed
}

func (wu witnessUtxo) SetScript(script []byte) {
	wu.UScript = script
}

func (wu witnessUtxo) SetUnconfidential(asset string, value uint64) {
	wu.UAsset = asset
	wu.UValue = value
}

func (wu witnessUtxo) SetConfidential(nonce, rangeProof, surjectionProof []byte) {
	wu.UNonce = make([]byte, 0)
	wu.UNonce = nonce
	wu.URangeProof = make([]byte, 0)
	wu.URangeProof = rangeProof
	wu.USurjectionProof = make([]byte, 0)
	wu.USurjectionProof = surjectionProof
}

func (wu witnessUtxo) Parse() (*transaction.TxInput, *transaction.TxOutput, error) {
	inHash, err := bufferutil.TxIDToBytes(wu.UHash)
	if err != nil {
		return nil, nil, err
	}
	input := transaction.NewTxInput(inHash, wu.UIndex)

	var witnessUtxo *transaction.TxOutput
	if len(wu.URangeProof) != 0 && len(wu.USurjectionProof) != 0 {
		assetCommitment, err := bufferutil.CommitmentToBytes(wu.UAssetCommitment)
		if err != nil {
			return nil, nil, err
		}
		valueCommitment, err := bufferutil.CommitmentToBytes(wu.UValueCommitment)
		if err != nil {
			return nil, nil, err
		}
		witnessUtxo = &transaction.TxOutput{
			Nonce:           wu.UNonce,
			Script:          wu.UScript,
			Asset:           assetCommitment,
			Value:           valueCommitment,
			RangeProof:      wu.URangeProof,
			SurjectionProof: wu.USurjectionProof,
		}
	} else {
		asset, err := bufferutil.AssetHashToBytes(wu.UAsset)
		if err != nil {
			return nil, nil, err
		}

		value, err := bufferutil.ValueToBytes(wu.UValue)
		if err != nil {
			return nil, nil, err
		}

		witnessUtxo = transaction.NewTxOutput(asset, value, wu.UScript)
	}

	return input, witnessUtxo, nil
}

func unblindUtxo(
	utxo Utxo,
	blindKeys [][]byte,
	chUnspents chan Utxo,
	chErr chan error,
) {
	unspent := utxo.(witnessUtxo)
	for i := range blindKeys {
		blindKey := blindKeys[i]
		// ignore the following errors because this function is called only if
		// asset and value commitments are defined. However, if a bad (nil) nonce
		// is passed to the UnblindOutput function, this will not be able to reveal
		// secrets of the output.
		assetCommitment, _ := bufferutil.CommitmentToBytes(utxo.AssetCommitment())
		valueCommitment, _ := bufferutil.CommitmentToBytes(utxo.ValueCommitment())

		txOut := &transaction.TxOutput{
			Nonce:           utxo.Nonce(),
			Asset:           assetCommitment,
			Value:           valueCommitment,
			Script:          utxo.Script(),
			RangeProof:      utxo.RangeProof(),
			SurjectionProof: utxo.SurjectionProof(),
		}
		unBlinded, ok := transactionutil.UnblindOutput(txOut, blindKey)
		if ok {
			asset := unBlinded.AssetHash
			unspent.UAsset = asset
			unspent.UValue = unBlinded.Value
			chUnspents <- unspent
			return
		}
	}

	chErr <- errors.New("unable to unblind utxo with provided keys")
}

func (e *explorer) getUtxoDetails(
	out Utxo,
	chUnspents chan Utxo,
	chErr chan error,
) {
	unspent := out.(witnessUtxo)

	// in case of error the status is defaulted to unconfirmed
	confirmed, _ := e.IsTransactionConfirmed(unspent.Hash())

	prevoutTxHex, err := e.GetTransactionHex(unspent.Hash())
	if err != nil {
		chErr <- err
		return
	}
	trx, _ := transaction.NewTxFromHex(prevoutTxHex)
	prevout := trx.Outputs[unspent.Index()]

	if unspent.IsConfidential() {
		unspent.UNonce = prevout.Nonce
		unspent.URangeProof = prevout.RangeProof
		unspent.USurjectionProof = prevout.SurjectionProof
	}
	unspent.UScript = prevout.Script
	unspent.UStatus = status{Confirmed: confirmed}

	chUnspents <- unspent
}
