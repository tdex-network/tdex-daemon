package domain

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
)

// Deposit holds info about txs with funds sent to some wallet account.
type Deposit struct {
	AccountName       string
	TxID              string
	TotAmountPerAsset map[string]uint64
	Timestamp         int64
}

func (d Deposit) Key() string {
	buf := []byte(fmt.Sprintf("%s:%s", d.AccountName, d.TxID))
	key := hex.EncodeToString(btcutil.Hash160(buf))
	return key
}
