package mapper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (m *mapperService) GetUnspentStatus(
	txid string, index uint32,
) (*UtxoStatus, error) {
	url := fmt.Sprintf("%s/tx/%s/outspend/%d", m.esploraUrl, txid, index)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	unspentStatus := &UtxoStatus{}
	if err := json.Unmarshal(body, unspentStatus); err != nil {
		return nil, fmt.Errorf("error on parsing unspent status: %s", err)
	}
	return unspentStatus, nil
}

type UtxoStatus struct {
	Spent  bool   `json:"spent"`
	Txid   string `json:"txid"`
	Vin    int    `json:"vin"`
	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int    `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int    `json:"block_time"`
	} `json:"status"`
}
