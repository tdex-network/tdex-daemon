package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/sony/gobreaker"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
)

var (
	url = "https://blockstream.info/liquid/api"
	// url = "http://localhost:3001"
	addr = "lq1qqvtxnetvuqy03gqfac53j8zgwamx02m904cjxv5fk9c50mlxfhcvuppa78agt4qvkq35dwasg0a4lx6k9fn3gr0clhyf88fq8"
	// addr           = "el1qqwj0wxrp22tgrm8vznwrk3jnswqtf33jk3q89x82e3ndalfzc45zg5rqcaze4jsde9fnar6qa7asn4l35flkvfd648tz2zjr3"
	explorerSvc, _ = esplora.NewService(url, 15000)
	attempts       = 345
)

type utxoResult struct {
	utxos []explorer.Utxo
	err   error
	i     int
}

func getUtxos(cb *gobreaker.CircuitBreaker) ([]explorer.Utxo, error) {
	utxos := make([]explorer.Utxo, 0)
	chRes := make(chan utxoResult)
	wg := &sync.WaitGroup{}
	wg.Add(attempts)

	go func() {
		wg.Wait()
		close(chRes)
	}()

	for i := 0; i < attempts; i++ {
		go getUtxo(cb, i, chRes, wg)
	}

	for r := range chRes {
		if r.err != nil {
			return nil, fmt.Errorf("error on %d call %s", r.i, r.err)
		}

		utxos = append(utxos, r.utxos...)
	}

	return utxos, nil
}

func getUtxo(cb *gobreaker.CircuitBreaker, i int, chRes chan utxoResult, wg *sync.WaitGroup) {
	defer wg.Done()

	iUtxos, err := cb.Execute(func() (interface{}, error) {
		return explorerSvc.GetUnspents(addr, nil)
	})

	if err != nil {
		chRes <- utxoResult{i: i, err: err}
		return
	}
	utxos := iUtxos.([]explorer.Utxo)

	chRes <- utxoResult{i: i, utxos: utxos}
}

func main() {
	cb := newCircuitBreaker()
	utxos, err := getUtxos(cb)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	fmt.Println(len(utxos))
}

func newCircuitBreaker() *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "explorer",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			fmt.Printf("%+v\n", counts)
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			fmt.Println("fr: ", failureRatio)
			return counts.Requests > 20 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			if to == gobreaker.StateOpen {
				fmt.Println("WARN: explorer seems down, stop allowing requests")
			}
			if from == gobreaker.StateOpen && to == gobreaker.StateHalfOpen {
				fmt.Println("LOG: checking explorer status")
			}
			if from == gobreaker.StateHalfOpen && to == gobreaker.StateClosed {
				fmt.Println("LOG: explorer seems ok, restart allowing requests")
			}
		},
	})
}
