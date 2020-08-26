package explorer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/util"
	"net/http"
	"sort"
)

func SelectUnSpents(
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

	sort.Slice(unblindedUtxos, func(i, j int) bool {
		return unblindedUtxos[i].Value() > unblindedUtxos[j].Value()
	})

	selected := 0
	for _, u := range unblindedUtxos {
		if u.Asset() == targetAsset {
			selected++
			totalAmount += u.Value()
		}
		if totalAmount >= targetAmount {
			break
		}
	}

	if totalAmount < targetAmount {
		coins = nil
		change = 0
		err = errors.New(
			"error on target amount: total utxo amount does not cover target amount",
		)
		return
	}

	changeAmount := totalAmount - targetAmount

	coins = unblindedUtxos[:selected]
	change = changeAmount

	return
}

func GetUnSpents(addr string) ([]Utxo, error) {
	url := fmt.Sprintf(
		"%s/address/%s/utxo",
		config.GetString(config.ExplorerEndpointKey),
		addr,
	)
	status, resp, err := util.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return nil, fmt.Errorf("error on retrieving utxos: %s", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf(resp)
	}

	var witnessOuts []witnessUtxo
	err = json.Unmarshal([]byte(resp), &witnessOuts)
	if err != nil {
		return nil, fmt.Errorf("error on retrieving utxos: %s", err)
	}

	unspents := make([]Utxo, len(witnessOuts))
	chUnspents := make(chan Utxo, len(witnessOuts))
	chErr := make(chan error, 1)

	for i := range witnessOuts {
		out := witnessOuts[i]
		go getUtxoDetails(out, chUnspents, chErr)
		select {
		case err := <-chErr:
			if err != nil {
				close(chErr)
				close(chUnspents)
				return nil, fmt.Errorf("error on retrieving utxos: %s", err)
			}
		case unspent := <-chUnspents:
			unspents[i] = unspent
		}
	}

	return unspents, nil
}
