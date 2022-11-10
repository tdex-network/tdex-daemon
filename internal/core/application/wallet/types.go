package wallet

import (
	"encoding/hex"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/vulpemventures/go-elements/elementsutil"
)

type input struct {
	txid   []byte
	index  uint32
	script []byte
}

func (i input) GetTxid() string {
	return elementsutil.TxIDFromBytes(i.txid)
}
func (i input) GetIndex() uint32 {
	return i.index
}
func (i input) GetScript() string {
	return hex.EncodeToString(i.script)
}
func (i input) GetScriptSigSize() int {
	return 0
}
func (i input) GetWitnessSize() int {
	return 0
}

type output struct {
	asset    string
	amount   uint64
	script   string
	blindKey string
}

func (o output) GetAsset() string {
	return o.asset
}

func (o output) GetAmount() uint64 {
	return o.amount
}

func (o output) GetScript() string {
	return o.script
}

func (o output) GetBlindingKey() string {
	return o.blindKey
}

type utxoNotificationQueue struct {
	lock *sync.Mutex
	list []func(ports.WalletUtxoNotification) bool
}

func (q *utxoNotificationQueue) len() int {
	return len(q.list)
}

func (q *utxoNotificationQueue) pushBack(
	handler func(ports.WalletUtxoNotification) bool,
) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.list = append(q.list, handler)
}

func (q *utxoNotificationQueue) pop() func(ports.WalletUtxoNotification) bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.list) <= 0 {
		return nil
	}
	handler := q.list[0]
	q.list = q.list[1:]
	return handler
}

type txNotificationQueue struct {
	lock *sync.Mutex
	list []func(ports.WalletTxNotification) bool
}

func (q *txNotificationQueue) len() int {
	return len(q.list)
}

func (q *txNotificationQueue) pushBack(
	handler func(ports.WalletTxNotification) bool,
) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.list = append(q.list, handler)
}

func (q *txNotificationQueue) pop() func(ports.WalletTxNotification) bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.list) <= 0 {
		return nil
	}
	handler := q.list[0]
	q.list = q.list[1:]
	return handler
}
