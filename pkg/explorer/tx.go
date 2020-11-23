package explorer

import (
	"encoding/hex"
	"encoding/json"

	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/transaction"
)

type Transaction interface {
	Hash() string
	Version() int
	Locktime() int
	Inputs() []*transaction.TxInput
	Outputs() []*transaction.TxOutput
	Size() int
	Weight() int
	Fee() int
	Confirmed() bool
}

type tx struct {
	TxHash     string                 `json:"txid"`
	TxVersion  int                    `json:"version"`
	TxLocktime int                    `json:"locktime"`
	TxInputs   []interface{}          `json:"vin"`
	TxOutputs  []interface{}          `json:"vout"`
	TxSize     int                    `json:"size"`
	TxWeight   int                    `json:"weight"`
	TxFee      int                    `json:"fee"`
	TxStatus   map[string]interface{} `json:"status"`
}

func NewTxFromJSON(txJSON string) (Transaction, error) {
	t := &tx{}
	if err := json.Unmarshal([]byte(txJSON), t); err != nil {
		return nil, err
	}
	return t, nil
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
	ins := make([]*transaction.TxInput, 0, len(t.TxInputs))
	for _, v := range t.TxInputs {
		in := parseInput(v)
		ins = append(ins, in)
	}
	return ins
}

func (t *tx) Outputs() []*transaction.TxOutput {
	outs := make([]*transaction.TxOutput, 0, len(t.TxOutputs))
	for _, v := range t.TxOutputs {
		out := parseOutput(v)
		outs = append(outs, out)
	}
	return outs
}

func (t *tx) Size() int {
	return t.TxSize
}

func (t *tx) Weight() int {
	return t.TxWeight
}

func (t *tx) Fee() int {
	return t.TxFee
}

func (t *tx) Confirmed() bool {
	return t.TxStatus["confirmed"].(bool)
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
