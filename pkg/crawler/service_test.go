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
	mockExplorerSvc := mockExplorer{}

	crawlSvc := NewService(Opts{
		ExplorerSvc: mockExplorerSvc,
		ErrorHandler: func(err error) {
			if err != nil {
				fmt.Println(err)
			}
		},
		CrawlerInterval:    500,
		ExplorerLimit:      10,
		ExplorerTokenBurst: 1,
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
	for {
		event := <-eventChan
		switch e := event.(type) {
		case CloseEvent:
			return
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

func stopCrawlerAfterTimeout(crawler Service) {
	crawler.Stop()
}

func removeObservableAfterTimeout(crawler Service) {
	crawler.RemoveObservable(&AddressObservable{
		AccountIndex: 0,
		Address:      "101",
	})
	time.Sleep(400 * time.Millisecond)
	crawler.RemoveObservable(NewTransactionObservable("102"))
}

func addObservableAfterTimeout(crawler Service) {
	time.Sleep(2 * time.Second)
	crawler.AddObservable(NewAddressObservable(0, "101", nil))
	time.Sleep(300 * time.Millisecond)
	crawler.AddObservable(NewTransactionObservable("102"))
}

// MOCK //

type mockExplorer struct{}

func (m mockExplorer) GetBlockHeight() (int, error) {
	panic("implement me")
}

func (m mockExplorer) GetUnspentsForAddresses(addresses []string, blindingKeys [][]byte) ([]explorer.Utxo, error) {
	panic("implement me")
}

func (m mockExplorer) IsTransactionConfirmed(txID string) (bool, error) {
	return false, nil
}

func (m mockExplorer) GetTransactionStatus(txID string) (
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

func (m mockExplorer) GetUnspents(addr string, blindKeys [][]byte) (
	[]explorer.Utxo,
	error,
) {
	if addr == "1" {
		return []explorer.Utxo{mockUtxo{value: 1}}, nil
	} else if addr == "2" {
		return []explorer.Utxo{mockUtxo{value: 2}}, nil
	} else if addr == "3" {
		return []explorer.Utxo{mockUtxo{value: 3}}, nil
	} else if addr == "101" {
		return []explorer.Utxo{mockUtxo{value: 101}}, nil
	}
	return nil, nil
}

func (m mockExplorer) GetTransaction(txID string) (explorer.Transaction, error) {
	return nil, errors.New("implement me")
}
func (m mockExplorer) GetTransactionHex(txID string) (string, error) {
	return "", errors.New("implement me")
}
func (m mockExplorer) GetTransactionsForAddress(addr string, blindingKey []byte) ([]explorer.Transaction, error) {
	return nil, errors.New("implement me")
}
func (m mockExplorer) BroadcastTransaction(txHex string) (string, error) {
	return "", errors.New("implement me")
}
func (m mockExplorer) Faucet(addr string, amount float64, asset string) (string, error) {
	return "", errors.New("implement me")
}
func (m mockExplorer) Mint(addr string, amount float64) (string, string, error) {
	return "", "", errors.New("implement me")
}

type mockUtxo struct {
	value uint64
}

func (m mockUtxo) Hash() string {
	panic("implement me")
}

func (m mockUtxo) Index() uint32 {
	panic("implement me")
}

func (m mockUtxo) Value() uint64 {
	return m.value
}

func (m mockUtxo) Asset() string {
	panic("implement me")
}

func (m mockUtxo) ValueCommitment() string {
	panic("implement me")
}

func (m mockUtxo) AssetCommitment() string {
	panic("implement me")
}

func (m mockUtxo) ValueBlinder() []byte {
	panic("implement me")
}

func (m mockUtxo) AssetBlinder() []byte {
	panic("implement me")
}

func (m mockUtxo) Nonce() []byte {
	panic("implement me")
}

func (m mockUtxo) Script() []byte {
	panic("implement me")
}

func (m mockUtxo) RangeProof() []byte {
	panic("implement me")
}

func (m mockUtxo) SurjectionProof() []byte {
	panic("implement me")
}

func (m mockUtxo) IsConfidential() bool {
	panic("implement me")
}

func (m mockUtxo) IsRevealed() bool {
	panic("implement me")
}

func (m mockUtxo) Parse() (*transaction.TxInput, *transaction.TxOutput, error) {
	panic("implement me")
}

func (m mockUtxo) IsConfirmed() bool {
	panic("implement me")
}
