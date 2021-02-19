package elements

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/transaction"
)

func (e *elements) GetUnspents(addr string, blindKeys [][]byte) ([]explorer.Utxo, error) {
	addrLabel, err := addressLabel(addr)
	if err != nil {
		return nil, fmt.Errorf("label: %w", err)
	}

	isAddressImported, err := e.isAddressImported(addrLabel)
	if err != nil {
		return nil, fmt.Errorf("check import: %w", err)
	}

	if !isAddressImported {
		if err := e.importAddress(addr, addrLabel); err != nil {
			return nil, fmt.Errorf("import: %w", err)
		}
	}

	isBlindKeyImported, err := e.isBlindKeyImported(addr)
	if err != nil {
		return nil, fmt.Errorf("check import key: %w", err)
	}

	blindKey, err := findBlindKeyForAddress(addr, blindKeys)
	if err != nil {
		return nil, fmt.Errorf("find key: %w", err)
	}

	if !isBlindKeyImported {
		if err := e.importBlindKey(addr, blindKey); err != nil {
			return nil, fmt.Errorf("import key: %w", err)
		}
	}

	r, err := e.client.call("listunspent", []interface{}{0, 9999999, []string{addr}})
	if err = handleError(err, &r); err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	var unspents []elementsUnspent
	if err := json.Unmarshal(r.Result, &unspents); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return e.toUtxos(unspents)
}

func (e *elements) GetUnspentsForAddresses(
	addresses []string,
	blindingKeys [][]byte,
) ([]explorer.Utxo, error) {
	chUnspents := make(chan []explorer.Utxo)
	chErr := make(chan error, 1)
	unspents := make([]explorer.Utxo, 0)

	for i, addr := range addresses {
		go e.getUnspentsForAddress(addr, [][]byte{blindingKeys[i]}, chUnspents, chErr)

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

func (e *elements) getUnspentsForAddress(
	addr string,
	blindingKeys [][]byte,
	chUnspents chan []explorer.Utxo,
	chErr chan error,
) {
	unspents, err := e.GetUnspents(addr, blindingKeys)
	if err != nil {
		chErr <- err
		return
	}
	chUnspents <- unspents
}

func (e *elements) toUtxos(unspents []elementsUnspent) ([]explorer.Utxo, error) {
	utxos := make([]explorer.Utxo, 0, len(unspents))
	chUnspents := make(chan explorer.Utxo)
	chErr := make(chan error, 1)
	for _, unspent := range unspents {
		go e.getUtxoDetails(unspent, chUnspents, chErr)
		select {
		case err := <-chErr:
			close(chUnspents)
			close(chErr)
			return nil, err

		case utxo := <-chUnspents:
			if utxo != nil {
				utxos = append(utxos, utxo)
			}
		}
	}
	return utxos, nil
}

// TODO: this won't be required as soon as the wallet pkg handles blinding txs
// also with blinders instead of unblinding ALL the owned inputs with blinding
// private keys.
func (e *elements) getUtxoDetails(unspent elementsUnspent, chUnspents chan explorer.Utxo, chErr chan error) {
	txhex, err := e.GetTransactionHex(unspent.Hash())
	if err != nil {
		chErr <- err
		return
	}
	tx, err := transaction.NewTxFromHex(txhex)
	if err != nil {
		chErr <- err
		return
	}
	unspent.UNonce = tx.Outputs[unspent.Index()].Nonce
	unspent.URangeProof = tx.Outputs[unspent.Index()].RangeProof
	unspent.USurjectionProof = tx.Outputs[unspent.Index()].SurjectionProof

	chUnspents <- unspent
}

// addressLabel returns the output script in hex format relative to the
// provided address as its label. The label is used to uniquely identify the
// address within those watched by the Elements node.
// This way, it's easy to know if an address is already watched by the node,
// preventing to re-import it.
func addressLabel(addr string) (string, error) {
	script, err := address.ToOutputScript(addr)
	if err != nil {
		return "", fmt.Errorf("address is not valid")
	}
	return hex.EncodeToString(script), nil
}

func findBlindKeyForAddress(addr string, blindKeys [][]byte) (string, error) {
	data, err := address.FromConfidential(addr)
	if err != nil {
		return "", err
	}

	for _, key := range blindKeys {
		prvkey, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), key)
		prvkeyBytes := prvkey.Serialize()
		pubkeyBytes := pubkey.SerializeCompressed()
		if bytes.Equal(data.BlindingKey, pubkeyBytes) {
			return hex.EncodeToString(prvkeyBytes), nil
		}
	}

	return "", fmt.Errorf("blindkey not found for address")
}
