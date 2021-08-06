package esplora

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

func (e *esplora) GetTransaction(hash string) (explorer.Transaction, error) {
	txHex, err := e.getTransactionHex(hash)
	if err != nil {
		return nil, err
	}
	confirmed, err := e.isTransactionConfirmed(hash)
	if err != nil {
		return nil, err
	}

	return NewTxFromHex(txHex, confirmed)
}

func (e *esplora) GetTransactionHex(hash string) (string, error) {
	return e.getTransactionHex(hash)
}

func (e *esplora) IsTransactionConfirmed(hash string) (bool, error) {
	return e.isTransactionConfirmed(hash)
}

func (e *esplora) GetTransactionStatus(
	hash string,
) (explorer.TransactionStatus, error) {
	return e.getTransactionStatus(hash)
}

func (e *esplora) GetTransactionsForAddress(address string, _ []byte) ([]explorer.Transaction, error) {
	url := fmt.Sprintf("%s/address/%s/txs", e.apiURL, address)
	status, resp, err := e.client.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf(resp)
	}

	return parseTransactions(resp)
}

func (e *esplora) BroadcastTransaction(txHex string) (string, error) {
	url := fmt.Sprintf("%s/tx", e.apiURL)
	headers := map[string]string{
		"Content-Type": "text/plain",
	}

	status, resp, err := e.client.NewHTTPRequest(
		"POST",
		url,
		txHex,
		headers,
	)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf(resp)
	}

	return resp, nil
}

func (e *esplora) Faucet(address string, amount float64, asset string) (string, error) {
	url := fmt.Sprintf("%s/faucet", e.apiURL)
	payload := map[string]interface{}{"address": address, "amount": amount, "asset": asset}
	body, _ := json.Marshal(payload)
	bodyString := string(body)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	status, resp, err := e.client.NewHTTPRequest(
		"POST", url, bodyString, headers,
	)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf(resp)
	}

	var rr map[string]string
	json.Unmarshal([]byte(resp), &rr)

	return rr["txId"], nil
}

func (e *esplora) Mint(address string, amount float64) (string, string, error) {
	url := fmt.Sprintf("%s/mint", e.apiURL)
	payload := map[string]interface{}{"address": address, "quantity": amount}
	body, _ := json.Marshal(payload)
	bodyString := string(body)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	status, resp, err := e.client.NewHTTPRequest(
		"POST", url, bodyString, headers,
	)
	if err != nil {
		return "", "", err
	}
	if status != http.StatusOK {
		return "", "", fmt.Errorf(resp)
	}

	var rr map[string]string
	json.Unmarshal([]byte(resp), &rr)

	return rr["txId"], rr["asset"], nil
}

func (e *esplora) getTransactionHex(hash string) (string, error) {
	url := fmt.Sprintf(
		"%s/tx/%s/hex",
		e.apiURL,
		hash,
	)
	status, resp, err := e.client.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf(resp)
	}

	return resp, nil
}

func (e *esplora) isTransactionConfirmed(hash string) (bool, error) {
	status, err := e.getTransactionStatus(hash)
	if err != nil {
		return false, err
	}
	return status.Confirmed(), nil
}

func (e *esplora) getTransactionStatus(
	hash string,
) (explorer.TransactionStatus, error) {
	url := fmt.Sprintf(
		"%s/tx/%s/status",
		e.apiURL,
		hash,
	)
	status, resp, err := e.client.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf(resp)
	}

	var txStatus txStatus
	if err := json.Unmarshal([]byte(resp), &txStatus); err != nil {
		return nil, err
	}

	return txStatus, nil
}

type txResult struct {
	tx  explorer.Transaction
	err error
}

func parseTransactions(txList string) ([]explorer.Transaction, error) {
	txInterfaces := make([]interface{}, 0)
	if err := json.Unmarshal([]byte(txList), &txInterfaces); err != nil {
		return nil, err
	}

	txs := make([]explorer.Transaction, 0, len(txInterfaces))
	chRes := make(chan txResult)
	wg := &sync.WaitGroup{}
	wg.Add(len(txInterfaces))

	go func() {
		wg.Wait()
		close(chRes)
	}()

	for i := range txInterfaces {
		txi := txInterfaces[i]

		go parseTransaction(txi, chRes, wg)
	}

	for r := range chRes {
		if r.err != nil {
			return nil, r.err
		}
		txs = append(txs, r.tx)
	}

	return txs, nil
}

func parseTransaction(txi interface{}, chRes chan txResult, wg *sync.WaitGroup) {
	defer wg.Done()

	trx, err := json.Marshal(txi)
	if err != nil {
		chRes <- txResult{err: err}
		return
	}
	tx, err := NewTxFromJSON(string(trx))
	if err != nil {
		chRes <- txResult{err: err}
		return
	}

	chRes <- txResult{tx: tx}
}
