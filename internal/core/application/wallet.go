package application

import (
	"github.com/tdex-network/tdex-daemon/internal/core/application/wallet"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/vulpemventures/go-elements/network"
)

type WalletService interface {
	Wallet() ports.WalletManager
	Account() ports.AccountManager
	Transaction() ports.TransactionManager
	Notification() ports.NotificationManager
	Network() network.Network
	NativeAsset() string
	SendToMany(
		account string, outs []ports.TxOutput, millisatPerByte uint64,
	) (string, error)
	CompleteSwap(
		account string, swapRequest ports.SwapRequest,
	) (string, []ports.Utxo, int64, error)
	RegisterHandlerForTxEvent(handler func(ports.WalletTxNotification) bool)
	RegisterHandlerForUtxoEvent(handler func(ports.WalletUtxoNotification) bool)
	Close()
}

func NewWalletService(w ports.OceanWallet) (WalletService, error) {
	return wallet.NewService(w)
}
