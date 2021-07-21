package esplora

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/vulpemventures/go-elements/transaction"
)

func (e *esplora) GetUnspents(addr string, blindingKeys [][]byte) (coins []explorer.Utxo, err error) {
	return e.getUtxos(addr, blindingKeys)
}

type utxosResult struct {
	utxos []explorer.Utxo
	err   error
}

func (e *esplora) GetUnspentsForAddresses(
	addresses []string,
	blindingKeys [][]byte,
) ([]explorer.Utxo, error) {
	chRes := make(chan utxosResult)
	utxos := make([]explorer.Utxo, 0)
	wg := &sync.WaitGroup{}
	wg.Add(len(addresses))

	go func() {
		wg.Wait()
		close(chRes)
	}()

	for i := range addresses {
		addr := addresses[i]
		go e.getUnspentsForAddress(addr, blindingKeys, chRes, wg)
		time.Sleep(1 * time.Millisecond)
	}

	for r := range chRes {
		if r.err != nil {
			return nil, r.err
		}

		utxos = append(utxos, r.utxos...)
	}

	return utxos, nil
}

func (e *esplora) GetUnspentStatus(
	hash string, index uint32,
) (*explorer.UtxoStatus, error) {
	url := fmt.Sprintf("%s/tx/%s/outspend/%d", e.apiURL, hash, index)
	status, resp, err := e.client.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf(resp)
	}

	var utxoStatus map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &utxoStatus); err != nil {
		return nil, fmt.Errorf("error on retrieving utxo status: %s", err)
	}

	spent := utxoStatus["spent"].(bool)
	txHash := ""
	if hash, ok := utxoStatus["txid"]; ok {
		txHash = hash.(string)
	}
	txInIndex := -1
	if index, ok := utxoStatus["vin"]; ok {
		txInIndex = int(index.(float64))
	}

	return &explorer.UtxoStatus{
		Spent:        spent,
		TxHash:       txHash,
		TxInputIndex: txInIndex,
	}, nil
}

type utxoResult struct {
	utxo explorer.Utxo
	err  error
}

func (e *esplora) getUtxos(addr string, blindingKeys [][]byte) ([]explorer.Utxo, error) {
	url := fmt.Sprintf("%s/address/%s/utxo", e.apiURL, addr)
	status, resp, err := e.client.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return nil, fmt.Errorf("error on retrieving utxos: %s", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf(resp)
	}

	var witnessOuts []witnessUtxo
	if err := json.Unmarshal([]byte(resp), &witnessOuts); err != nil {
		return nil, fmt.Errorf("error on retrieving utxos: %s", err)
	}

	utxos := make([]explorer.Utxo, 0, len(witnessOuts))
	chRes := make(chan utxoResult)
	wg := &sync.WaitGroup{}
	wg.Add(len(witnessOuts))

	go func() {
		wg.Wait()
		close(chRes)
	}()

	for i := range witnessOuts {
		out := witnessOuts[i]
		go e.getUtxoDetails(out, chRes, wg)
	}

	utxosToUnblind := make([]explorer.Utxo, 0)
	for r := range chRes {
		if r.err != nil {
			return nil, fmt.Errorf("error on retrieving utxos: %s", r.err)
		}

		if len(blindingKeys) > 0 && r.utxo.IsConfidential() {
			utxosToUnblind = append(utxosToUnblind, r.utxo)
		} else {
			utxos = append(utxos, r.utxo)
		}
	}

	if len(utxosToUnblind) > 0 {
		chRes := make(chan utxoResult)
		wg.Add(len(utxosToUnblind))

		go func() {
			wg.Wait()
			close(chRes)
		}()

		for i := range utxosToUnblind {
			utxo := utxosToUnblind[i]
			go unblindUtxo(utxo, blindingKeys, chRes, wg)
		}

		for r := range chRes {
			if r.err != nil {
				return nil, fmt.Errorf("error on unblinding utxos: %s", r.err)
			}

			utxos = append(utxos, r.utxo)
		}
	}

	return utxos, nil
}

func (e *esplora) getUnspentsForAddress(
	addr string,
	blindingKeys [][]byte,
	chRes chan utxosResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	utxos, err := e.getUtxos(addr, blindingKeys)
	if err != nil {
		chRes <- utxosResult{err: err}
		return
	}
	chRes <- utxosResult{utxos: utxos}
}

func (e *esplora) getUtxoDetails(
	utxo witnessUtxo,
	chRes chan utxoResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// in case of error the status is defaulted to unconfirmed
	confirmed, _ := e.IsTransactionConfirmed(utxo.Hash())

	prevoutTxHex, err := e.GetTransactionHex(utxo.Hash())
	if err != nil {
		chRes <- utxoResult{err: err}
		return
	}
	trx, _ := transaction.NewTxFromHex(prevoutTxHex)
	prevout := trx.Outputs[utxo.Index()]

	if utxo.IsConfidential() {
		utxo.UNonce = prevout.Nonce
		utxo.URangeProof = prevout.RangeProof
		utxo.USurjectionProof = prevout.SurjectionProof
	}
	utxo.UScript = prevout.Script
	utxo.UStatus = status{Confirmed: confirmed}

	chRes <- utxoResult{utxo: utxo}
}

func unblindUtxo(
	u explorer.Utxo,
	blindKeys [][]byte,
	chRes chan utxoResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	utxo := u.(witnessUtxo)
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

	for i := range blindKeys {
		blindKey := blindKeys[i]
		unblinded, ok := transactionutil.UnblindOutput(txOut, blindKey)
		if ok {
			utxo.UAsset = unblinded.AssetHash
			utxo.UValue = unblinded.Value
			utxo.UValueBlinder = unblinded.ValueBlinder
			utxo.UAssetBlinder = unblinded.AssetBlinder
			chRes <- utxoResult{utxo: utxo}
			return
		}
	}

	chRes <- utxoResult{err: fmt.Errorf("unable to unblind utxo with provided keys")}
}
