package crawler

import (
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/transaction"
	"strconv"
	"testing"
	"time"
)

func TestCrawler(t *testing.T) {
	mockExplorerSvc := MockExplorer{}

	observables := make([]Observable, 0)
	for i := 0; i < 100; i++ {
		observable := Observable{
			accountType: 1,
			address:     strconv.Itoa(i),
		}
		observables = append(observables, observable)
	}

	crawlSvc := NewService(mockExplorerSvc, observables)

	go crawlSvc.Start()

	go removeObservableAfterTimeout(crawlSvc)

	go addObservableAfterTimeout(crawlSvc)

	go stopCrawlerAfterTimeout(crawlSvc)

	for event := range crawlSvc.GetEventChannel() {
		t.Log(event.utxo)
	}

	t.Log("finished")

}

func stopCrawlerAfterTimeout(crawler Service) {
	time.Sleep(8 * time.Second)
	crawler.Stop()
}

func removeObservableAfterTimeout(crawler Service) {
	time.Sleep(2 * time.Second)
	crawler.RemoveObservable(Observable{
		accountType: 0,
		address:     "2",
	})
}

func addObservableAfterTimeout(crawler Service) {
	time.Sleep(5 * time.Second)
	crawler.AddObservable(Observable{
		accountType: 0,
		address:     "101",
	})
}

// MOCK //

type MockExplorer struct{}

func (m MockExplorer) GetUnSpents(addr string) ([]explorer.Utxo, error) {
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
