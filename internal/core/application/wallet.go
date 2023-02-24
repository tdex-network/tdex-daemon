package application

import (
	"github.com/tdex-network/tdex-daemon/internal/core/application/wallet"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/vulpemventures/go-elements/network"
)

type WalletService interface {
	Wallet() ports.Wallet
	Account() ports.Account
	Transaction() ports.Transaction
	Notification() ports.Notification
	Network() network.Network
	NativeAsset() string
	SendToMany(
		account string, outs []ports.TxOutput, msatsPerByte uint64,
	) (string, error)
	CompleteSwap(
		account string, swapRequest ports.SwapRequest, msatsPerByte uint64,
	) (string, []ports.Utxo, int64, error)
	RegisterHandlerForTxEvent(handler func(ports.WalletTxNotification) bool)
	RegisterHandlerForUtxoEvent(handler func(ports.WalletUtxoNotification) bool)
	Close()
}

func NewWalletService(w ports.WalletService) (WalletService, error) {
	return wallet.NewService(w)
}
