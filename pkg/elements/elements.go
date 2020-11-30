package elements

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type elements struct {
	client *RpcClient
}

// NewService returns and Elements based explorer
func NewService(host string, port int, user, passwd string) (explorer.Service, error) {
	client, err := NewClient(host, port, user, passwd, false, 30)
	if err != nil {
		return nil, err
	}

	return &elements{client}, nil
}

func (e *elements) GetUnspents(addr string, blindKeys [][]byte) (utxos []explorer.Utxo, err error) {
	r, err := e.client.call("getblockchaininfo", nil)
	if err = handleError(err, &r); err != nil {
		return nil, fmt.Errorf("info: %w", err)
	}
	r, err = e.client.call("importaddress", []string{addr})
	if err = handleError(err, &r); err != nil {
		return nil, fmt.Errorf("addr: %w", err)
	}
	r, err = e.client.call("importblindingkey", []string{addr, hex.EncodeToString(blindKeys[0])})
	if err = handleError(err, &r); err != nil {
		return nil, fmt.Errorf("blind: %w", err)
	}
	r, err = e.client.call("listunspent", []interface{}{1, 9999999, []string{addr}})
	if err = handleError(err, &r); err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	var unspents []Unspent
	err = json.Unmarshal(r.Result, &unspents)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	for _, un := range unspents {
		amountInSatoshis := uint64(un.Amount * math.Pow10(8))
		scriptPubKey, _ := hex.DecodeString(un.ScriptPubKey)

		utxos = append(utxos, explorer.NewUnconfidentialWitnessUtxo(
			un.TxID,
			un.Vout,
			amountInSatoshis,
			un.Asset,
			scriptPubKey,
		))
	}

	return
}

func (e *elements) GetUnspentsForAddresses(
	addresses []string,
	blindingKeys [][]byte,
) ([]explorer.Utxo, error) {
	return nil, nil
}

func (e *elements) GetTransactionHex(hash string) (string, error) {
	return "", nil
}

func (e *elements) IsTransactionConfirmed(txID string) (bool, error) {
	return false, nil
}

func (e *elements) GetTransactionStatus(txID string) (map[string]interface{}, error) {
	return nil, nil
}
func (e *elements) GetTransactionsForAddress(address string) ([]explorer.Transaction, error) {
	return nil, nil
}

func (e *elements) BroadcastTransaction(txHex string) (string, error) {
	return "", nil
}

// Regtest only
func (e *elements) Faucet(address string) (string, error) {
	return "", nil
}
func (e *elements) Mint(address string, amount int) (string, string, error) {
	return "", "", nil
}
