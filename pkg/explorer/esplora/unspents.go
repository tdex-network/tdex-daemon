package esplora

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/vulpemventures/go-elements/transaction"
)

func (e *esplora) GetUnspents(addr string, blindingKeys [][]byte) (coins []explorer.Utxo, err error) {
	return e.getUnspents(addr, blindingKeys)
}

func (e *esplora) GetUnspentsForAddresses(
	addresses []string,
	blindingKeys [][]byte,
) ([]explorer.Utxo, error) {
	chUnspents := make(chan []explorer.Utxo)
	chErr := make(chan error, 1)
	unspents := make([]explorer.Utxo, 0)

	for _, addr := range addresses {
		go e.getUnspentsForAddress(addr, blindingKeys, chUnspents, chErr)

		select {
		case err := <-chErr:
			close(chErr)
			close(chUnspents)
			return nil, err
		case unspentsForAddress := <-chUnspents:
			unspents = append(unspents, unspentsForAddress...)
		}
	}

	return unspents, nil
}

func (e *esplora) getUnspents(addr string, blindingKeys [][]byte) ([]explorer.Utxo, error) {
	url := fmt.Sprintf(
		"%s/address/%s/utxo",
		e.apiURL,
		addr,
	)
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

	unspents := make([]explorer.Utxo, len(witnessOuts))
	chUnspents := make(chan explorer.Utxo)
	chErr := make(chan error, 1)

	for i := range witnessOuts {
		out := witnessOuts[i]
		go e.getUtxoDetails(out, chUnspents, chErr)

		select {
		case err := <-chErr:
			if err != nil {
				close(chErr)
				close(chUnspents)
				return nil, fmt.Errorf("error on retrieving utxos: %s", err)
			}

		case unspent := <-chUnspents:
			if out.IsConfidential() && len(blindingKeys) > 0 {
				go unblindUtxo(unspent, blindingKeys, chUnspents, chErr)
				select {

				case err := <-chErr:
					close(chErr)
					close(chUnspents)
					return nil, fmt.Errorf("error on unblinding utxos: %s", err)

				case u := <-chUnspents:
					unspents[i] = u
				}

			} else {
				unspents[i] = unspent
			}
		}
	}

	return unspents, nil
}

func (e *esplora) getUnspentsForAddress(
	addr string,
	blindingKeys [][]byte,
	chUnspents chan []explorer.Utxo,
	chErr chan error,
) {
	unspents, err := e.getUnspents(addr, blindingKeys)
	if err != nil {
		chErr <- err
		return
	}
	chUnspents <- unspents
}

func (e *esplora) getUtxoDetails(
	unspent witnessUtxo,
	chUnspents chan explorer.Utxo,
	chErr chan error,
) {
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

func unblindUtxo(
	utxo explorer.Utxo,
	blindKeys [][]byte,
	chUnspents chan explorer.Utxo,
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
		unblinded, ok := transactionutil.UnblindOutput(txOut, blindKey)
		if ok {
			unspent.UAsset = unblinded.AssetHash
			unspent.UValue = unblinded.Value
			unspent.UValueBlinder = unblinded.ValueBlinder
			unspent.UAssetBlinder = unblinded.AssetBlinder
			chUnspents <- unspent
			return
		}
	}

	chErr <- errors.New("unable to unblind utxo with provided keys")
}
