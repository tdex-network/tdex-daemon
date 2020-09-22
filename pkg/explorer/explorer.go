package explorer

import (
	"errors"
	"fmt"
)

type Service interface {
	GetUnspents(addr string, blindKeys [][]byte) ([]Utxo, error)
	GetTransactionHex(hash string) (string, error)
	IsTransactionConfirmed(txID string) (bool, error)
	GetTransactionStatus(txID string) (map[string]interface{}, error)
	BroadcastTransaction(txHex string) (string, error)
	// regtest only
	Faucet(address string) (string, error)
	Mint(address string, amount int) (string, string, error)
}

type explorer struct {
	apiUrl string
}

func NewService(apiUrl string) Service {
	return &explorer{apiUrl}
}

func SelectUnspents(
	utxos []Utxo,
	blindKeys [][]byte,
	targetAmount uint64,
	targetAsset string,
) (coins []Utxo, change uint64, err error) {
	chUnspents := make(chan Utxo, len(utxos))
	chErr := make(chan error, 1)

	unblindedUtxos := make([]Utxo, 0)
	totalAmount := uint64(0)

	for i := range utxos {
		utxo := utxos[i]
		if utxo.IsConfidential() {
			go unblindUtxo(utxo, blindKeys, chUnspents, chErr)
			select {

			case err1 := <-chErr:
				close(chErr)
				close(chUnspents)
				coins = nil
				change = 0
				err = fmt.Errorf("error on unblinding utxos: %s", err1)
				return

			case unspent := <-chUnspents:
				if unspent.Asset() == targetAsset {
					unblindedUtxos = append(unblindedUtxos, unspent)
				}
			}

		} else {
			if utxo.Asset() == targetAsset {
				unblindedUtxos = append(unblindedUtxos, utxo)
			}
		}
	}

	indexes := getCoinsIndexes(targetAmount, unblindedUtxos)

	selectedUtxos := make([]Utxo, 0)
	if len(indexes) > 0 {
		for _, v := range indexes {
			totalAmount += unblindedUtxos[v].Value()
			selectedUtxos = append(selectedUtxos, unblindedUtxos[v])
		}
	} else {
		coins = nil
		change = 0
		err = errors.New(
			"error on target amount: total utxo amount does not cover target amount",
		)
		return
	}

	changeAmount := totalAmount - targetAmount
	coins = selectedUtxos
	change = changeAmount

	return
}
