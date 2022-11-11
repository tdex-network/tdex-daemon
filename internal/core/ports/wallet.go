package ports

import "context"

type WalletService interface {
	Wallet() Wallet
	Account() Account
	Transaction() Transaction
	Notification() Notification
	Close()
}

type Wallet interface {
	GenSeed(ctx context.Context) ([]string, error)
	InitWallet(ctx context.Context, mnemonic []string, password string) error
	RestoreWallet(ctx context.Context, mnemonic []string, password string) error
	Unlock(ctx context.Context, password string) error
	Lock(ctx context.Context, password string) error
	ChangePassword(ctx context.Context, oldPwd, newPwd string) error
	Status(ctx context.Context) (WalletStatus, error)
	Info(ctx context.Context) (WalletInfo, error)
}

type Account interface {
	CreateAccount(ctx context.Context, accountName string) (WalletAccount, error)
	DeriveAddresses(
		ctx context.Context, accountName string, num int,
	) ([]string, error)
	DeriveChangeAddresses(
		ctx context.Context, accountName string, num int,
	) ([]string, error)
	ListAddresses(ctx context.Context, accountName string) ([]string, error)
	GetBalance(
		ctx context.Context, accountName string,
	) (map[string]Balance, error)
	ListUtxos(
		ctx context.Context, accountName string,
	) ([]Utxo, []Utxo, error)
	DeleteAccount(ctx context.Context, accountName string) error
}

type Transaction interface {
	GetTransaction(ctx context.Context, txid string) (string, error)
	EstimateFees(
		ctx context.Context, ins []TxInput, outs []TxOutput,
	) (uint64, error)
	SelectUtxos(
		ctx context.Context, accountName, asset string, amount uint64,
	) ([]Utxo, uint64, int64, error)
	CreatePset(
		ctx context.Context, ins []TxInput, outs []TxOutput,
	) (string, error)
	UpdatePset(
		ctx context.Context, pset string,
		ins []TxInput, outs []TxOutput,
	) (string, error)
	BlindPset(
		ctx context.Context, pset string, extraUnblindedIns []UnblindedInput,
	) (string, error)
	SignPset(
		ctx context.Context, pset string, extractRawTx bool,
	) (string, error)
	Transfer(
		ctx context.Context,
		accountName string, outs []TxOutput, millisatsPerByte uint64,
	) (string, error)
	BroadcastTransaction(ctx context.Context, txHex string) (string, error)
}

type Notification interface {
	GetTxNotifications() chan WalletTxNotification
	GetUtxoNotifications() chan WalletUtxoNotification
}

type WalletStatus interface {
	IsInitialized() bool
	IsUnlocked() bool
	IsSynced() bool
}

type WalletInfo interface {
	GetNetwork() string
	GetNativeAsset() string
	GetRootPath() string
	GetMasterBlindingKey() string
	GetBirthdayBlockHash() string
	GetBirthdayBlockHeight() uint32
	GetAccounts() []WalletAccount
}

type WalletAccount interface {
	GetName() string
	GetDerivationPath() string
	GetXpub() string
}

type WalletTxNotification interface {
	GetEventType() WalletTxEventType
	GetAccountNames() []string
	GetTxHex() string
	GetBlockDetails() BlockInfo
}

type WalletUtxoNotification interface {
	GetEventType() WalletUtxoEventType
	GetUtxos() []Utxo
}

type WalletTxEventType interface {
	IsUnconfirmed() bool
	IsConfirmed() bool
	IsBroadcasted() bool
}

type WalletUtxoEventType interface {
	IsUnconfirmed() bool
	IsConfirmed() bool
	IsLocked() bool
	IsUnlocked() bool
	IsSpent() bool
}
