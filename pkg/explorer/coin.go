package explorer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/httputil"
)

type Service interface {
	GetUnSpents(addr string, blindKeys [][]byte) (coins []Utxo, err error)
}

type explorer struct {
	apiUrl string
}

func NewService() Service {
	return &explorer{
		apiUrl: config.GetString(config.ExplorerEndpointKey),
	}
}

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

func (e *explorer) GetUnSpents(addr string, blindKeys [][]byte) (coins []Utxo, err error) {
	url := fmt.Sprintf(
		"%s/address/%s/utxo",
		e.apiUrl,
		addr,
	)
	status, resp, err1 := httputil.NewHTTPRequest("GET", url, "", nil)
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

			if out.IsConfidential() {
				go unblindUtxo(unspent, blindKeys, chUnspents, chErr)
				select {

				case err1 := <-chErr:
					close(chErr)
					close(chUnspents)
					coins = nil
					err = fmt.Errorf("error on unblinding utxos: %s", err1)
					return

				case u := <-chUnspents:
					unspents[i] = u
				}

			} else {
				unspents[i] = unspent
			}

		}
	}
	coins = unspents

	return
}

//getCoinsIndexes method returns utxo indexes that are going to be selected
//the goal of the selection strategy is to select as less as possible utxo's
//until a 10x ratio
func getCoinsIndexes(targetAmount uint64, unblindedUtxos []Utxo) []int {
	sort.Slice(unblindedUtxos, func(i, j int) bool {
		return unblindedUtxos[i].Value() > unblindedUtxos[j].Value()
	})

	unblindedUtxosValues := []uint64{}

	for _, v := range unblindedUtxos {
		unblindedUtxosValues = append(unblindedUtxosValues, v.Value())
	}

	//actual strategy calculation output
	list := getBestCombination(unblindedUtxosValues, targetAmount)

	indexes := []int{}

	//since list variable contains values,
	//indexes holding those values needs to be calculated
	for _, v := range list {
		for i, v1 := range unblindedUtxosValues {
			if v == v1 {
				if isIndexOccupied(i, indexes) {
					continue
				} else {
					indexes = append(indexes, i)
				}
			}
		}
	}
	return indexes
}

func isIndexOccupied(i int, list []int) bool {
	for _, v := range list {
		if v == i {
			return true
		}
	}
	return false
}

var combination = []uint64{}

//getCombination is calculating all combinations for 'size' the elements of src array
// number of combination formula -> len(src)! / size! * (len(src) - size)!
func getCombination(src []uint64, size int, offset int) [][]uint64 {
	result := [][]uint64{}
	if size == 0 {
		temp := make([]uint64, len(combination))
		copy(temp, combination)
		return append(result, temp)
	}
	for i := offset; i <= len(src)-size; i++ {
		combination = append(combination, src[i])
		temp := getCombination(src, size-1, i+1)
		result = append(result, temp...)
		combination = combination[:len(combination)-1]
	}
	return result[:]
}
func sum(items []uint64) uint64 {
	var total uint64
	for _, v := range items {
		total += v
	}
	return total
}

//getBestCombination method implement strategy of selecting as less as possible
//elements from items slice so that sum of elements is equal or greater than
//target, with 10x ratio
//It uses bellow logic:
//1. set size = 1
//2. uses Recursion (getCombination) to get all combinations for size elements in the input Array.
//3. check each combination if meet the requirements from 0 -> i, if yes, return it (finish)
//4. if none of combination matches, then size++ and go to Step 2.
func getBestCombination(items []uint64, target uint64) []uint64 {
	result := [][]uint64{}
	for i := 1; i < len(items)+1; i++ {
		result = append(result, getCombination(items, i, 0)...)
		for j := 0; j < len(result); j++ {
			total := sum(result[j])
			if total < target {
				continue
			}
			if total == target {
				return result[j]
			}
			if total <= target*10 {
				return result[j]
			}
		}
	}

	//if there is no good combination just return first which is greater
	for _, v := range items {
		if v > target {
			return []uint64{v}
		}
	}

	return []uint64{}
}
