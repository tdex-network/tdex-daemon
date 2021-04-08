package elements

import (
	"encoding/hex"
	"encoding/json"
	"math"

	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/transaction"
)

type tx struct {
	TxHash      string      `json:"txid"`
	TxVersion   int         `json:"version"`
	TxLocktime  int         `json:"locktime"`
	TxSize      int         `json:"size"`
	TxWeight    int         `json:"weight"`
	TxInputs    interface{} `json:"vin"`
	TxOutputs   interface{} `json:"vout"`
	TxConfirmed bool
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

func NewTxFromJSON(txJSON string) (explorer.Transaction, error) {
	var t tx
	if err := json.Unmarshal([]byte(txJSON), &t); err != nil {
		return nil, ErrInvalidTxJSON
	}
	return &t, nil
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
	return t.TxInputs.([]*transaction.TxInput)
}

func (t *tx) Outputs() []*transaction.TxOutput {
	return t.TxOutputs.([]*transaction.TxOutput)
}

func (t *tx) Size() int {
	return t.TxSize
}

func (t *tx) Weight() int {
	return t.TxWeight
}

func (t *tx) Fee() int {
	var fee uint64
	for _, out := range t.Outputs() {
		if len(out.Script) <= 0 {
			fee, _ = elementsutil.ElementsToSatoshiValue(out.Value)
		}
	}
	return int(fee)
}

func (t *tx) Confirmed() bool {
	return t.TxConfirmed
}

type elementsUnspent struct {
	UAddress          string  `json:"address,omitempty"`
	ULabel            string  `json:"label,omitempty"`
	UScriptPubKey     string  `json:"scriptPubKey,omitempty"`
	UConfirmations    int64   `json:"confirmations"`
	UTxID             string  `json:"txid"`
	UVout             uint32  `json:"vout"`
	UAmount           float64 `json:"amount"`
	UAsset            string  `json:"asset,omitempty"`
	UAmountCommitment string  `json:"amountcommitment,omitempty"`
	UAssetCommitment  string  `json:"assetcommitment,omitempty"`
	UAmountBlinder    string  `json:"amountblinder,omitempty"`
	UAssetBlinder     string  `json:"assetblinder,omitempty"`
	UNonce            []byte
	URangeProof       []byte
	USurjectionProof  []byte
}

func (eu elementsUnspent) Hash() string {
	return eu.UTxID
}

func (eu elementsUnspent) Index() uint32 {
	return eu.UVout
}

func (eu elementsUnspent) Value() uint64 {
	return uint64(eu.UAmount * math.Pow10(8))
}

func (eu elementsUnspent) Asset() string {
	return eu.UAsset
}

func (eu elementsUnspent) ValueCommitment() string {
	return eu.UAmountCommitment
}

func (eu elementsUnspent) AssetCommitment() string {
	return eu.UAssetCommitment
}

func (eu elementsUnspent) ValueBlinder() []byte {
	amountBlinder, _ := hex.DecodeString(eu.UAmountBlinder)
	return amountBlinder
}
func (eu elementsUnspent) AssetBlinder() []byte {
	assetBlinder, _ := hex.DecodeString(eu.UAssetBlinder)
	return assetBlinder
}

func (eu elementsUnspent) Script() []byte {
	script, _ := hex.DecodeString((eu.UScriptPubKey))
	return script
}

// Elements node does not return utxos' nonce, and proofs but it's ok as long
// as they are always revealed.
func (eu elementsUnspent) Nonce() []byte {
	return eu.UNonce
}

func (eu elementsUnspent) RangeProof() []byte {
	return eu.URangeProof
}

func (eu elementsUnspent) SurjectionProof() []byte {
	return eu.USurjectionProof
}

func (eu elementsUnspent) IsConfidential() bool {
	return len(eu.UAmountCommitment) > 0 && len(eu.UAssetCommitment) > 0
}

func (eu elementsUnspent) IsConfirmed() bool {
	return eu.UConfirmations > 0
}

func (eu elementsUnspent) IsRevealed() bool {
	return len(eu.UAmountBlinder) > 0 && len(eu.UAssetBlinder) > 0
}

func (eu elementsUnspent) Parse() (*transaction.TxInput, *transaction.TxOutput, error) {
	inHash, err := bufferutil.TxIDToBytes(eu.UTxID)
	if err != nil {
		return nil, nil, err
	}
	input := transaction.NewTxInput(inHash, eu.UVout)

	var witnessUtxo *transaction.TxOutput
	if eu.IsConfidential() {
		assetCommitment, err := bufferutil.CommitmentToBytes(eu.UAssetCommitment)
		if err != nil {
			return nil, nil, err
		}
		valueCommitment, err := bufferutil.CommitmentToBytes(eu.UAmountCommitment)
		if err != nil {
			return nil, nil, err
		}
		witnessUtxo = &transaction.TxOutput{
			Nonce:           eu.Nonce(),
			Script:          eu.Script(),
			Asset:           assetCommitment,
			Value:           valueCommitment,
			RangeProof:      eu.RangeProof(),
			SurjectionProof: eu.SurjectionProof(),
		}
	} else {
		asset, err := bufferutil.AssetHashToBytes(eu.UAsset)
		if err != nil {
			return nil, nil, err
		}

		value, err := bufferutil.ValueToBytes(eu.Value())
		if err != nil {
			return nil, nil, err
		}

		witnessUtxo = transaction.NewTxOutput(asset, value, eu.Script())
	}

	return input, witnessUtxo, nil
}
