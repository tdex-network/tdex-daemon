package crawler

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/transaction"
)

func TestCrawler(t *testing.T) {
	mockExplorerSvc := MockExplorer{}

	observables := make([]Observable, 0)
	for i := 0; i < 100; i++ {
		adrObservable := AddressObservable{
			AccountIndex: 1,
			Address:      strconv.Itoa(i),
		}
		trxObservable := TransactionObservable{
			TxID: strconv.Itoa(i),
		}
		observables = append(observables, &adrObservable)
		observables = append(observables, &trxObservable)
	}

	crawlSvc := NewService(mockExplorerSvc, observables, nil)

	go crawlSvc.Start()

	go removeObservableAfterTimeout(crawlSvc)

	go addObservableAfterTimeout(crawlSvc)

	go stopCrawlerAfterTimeout(crawlSvc)

	for event := range crawlSvc.GetEventChannel() {

		switch e := event.(type) {
		case AddressEvent:
			for _, u := range e.Utxos {
				t.Log(fmt.Sprintf("%v %v %v", "ADR", e.EventType, u.Value()))
			}
		case TransactionEvent:
			if e.EventType == TransactionConfirmed {
				t.Log(fmt.Sprintf("%v %v %v", "TX", e.EventType, e.TxID))
			}
		}

	}

	t.Log("finished")

}

func stopCrawlerAfterTimeout(crawler Service) {
	time.Sleep(7 * time.Second)
	crawler.Stop()
}

func removeObservableAfterTimeout(crawler Service) {
	time.Sleep(2 * time.Second)
	crawler.RemoveObservable(&AddressObservable{
		AccountIndex: 0,
		Address:      "2",
	})
	crawler.RemoveObservable(&TransactionObservable{
		TxID: "5",
	})
}

func addObservableAfterTimeout(crawler Service) {
	time.Sleep(5 * time.Second)
	crawler.AddObservable(&AddressObservable{
		AccountIndex: 0,
		Address:      "101",
	})
	crawler.AddObservable(&TransactionObservable{
		TxID: "102",
	})
}

// MOCK //

type MockExplorer struct{}

func (m MockExplorer) GetUnspentsForAddresses(addresses []string, blindingKeys [][]byte) ([]explorer.Utxo, error) {
	panic("implement me")
}

func (m MockExplorer) IsTransactionConfirmed(txID string) (bool, error) {
	return false, nil
}

func (m MockExplorer) GetTransactionStatus(txID string) (
	map[string]interface{},
	error,
) {
	status := make(map[string]interface{}, 0)
	status["confirmed"] = true
	status["block_hash"] = "afbd0d4e3db10be68371b3fee107397297e9c057e3c52ee9e9a76fd62fc069a6"
	status["block_time"] = 1600178119

	if txID == "4" {
		return status, nil
	} else if txID == "5" {
		return status, nil
	} else if txID == "6" {
		return status, nil
	} else if txID == "102" {
		return status, nil
	}
	return nil, nil
}

func (m MockExplorer) GetUnspents(addr string, blindKeys [][]byte) (
	[]explorer.Utxo,
	error,
) {
	if addr == "1" {
		return []explorer.Utxo{MockUtxo{value: 1}}, nil
	} else if addr == "2" {
		return []explorer.Utxo{MockUtxo{value: 2}}, nil
	} else if addr == "3" {
		return []explorer.Utxo{MockUtxo{value: 3}}, nil
	} else if addr == "101" {
		return []explorer.Utxo{MockUtxo{value: 101}}, nil
	}
	return nil, nil
}

func (m MockExplorer) GetTransactionHex(txID string) (string, error) {
	return "", errors.New("implement me")
}
func (m MockExplorer) BroadcastTransaction(txHex string) (string, error) {
	return "", errors.New("implement me")
}
func (m MockExplorer) Faucet(addr string) (string, error) {
	return "", errors.New("implement me")
}
func (m MockExplorer) Mint(addr string, amount int) (string, string, error) {
	return "", "", errors.New("implement me")
}

type MockUtxo struct {
	value uint64
}

func (m MockUtxo) Hash() string {
	panic("implement me")
}

func (m MockUtxo) Index() uint32 {
	panic("implement me")
}

func (m MockUtxo) Value() uint64 {
	return m.value
}

func (m MockUtxo) Asset() string {
	panic("implement me")
}

func (m MockUtxo) ValueCommitment() string {
	panic("implement me")
}

func (m MockUtxo) AssetCommitment() string {
	panic("implement me")
}

func (m MockUtxo) Nonce() []byte {
	panic("implement me")
}

func (m MockUtxo) Script() []byte {
	panic("implement me")
}

func (m MockUtxo) RangeProof() []byte {
	panic("implement me")
}

func (m MockUtxo) SurjectionProof() []byte {
	panic("implement me")
}

func (m MockUtxo) IsConfidential() bool {
	panic("implement me")
}

func (m MockUtxo) SetScript(script []byte) {
	panic("implement me")
}

func (m MockUtxo) SetUnconfidential(asset string, value uint64) {
	panic("implement me")
}

func (m MockUtxo) SetConfidential(nonce, rangeProof, surjectionProof []byte) {
	panic("implement me")
}

func (m MockUtxo) Parse() (*transaction.TxInput, *transaction.TxOutput, error) {
	panic("implement me")
}

func (m MockUtxo) IsConfirmed() bool {
	panic("implement me")
}
