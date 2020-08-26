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

func GetUnSpents(addr string) (coins []Utxo, err error) {
	url := fmt.Sprintf(
		"%s/address/%s/utxo",
		config.GetString(config.ExplorerEndpointKey),
		addr,
	)
	status, resp, err1 := util.NewHTTPRequest("GET", url, "", nil)
	if err1 != nil {
		coins = nil
		err = fmt.Errorf("error on retrieving utxos: %s", err)
		return
	}
	if status != http.StatusOK {
		coins = nil
		err = fmt.Errorf(resp)
		return
	}

	var witnessOuts []witnessUtxo
	err1 = json.Unmarshal([]byte(resp), &witnessOuts)
	if err1 != nil {
		coins = nil
		err = fmt.Errorf("error on retrieving utxos: %s", err)
		return
	}

	unspents := make([]Utxo, len(witnessOuts))
	chUnspents := make(chan Utxo, len(witnessOuts))
	chErr := make(chan error, 1)

	for i := range witnessOuts {
		out := witnessOuts[i]
		go getUtxoDetails(out, chUnspents, chErr)
		select {
		case err1 := <-chErr:
			if err1 != nil {
				close(chErr)
				close(chUnspents)
				coins = nil
				err = fmt.Errorf("error on retrieving utxos: %s", err1)
				return
			}
		case unspent := <-chUnspents:
			unspents[i] = unspent
		}
	}
	coins = unspents

	return
}
