package crawler

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/transaction"
)

func TestCrawler(t *testing.T) {
	mockExplorerSvc := MockExplorer{}

	crawlSvc := NewService(Opts{
		ExplorerSvc: mockExplorerSvc,
		ErrorHandler: func(err error) {
			if err != nil {
				fmt.Println(err)
			}
		},
		IntervalInMilliseconds: 500,
	})

	go crawlSvc.Start()
	go listen(t, crawlSvc)

	addObservableAfterTimeout(crawlSvc)
	time.Sleep(2 * time.Second)
	removeObservableAfterTimeout(crawlSvc)
	crawlSvc.Stop()
}

func listen(t *testing.T, crawlSvc Service) {
	eventChan := crawlSvc.GetEventChannel()
loop:
	for {
		select {
		case event, more := <-eventChan:
			if !more {
				break loop
			}
			switch e := event.(type) {
			// case QuitEvent:
			// 	break
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
	}

	t.Log("finished")
}

func stopCrawlerAfterTimeout(crawler Service) {
	crawler.Stop()
}

func removeObservableAfterTimeout(crawler Service) {
	crawler.RemoveObservable(&AddressObservable{
		AccountIndex: 0,
		Address:      "101",
	})
	time.Sleep(400 * time.Millisecond)
	crawler.RemoveObservable(&TransactionObservable{
		TxID: "102",
	})
}

func addObservableAfterTimeout(crawler Service) {
	time.Sleep(2 * time.Second)
	crawler.AddObservable(&AddressObservable{
		AccountIndex: 0,
		Address:      "101",
	})
	time.Sleep(300 * time.Millisecond)
	crawler.AddObservable(&TransactionObservable{
		TxID: "102",
	})
}

// MOCK //

type MockExplorer struct{}

func (m MockExplorer) GetBlockHeight() (int, error) {
	panic("implement me")
}

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
	status["confirmed"] = false

	if txID == "4" || txID == "5" || txID == "6" || txID == "102" {
		status["confirmed"] = true
		status["block_hash"] = "afbd0d4e3db10be68371b3fee107397297e9c057e3c52ee9e9a76fd62fc069a6"
		status["block_time"] = 1600178119
	}

	return status, nil
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
func (m MockExplorer) GetTransactionsForAddress(addr string) ([]explorer.Transaction, error) {
	return nil, errors.New("implement me")
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

func (m MockUtxo) ValueBlinder() []byte {
	panic("implement me")
}

func (m MockUtxo) AssetBlinder() []byte {
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

func (m MockUtxo) IsRevealed() bool {
	panic("implement me")
}

func (m MockUtxo) Parse() (*transaction.TxInput, *transaction.TxOutput, error) {
	panic("implement me")
}

func (m MockUtxo) IsConfirmed() bool {
	panic("implement me")
}
