package vault

import (
	"github.com/btcsuite/btcutil/hdkeychain"
)

// Account defines the entity data struture for a derived account of the
// daemon's HD wallet
type Account struct {
	accountIndex           uint32
	lastExternalIndex      uint32
	lastInternalIndex      uint32
	derivationPathByScript map[string]string
}

// NewAccount returns an empty Account instance
func NewAccount(accountIndex uint32) *Account {
	return &Account{
		accountIndex:           accountIndex,
		derivationPathByScript: map[string]string{},
		lastExternalIndex:      0,
		lastInternalIndex:      0,
	}
}

// Index returns the index of the current account
func (a *Account) Index() uint32 {
	return a.accountIndex
}

// LastExternalIndex returns the last address index of external chain (0)
func (a *Account) LastExternalIndex() uint32 {
	return a.lastExternalIndex
}

// LastInternalIndex returns the last address index of internal chain (1)
func (a *Account) LastInternalIndex() uint32 {
	return a.lastInternalIndex
}

// DerivationPathByScript returns the derivation path that generates the
// provided output script
func (a *Account) DerivationPathByScript(outputScript string) (string, bool) {
	derivationPath, ok := a.derivationPathByScript[outputScript]
	return derivationPath, ok
}

// NextExternalIndex increments the last external index by one and returns the new last
func (a *Account) nextExternalIndex() uint32 {
	// restart from 0 if index has reached the its max value
	if a.lastExternalIndex == hdkeychain.HardenedKeyStart-1 {
		a.lastExternalIndex = 0
	} else {
		a.lastExternalIndex++
	}
	return a.lastExternalIndex
}

// NextInternalIndex increments the last internal index by one and returns the new last
func (a *Account) nextInternalIndex() uint32 {
	if a.lastInternalIndex == hdkeychain.HardenedKeyStart-1 {
		a.lastInternalIndex = 0
	} else {
		a.lastInternalIndex++
	}
	return a.lastInternalIndex
}

// AddDerivationPath adds an entry outputScript-derivationPath to the inner to
// the inner derivationPathByScript map
func (a *Account) addDerivationPath(outputScript, derivationPath string) {
	if _, ok := a.derivationPathByScript[outputScript]; !ok {
		a.derivationPathByScript[outputScript] = derivationPath
	}
}
