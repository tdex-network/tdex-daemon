package esplora

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/transaction"
)

/**** TRANSACTION ****/

// tx is the implementation of the explorer's Transaction interface
type tx struct {
	TxHash      string
	TxVersion   int
	TxLocktime  int
	TxInputs    []*transaction.TxInput
	TxOutputs   []*transaction.TxOutput
	TxSize      int
	TxWeight    int
	TxConfirmed bool
}

// NewTxFromJSON is the factory for a Transaction in given its JSON format.
func NewTxFromJSON(txJSON string) (explorer.Transaction, error) {
	t := &tx{}
	if err := json.Unmarshal([]byte(txJSON), t); err != nil {
		return nil, fmt.Errorf("invalid tx JSON")
	}
	return t, nil
}

func NewTxFromHex(txhex string, confirmed bool) (explorer.Transaction, error) {
	t, err := transaction.NewTxFromHex(txhex)
	if err != nil {
		return nil, err
	}

	return &tx{
		TxHash:      t.TxHash().String(),
		TxVersion:   int(t.Version),
		TxLocktime:  int(t.Locktime),
		TxSize:      t.VirtualSize(),
		TxWeight:    t.Weight(),
		TxInputs:    t.Inputs,
		TxOutputs:   t.Outputs,
		TxConfirmed: confirmed,
	}, nil
}

func (t *tx) Hash() string {
	return t.TxHash
}

func (t *tx) Version() int {
	return t.TxVersion
}

func (t *tx) Locktime() int {
	return t.TxLocktime
}

func (t *tx) Inputs() []*transaction.TxInput {
	return t.TxInputs
}

func (t *tx) Outputs() []*transaction.TxOutput {
	return t.TxOutputs
}

func (t *tx) Size() int {
	return t.TxSize
}

func (t *tx) Weight() int {
	return t.TxWeight
}

func (t *tx) Confirmed() bool {
	return t.TxConfirmed
}

func parseInput(i interface{}) *transaction.TxInput {
	m := i.(map[string]interface{})

	txid := interfaceToBytes(m["txid"])
	vout := uint32(m["vout"].(float64))
	in := transaction.NewTxInput(elementsutil.ReverseBytes(txid), vout)

	in.Script = interfaceToBytes(m["scriptsig"])

	if m["witness"] != nil {
		in.Witness = parseWitness(m["witness"])
	}

	in.Sequence = uint32(m["sequence"].(float64))
	in.IsPegin = m["is_pegin"].(bool)

	// TODO: parse issuance
	return in
}

func parseOutput(i interface{}) *transaction.TxOutput {
	m := i.(map[string]interface{})

	asset := make([]byte, 33)
	if m["asset"] != nil {
		asset, _ = bufferutil.AssetHashToBytes(m["asset"].(string))
	} else {
		asset = interfaceToBytes(m["assetcommitment"])
	}

	value := make([]byte, 0)
	if m["value"] != nil {
		value, _ = bufferutil.ValueToBytes(uint64(m["value"].(float64)))
	} else {
		value = interfaceToBytes(m["valuecommitment"])
	}
	script := interfaceToBytes(m["scriptpubkey"])

	// TODO: parse nonce and proof; these values must be retrieved by parsing the
	// hex of the transaction like done in getUtxoDetails
	return transaction.NewTxOutput(asset, value, script)
}

func interfaceToBytes(i interface{}) []byte {
	s := i.(string)
	buf, _ := hex.DecodeString(s)
	return buf
}

func parseWitness(i interface{}) transaction.TxWitness {
	witnesses := i.([]interface{})
	buf := make(transaction.TxWitness, 0, len(witnesses))
	for _, wit := range witnesses {
		b := interfaceToBytes(wit)
		buf = append(buf, b)
	}
	return buf
}

/**** UTXO *****/

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
	UValueBlinder    []byte
	UAssetBlinder    []byte
}

type status struct {
	Confirmed bool `json:"confirmed"`
}

// NewUnconfidentialWitnessUtxo is the factory for a non-confidential witnessUtxo.
func NewUnconfidentialWitnessUtxo(
	hash string,
	index uint32,
	value uint64,
	asset string,
	script []byte,
) explorer.Utxo {
	return witnessUtxo{
		UHash:   hash,
		UIndex:  index,
		UValue:  value,
		UAsset:  asset,
		UScript: script,
	}
}

// NewConfidentialWitnessUtxo is the factory for a confidential witnessUtxo.
func NewConfidentialWitnessUtxo(
	hash string,
	index uint32,
	valueCommitment, assetCommitment string,
	script, nonce, rangeProof, surjectionProof []byte,
) explorer.Utxo {
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

// NewWitnessUtxo is the factory for an unblinded confidential witnessUtxo.
func NewWitnessUtxo(
	hash string, index uint32,
	value uint64, asset string,
	valueCommitment, assetCommitment string,
	valueBlinder, assetBlinder []byte,
	script, nonce, rangeProof, surjectionProof []byte,
	confirmed bool,
) explorer.Utxo {
	return witnessUtxo{
		UHash:            hash,
		UIndex:           index,
		UValue:           value,
		UAsset:           asset,
		UValueCommitment: valueCommitment,
		UAssetCommitment: assetCommitment,
		UValueBlinder:    valueBlinder,
		UAssetBlinder:    assetBlinder,
		UScript:          script,
		UNonce:           nonce,
		URangeProof:      rangeProof,
		USurjectionProof: surjectionProof,
		UStatus:          status{Confirmed: confirmed},
	}
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

func (wu witnessUtxo) ValueBlinder() []byte {
	return wu.UValueBlinder
}

func (wu witnessUtxo) AssetBlinder() []byte {
	return wu.UAssetBlinder
}

func (wu witnessUtxo) IsConfidential() bool {
	return len(wu.UValueCommitment) > 0 && len(wu.UAssetCommitment) > 0
}

func (wu witnessUtxo) IsConfirmed() bool {
	return wu.UStatus.Confirmed
}

func (wu witnessUtxo) IsRevealed() bool {
	return len(wu.ValueBlinder()) > 0 && len(wu.AssetBlinder()) > 0
}

func (wu witnessUtxo) Parse() (*transaction.TxInput, *transaction.TxOutput, error) {
	inHash, err := bufferutil.TxIDToBytes(wu.UHash)
	if err != nil {
		return nil, nil, err
	}
	input := transaction.NewTxInput(inHash, wu.UIndex)

	var witnessUtxo *transaction.TxOutput
	if wu.IsConfidential() {
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
