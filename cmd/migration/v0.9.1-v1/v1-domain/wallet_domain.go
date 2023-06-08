package v1domain

import (
	"context"
	"fmt"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	"github.com/dgraph-io/badger/v2/options"

	"github.com/dgraph-io/badger/v2"
	"github.com/timshannon/badgerhold/v2"
)

const (
	walletKey = "wallet"
)

type AddressInfo struct {
	Account        string
	Address        string
	BlindingKey    []byte
	DerivationPath string
	Script         string
}

type Wallet struct {
	EncryptedMnemonic   []byte
	PasswordHash        []byte
	BirthdayBlockHeight uint32
	RootPath            string
	NetworkName         string
	Accounts            map[string]*Account
	AccountsByLabel     map[string]string
	NextAccountIndex    uint32
}

type Account struct {
	AccountInfo
	Index                  uint32
	BirthdayBlock          uint32
	NextExternalIndex      uint
	NextInternalIndex      uint
	DerivationPathByScript map[string]string
}

type AccountInfo struct {
	Namespace      string
	Label          string
	Xpub           string
	DerivationPath string
}

type WalletRepository interface {
	InsertWallet(ctx context.Context, wallet *Wallet) error
}

type walletRepository struct {
	store *badgerhold.Store
}

func NewWalletRepository(dbDir string, logger badger.Logger) (WalletRepository, error) {
	opts := badger.DefaultOptions(dbDir)
	opts.Logger = logger
	opts.Compression = options.ZSTD

	store, err := badgerhold.Open(badgerhold.Options{
		Encoder:          badgerhold.DefaultEncode,
		Decoder:          badgerhold.DefaultDecode,
		SequenceBandwith: 100,
		Options:          opts,
	})
	if err != nil {
		return nil, err
	}

	return &walletRepository{store: store}, nil
}

func (r *walletRepository) InsertWallet(
	ctx context.Context, wallet *Wallet,
) error {
	var err error

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = r.store.TxInsert(tx, walletKey, *wallet)
	} else {
		err = r.store.Insert(walletKey, *wallet)
	}
	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return fmt.Errorf("wallet is already initialized")
		}
		return err
	}

	return nil
}

func FromV091VaultToV1Wallet(vault domain.Vault) *Wallet {
	accounts := make(map[string]*Account)
	highestAccountIndex := 0
	for _, v := range vault.Accounts {
		if v.AccountIndex > highestAccountIndex {
			highestAccountIndex = v.AccountIndex
		}
		accounts[""] = &Account{
			AccountInfo: AccountInfo{
				Namespace:      "",
				Label:          "",
				Xpub:           "",
				DerivationPath: "",
			},
			Index:                  uint32(v.AccountIndex),
			BirthdayBlock:          0, // TODO check block height of tdexd start ?
			NextExternalIndex:      uint(v.LastExternalIndex + 1),
			NextInternalIndex:      uint(v.LastInternalIndex + 1),
			DerivationPathByScript: v.DerivationPathByScript,
		}
	}

	return &Wallet{
		EncryptedMnemonic:   []byte(vault.EncryptedMnemonic),
		PasswordHash:        vault.PassphraseHash,
		BirthdayBlockHeight: 0,  // TODO check block height of tdexd start ?
		RootPath:            "", // ?
		NetworkName:         vault.Network.Name,
		Accounts:            accounts,
		AccountsByLabel:     nil, // ?
		NextAccountIndex:    uint32(highestAccountIndex + 1),
	}
}
