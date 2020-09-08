package vault

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcutil/hdkeychain"
)

const (
	// FeeAccountPath ...
	FeeAccountPath = iota
	// WalletAccountPath ...
	WalletAccountPath
	// FirstUnusedAccountPath ...
	FirstUnusedAccountPath
	// SecondUnusedAccountPath ...
	SecondUnusedAccountPath
	// FirstMarketAccountPath ...
	FirstMarketAccountPath
)

var (
	// ErrEmptyKeyDerivationMap ...
	ErrEmptyKeyDerivationMap = errors.New("no keys derived")
	// ErrAccountIndexOutOfRange ...
	ErrAccountIndexOutOfRange = errors.New("index is out of range of derived accounts")
)

// Account defines the entity data struture for a derived account of the
// daemon's HD wallet
type Account struct {
	lastExternalIndex      uint32
	lastInternalIndex      uint32
	derivationPathByScript map[string]string
}

// NewAccount returns an empty Account instance
func NewAccount() *Account {
	return &Account{
		derivationPathByScript: map[string]string{},
		lastExternalIndex:      0,
		lastInternalIndex:      0,
	}
}

// LastExternalIndex returns the last address index of external chain (0)
func (a *Account) LastExternalIndex() uint32 {
	return a.lastExternalIndex
}

// LastInternalIndex returns the last address index of internal chain (1)
func (a *Account) LastInternalIndex() uint32 {
	return a.lastInternalIndex
}

// NextExternalIndex increments the last external index by one and returns the new last
func (a *Account) NextExternalIndex() uint32 {
	// restart from 0 if index has reached the its max value
	if a.lastExternalIndex == hdkeychain.HardenedKeyStart-1 {
		a.lastExternalIndex = 0
	} else {
		a.lastExternalIndex++
	}
	return a.lastExternalIndex
}

// NextInternalIndex increments the last internal index by one and returns the new last
func (a *Account) NextInternalIndex() uint32 {
	if a.lastInternalIndex == hdkeychain.HardenedKeyStart-1 {
		a.lastInternalIndex = 0
	} else {
		a.lastInternalIndex++
	}
	return a.lastInternalIndex
}

// AddDerivationPath adds an entry outputScript-derivationPath to the inner to
// the inner derivationPathByScript map
func (a *Account) AddDerivationPath(outputScript, derivationPath string) {
	if _, ok := a.derivationPathByScript[outputScript]; !ok {
		a.derivationPathByScript[outputScript] = derivationPath
	}
}

// DerivationPathForScript returns the derivation path that generates the
// provided output script
func (a *Account) DerivationPathForScript(outputScript string) (string, error) {
	derivationPath, ok := a.derivationPathByScript[outputScript]
	if !ok {
		return "", fmt.Errorf("derivation path not found for output script '%s'", outputScript)
	}
	return derivationPath, nil
}
