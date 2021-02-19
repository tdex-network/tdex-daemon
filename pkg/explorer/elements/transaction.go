package elements

import (
	"encoding/json"
	"fmt"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

func (e *elements) GetTransactionHex(txid string) (string, error) {
	r, err := e.client.call("getrawtransaction", []interface{}{txid})
	if err = handleError(err, &r); err != nil {
		return "", fmt.Errorf("rawtx: %w", err)
	}

	var txhex string
	if err := json.Unmarshal(r.Result, &txhex); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}
	return txhex, nil
}

func (e *elements) IsTransactionConfirmed(txid string) (bool, error) {
	data, err := e.GetTransactionStatus(txid)
	if err != nil {
		return false, err
	}
	return data["confirmed"].(bool), nil
}

func (e *elements) GetTransactionStatus(txid string) (map[string]interface{}, error) {
	data, err := e.getTransaction(txid)
	if err != nil {
		return nil, err
	}

	isConfirmed := data["confirmations"].(float64) > 0
	if isConfirmed {
		r, err := e.client.call("getblock", []interface{}{data["blockhash"]})
		if err = handleError(err, &r); err != nil {
			return nil, fmt.Errorf("block: %w", err)
		}

		data := map[string]interface{}{}
		if err := json.Unmarshal(r.Result, &data); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}

		return map[string]interface{}{
			"confirmed":    true,
			"block_hash":   data["hash"],
			"block_height": data["height"],
			"block_time":   data["time"],
		}, nil
	}
	return map[string]interface{}{
		"confirmed": false,
	}, nil
}

func (e *elements) GetTransactionsForAddress(addr string) ([]explorer.Transaction, error) {
	addrLabel, err := addressLabel(addr)
	if err != nil {
		return nil, fmt.Errorf("label: %w", err)
	}
	isImportedAddress, err := e.isAddressImported(addr)
	if err != nil {
		return nil, fmt.Errorf("check import: %w", err)
	}
	if !isImportedAddress {
		if err := e.importAddress(addr, addrLabel); err != nil {
			return nil, fmt.Errorf("import: %w", err)
		}
	}

	r, err := e.client.call("listreceivedbyaddress", []interface{}{0, true, true, addr})
	if err = handleError(err, &r); err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	var data []interface{}
	if err := json.Unmarshal(r.Result, &data); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	if len(data) > 0 {
		info := data[0].(map[string]interface{})
		txids := info["txids"].([]interface{})
		confirmations := int(info["confirmations"].(float64))

		txs := make([]explorer.Transaction, 0, len(txids))
		chTxs := make(chan explorer.Transaction)
		chErr := make(chan error, 1)

		for _, txid := range txids {
			go e.getTxDetails(txid.(string), confirmations, chTxs, chErr)
			select {
			case tx := <-chTxs:
				if tx != nil {
					txs = append(txs, tx)
				}
			case err := <-chErr:
				close(chTxs)
				close(chErr)
				return nil, fmt.Errorf("tx details: %w", err)
			}
		}
		return txs, nil
	}
	return nil, nil
}

func (e *elements) BroadcastTransaction(txhex string) (string, error) {
	r, err := e.client.call("sendrawtransaction", []interface{}{txhex})
	if err = handleError(err, &r); err != nil {
		return "", fmt.Errorf("send: %w", err)
	}

	var txid string
	if err := json.Unmarshal(r.Result, &txid); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}
	return txid, nil
}

// Regtest only
func (e *elements) Faucet(address string) (string, error) {
	r, err := e.client.call("sendtoaddress", []interface{}{address, 1})
	if err = handleError(err, &r); err != nil {
		return "", fmt.Errorf("send: %w", err)
	}
	var txid string
	if err := json.Unmarshal(r.Result, &txid); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}

	if _, err := e.mine(); err != nil {
		return "", fmt.Errorf("mine: %w", err)
	}

	return txid, nil
}

func (e *elements) Mint(address string, amount int) (string, string, error) {
	r, err := e.client.call("issueasset", []interface{}{amount, 0})
	if err = handleError(err, &r); err != nil {
		return "", "", fmt.Errorf("asset: %w", err)
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(r.Result, &data); err != nil {
		return "", "", fmt.Errorf("asset unmarshal: %w", err)
	}
	asset := data["asset"].(string)

	r, err = e.client.call("sendtoaddress", []interface{}{address, amount, "", "", false, false, 1, "UNSET", asset})
	if err = handleError(err, &r); err != nil {
		return "", "", fmt.Errorf("send: %w", err)
	}
	var txid string
	if err := json.Unmarshal(r.Result, &txid); err != nil {
		return "", "", fmt.Errorf("send unmarhal: %w", err)
	}

	if _, err := e.mine(); err != nil {
		return "", "", fmt.Errorf("mine: %w", err)
	}

	return "", "", nil
}

func (e *elements) getTransaction(txid string) (map[string]interface{}, error) {
	r, err := e.client.call("gettransaction", []interface{}{txid, true})
	if err = handleError(err, &r); err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(r.Result, &data); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return data, nil
}

func (e *elements) getTxDetails(
	txid string,
	confirmations int,
	chTxs chan explorer.Transaction,
	chErr chan error,
) {
	txhex, err := e.GetTransactionHex(txid)
	if err != nil {
		chErr <- err
		return
	}

	tx, err := NewTxFromHex(txhex, confirmations)
	if err != nil {
		chErr <- err
		return
	}
	chTxs <- tx
}
