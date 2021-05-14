package elements

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/address"
)

// GetTransaction returns the transaction indentified by its hash
// by calling the getrawtrancation RPC
func (e *elements) GetTransaction(txid string) (explorer.Transaction, error) {
	r, err := e.client.call("getrawtransaction", []interface{}{txid, 1})
	if err = handleError(err, &r); err != nil {
		return nil, fmt.Errorf("rawtx: %w", err)
	}
	return NewTxFromJSON(string(r.Result))
}

// GetTransactionHex returns the hex of the transaction indentified by its hash
// by calling the getrawtrancation RPC
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

// IsTransactionConfirmed returns whether a tx is already confirmed by calling
// the gettransaction RPC
func (e *elements) IsTransactionConfirmed(txid string) (bool, error) {
	data, err := e.getTransaction(txid)
	if err != nil {
		return false, err
	}
	confirmations, ok := data["confirmations"].(float64)
	if !ok {
		return false, nil
	}
	return confirmations > 0, nil
}

// GetTransactionStatus returns info about the status of a transaction. In case
// it's not yet conifrmed a {"confirmed": false} response is returned.
// Otherwise some other info about the block that includes the tx are returned
// along with its confirmation status. This method makes use of gettransaction
// and getblock RPCs.
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

type txResult struct {
	tx  explorer.Transaction
	err error
}

// GetTransactionsForAddress returns all the transactions for the provided
// address. It makes use of importaddress to add the address to those
// tracked by the Elements node. A label, (ie. the resulting output script) is
// associated with it when importing to prevent doing this operation for those
// already tracked. The transactions are retrieved via the
// listreceivedbyaddress RPC.
func (e *elements) GetTransactionsForAddress(addr string, blindingKey []byte) ([]explorer.Transaction, error) {
	addrLabel, err := addressLabel(addr)
	if err != nil {
		return nil, fmt.Errorf("label: %w", err)
	}
	isImportedAddress, err := e.isAddressImported(addr)
	if err != nil {
		return nil, fmt.Errorf("check import: %w", err)
	}
	if !isImportedAddress {
		if err := e.importAddress(addr, addrLabel, blindingKey, true); err != nil {
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

		txs := make([]explorer.Transaction, 0, len(txids))
		chRes := make(chan txResult)
		wg := &sync.WaitGroup{}
		wg.Add(len(txids))

		go func() {
			wg.Wait()
			close(chRes)
		}()

		for i := range txids {
			txid := txids[i]
			go e.getTxDetails(txid.(string), chRes, wg)
		}

		for r := range chRes {
			if r.err != nil {
				return nil, r.err
			}
			txs = append(txs, r.tx)
		}

		return txs, nil
	}
	return nil, nil
}

// BroadcastTransaction publishes a new transaction (provided in hex format) to
// the network by calling the sendrawtransaction RPC.
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

/**** Regtest only ****/

// Faucet sends the requested sats to the given address using the sendtoaddress
// RPC. Also, 1 block is mined with generatetoaddress to get the faucet tx
// confirmed.
func (e *elements) Faucet(addr string, amount float64, asset string) (string, error) {
	net, err := address.NetworkForAddress(addr)
	if err != nil {
		return "", fmt.Errorf("args: %w", err)
	}

	if len(asset) <= 0 {
		asset = net.AssetID
	}

	btcAmount := float64(amount) / math.Pow10(8)
	r, err := e.client.call("sendtoaddress", []interface{}{
		addr, btcAmount, "", "", false, false, 1, "UNSET", asset,
	})
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

// Mint issues a new asset with the given issuance amount and send it to the
// provided address. It uses issueasset RPC for minting a new asset, and
// sendtoaddress for funding the provided address. Simalrly to Faucet, also
// this mines 1 block to confirm the mint tx.
func (e *elements) Mint(address string, amount float64) (string, string, error) {
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
	chRes chan txResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	tx, err := e.GetTransaction(txid)
	if err != nil {
		chRes <- txResult{err: err}
		return
	}
	chRes <- txResult{tx: tx}
}
