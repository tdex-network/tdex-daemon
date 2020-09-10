package vault

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/btcsuite/btcutil"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/transaction"
)

var (
	// ErrMustBeLocked is thrown when trying to change the passphrase with an unlocked wallet
	ErrMustBeLocked = errors.New("wallet must be locked to perform this operation")
	// ErrMustBeUnlocked is thrown when trying to make an operation that requires the wallet to be unlocked
	ErrMustBeUnlocked = errors.New("wallet must be unlocked to perform this operation")
	// ErrInvalidPassphrase ...
	ErrInvalidPassphrase = errors.New("passphrase is not valid")
	// ErrVaultAlreadyInitialized ...
	ErrVaultAlreadyInitialized = errors.New("vault is already initialized")
)

type Vault struct {
	mnemonic          []string
	encryptedMnemonic string
	passphraseHash    []byte
	accounts          map[int]*Account
	accountsByAddress map[string]int
}

// NewVault encrypts the provided mnemonic with the passhrase and returns a new
// Vault initialized with the encrypted mnemonic and the hash of the passphrase.
// The Vault is locked by default since it is initialized without the mnemonic
// in plain text
func NewVault(mnemonic []string, passphrase string) (*Vault, error) {
	encryptedMnemonic, err := wallet.Encrypt(wallet.EncryptOpts{
		PlainText:  strings.Join(mnemonic, " "),
		Passphrase: passphrase,
	})
	if err != nil {
		return nil, err
	}

	return &Vault{
		encryptedMnemonic: encryptedMnemonic,
		passphraseHash:    btcutil.Hash160([]byte(passphrase)),
		accounts:          map[int]*Account{},
		accountsByAddress: map[string]int{},
	}, nil
}

// Lock locks the Vault by wiping its mnemonic field
func (v *Vault) Lock() error {
	if v.IsLocked() {
		return nil
	}
	// flush mnemonic in plain text
	v.mnemonic = nil
	return nil
}

// Unlock attempts to decrypt the mnemonic with the provided passphrase
func (v *Vault) Unlock(passphrase string) error {
	if !v.IsLocked() {
		return nil
	}

	mnemonic, err := wallet.Decrypt(wallet.DecryptOpts{
		CypherText: v.encryptedMnemonic,
		Passphrase: passphrase,
	})
	if err != nil {
		return err
	}

	v.mnemonic = strings.Split(mnemonic, " ")
	return nil
}

// ChangePassphrase attempts to unlock the
func (v *Vault) ChangePassphrase(currentPassphrase, newPassphrase string) error {
	if !v.IsLocked() {
		return ErrMustBeLocked
	}
	if !v.isValidPassphrase(currentPassphrase) {
		return ErrInvalidPassphrase
	}

	mnemonic, err := wallet.Decrypt(wallet.DecryptOpts{
		CypherText: v.encryptedMnemonic,
		Passphrase: currentPassphrase,
	})
	if err != nil {
		return err
	}

	encryptedMnemonic, err := wallet.Encrypt(wallet.EncryptOpts{
		PlainText:  mnemonic,
		Passphrase: newPassphrase,
	})
	if err != nil {
		return err
	}

	v.encryptedMnemonic = encryptedMnemonic
	v.passphraseHash = btcutil.Hash160([]byte(newPassphrase))
	return nil
}

// IsZero returns whether the Vault is initialized without holding any data
func (v *Vault) IsZero() bool {
	return reflect.DeepEqual(*v, Vault{})
}

// IsInitialized returnes whether the Vault has been inizitialized by checking
// if the mnemonic has been encrypted, its plain text version has been wiped
// and a passphrase (hash) has been set
func (v *Vault) IsInitialized() bool {
	return len(v.encryptedMnemonic) > 0
}

// IsLocked returns whether the Vault is initialized and locked
func (v *Vault) IsLocked() bool {
	return v.IsInitialized() && len(v.mnemonic) == 0
}

// DeriveNextExternalAddressForAccount returns the next unused address for the
// provided account and the corresponding output script
func (v *Vault) DeriveNextExternalAddressForAccount(accountIndex int) (string, string, error) {
	if v.IsLocked() {
		return "", "", ErrMustBeUnlocked
	}

	return v.deriveNextAddressForAccount(accountIndex, ExternalChain)
}

// DeriveNextInternalAddressForAccount returns the next unused change address for the
// provided account and the corresponding output script
func (v *Vault) DeriveNextInternalAddressForAccount(accountIndex int) (string, string, error) {
	if v.IsLocked() {
		return "", "", ErrMustBeUnlocked
	}

	return v.deriveNextAddressForAccount(accountIndex, InternalChain)
}

// AccountByIndex returns the account with the given index
func (v *Vault) AccountByIndex(accountIndex int) (*Account, error) {
	account, ok := v.accounts[accountIndex]
	if !ok {
		return nil, fmt.Errorf("account not found with index %d", accountIndex)
	}
	return account, nil
}

// AccountByAddress returns the account to which the provided address belongs
func (v *Vault) AccountByAddress(addr string) (*Account, int, error) {
	accountIndex, ok := v.accountsByAddress[addr]
	if !ok {
		return nil, -1, fmt.Errorf("account not found for address '%s", addr)
	}
	account, err := v.AccountByIndex(accountIndex)
	if err != nil {
		return nil, -1, err
	}
	return account, int(accountIndex), nil
}

// AllDerivedAddressesAndBlindingKeysForAccount returns all the external and
// internal addresses derived for the provided account along with the
// respective private blinding keys
func (v *Vault) AllDerivedAddressesAndBlindingKeysForAccount(accountIndex int) ([]string, [][]byte, error) {
	if v.IsLocked() {
		return nil, nil, ErrMustBeUnlocked
	}

	return v.allDerivedAddressesAndBlindingKeysForAccount(accountIndex)
}

// SendToMany creates, blinds and signs a partial transaction for sending
// different type of assets and amounts to various receivers.
// After signing the transaction, this is finalized and the final transaction
// is extracted and returned in its hex string format
func (v *Vault) SendToMany(
	accountIndex int,
	unspents []explorer.Utxo,
	outputs []*transaction.TxOutput,
	outputsBlindingKeys [][]byte,
	satsPerBytes int,
) (string, error) {
	if v.IsLocked() {
		return "", ErrMustBeUnlocked
	}

	return v.sendToMany(
		accountIndex,
		unspents,
		outputs,
		outputsBlindingKeys,
		satsPerBytes,
	)
}

func (v *Vault) isValidPassphrase(passphrase string) bool {
	return bytes.Equal(v.passphraseHash, btcutil.Hash160([]byte(passphrase)))
}

func (v *Vault) isPassphraseSet() bool {
	return len(v.passphraseHash) > 0
}

func (v *Vault) deriveNextAddressForAccount(accountIndex, chainIndex int) (string, string, error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: v.mnemonic,
	})
	if err != nil {
		return "", "", err
	}

	account, ok := v.accounts[accountIndex]
	if !ok {
		account, err = NewAccount(accountIndex)
		if err != nil {
			return "", "", err
		}
		v.accounts[accountIndex] = account
	}

	addressIndex := account.LastExternalIndex()
	if chainIndex == InternalChain {
		addressIndex = account.LastInternalIndex()
	}
	derivationPath := fmt.Sprintf(
		"%d'/%d/%d",
		account.Index(), chainIndex, addressIndex,
	)
	addr, script, err := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
		DerivationPath: derivationPath,
		Network:        config.GetNetwork(),
	})
	if err != nil {
		return "", "", err
	}
	account.addDerivationPath(hex.EncodeToString(script), derivationPath)
	if chainIndex == InternalChain {
		account.nextInternalIndex()
	} else {
		account.nextExternalIndex()
	}
	v.accountsByAddress[addr] = account.Index()

	return addr, hex.EncodeToString(script), err
}

func (v *Vault) allDerivedAddressesAndBlindingKeysForAccount(accountIndex int) ([]string, [][]byte, error) {
	account, err := v.AccountByIndex(accountIndex)
	if err != nil {
		return nil, nil, err
	}

	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: v.mnemonic,
	})
	if err != nil {
		return nil, nil, err
	}

	addresses := make([]string, 0, account.lastExternalIndex+account.lastInternalIndex)
	externalAddresses := deriveAddressesInRange(
		w,
		accountIndex,
		ExternalChain,
		0,
		account.lastExternalIndex-1,
	)
	internalAddresses := deriveAddressesInRange(
		w,
		accountIndex,
		InternalChain,
		0,
		account.lastExternalIndex-1,
	)
	addresses = append(addresses, externalAddresses...)
	addresses = append(addresses, internalAddresses...)

	blindingKeys := make([][]byte, 0, len(addresses))
	for _, addr := range addresses {
		script, _ := address.ToOutputScript(addr, *config.GetNetwork())
		key, _, _ := w.DeriveBlindingKeyPair(wallet.DeriveBlindingKeyPairOpts{
			Script: script,
		})
		blindingKeys = append(blindingKeys, key.Serialize())
	}

	return addresses, blindingKeys, nil
}

func (v *Vault) sendToMany(
	accountIndex int,
	unspents []explorer.Utxo,
	outputs []*transaction.TxOutput,
	outputsBlindingKeys [][]byte,
	milliSatsPerBytes int,
) (string, error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: v.mnemonic,
	})
	if err != nil {
		return "", nil
	}

	newPset, err := w.CreateTx()
	if err != nil {
		return "", err
	}

	account := v.accounts[accountIndex]

	changePathsByAsset := map[string]string{}
	for _, asset := range getOutputsAssets(outputs) {
		_, script, err := v.DeriveNextInternalAddressForAccount(accountIndex)
		if err != nil {
			return "", err
		}
		derivationPath, _ := account.DerivationPathByScript(script)
		changePathsByAsset[asset] = derivationPath
	}

	updateResult, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64:         newPset,
		Unspents:           unspents,
		Outputs:            outputs,
		ChangePathsByAsset: changePathsByAsset,
		MilliSatsPerBytes:  milliSatsPerBytes,
	})
	if err != nil {
		return "", err
	}

	outputsPlusChangesBlindingKeys := append(
		outputsBlindingKeys,
		updateResult.ChangeOutputsBlindingKeys...,
	)

	blindedPset, err := w.BlindTransaction(wallet.BlindTransactionOpts{
		PsetBase64:         updateResult.PsetBase64,
		OutputBlindingKeys: outputsPlusChangesBlindingKeys,
	})
	if err != nil {
		return "", err
	}

	blindedPlusFees, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64: blindedPset,
		Outputs:    createFeeOutput(updateResult.FeeAmount),
	})

	signedPset, err := w.SignTransaction(wallet.SignTransactionOpts{
		PsetBase64:        blindedPlusFees.PsetBase64,
		DerivationPathMap: account.derivationPathByScript,
	})

	return wallet.FinalizeAndExtractTransaction(
		wallet.FinalizeAndExtractTransactionOpts{
			PsetBase64: signedPset,
		},
	)
}
