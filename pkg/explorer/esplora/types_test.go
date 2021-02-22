package esplora_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
)

func TestNewTxFromJSON(t *testing.T) {
	file, err := ioutil.ReadFile("testdata/fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures map[string]interface{}
	if err = json.Unmarshal(file, &fixtures); err != nil {
		t.Fatal(err)
	}
	tests := fixtures["transactions"].([]interface{})

	for _, test := range tests {
		tt := test.(map[string]interface{})
		trx, err := esplora.NewTxFromJSON(tt["tx"].(string))
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, tt["hash"].(string), trx.Hash())
		assert.Equal(t, toInt(tt["version"]), trx.Version())
		assert.Equal(t, toInt(tt["locktime"]), trx.Locktime())
		assert.Equal(t, toInt(tt["numInputs"]), len(trx.Inputs()))
		assert.Equal(t, toInt(tt["numOutputs"]), len(trx.Outputs()))
		assert.Equal(t, toInt(tt["size"]), trx.Size())
		assert.Equal(t, toInt(tt["weight"]), trx.Weight())
		assert.Equal(t, toInt(tt["fee"]), trx.Fee())
		assert.Equal(t, tt["confirmed"].(bool), trx.Confirmed())
	}
}

func toInt(v interface{}) int {
	return int(v.(float64))
}
