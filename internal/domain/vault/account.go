package vault

import (
	"github.com/btcsuite/btcutil/hdkeychain"
)

// Account defines the entity data struture for a derived account of the
// daemon's HD wallet
type Account struct {
	accountIndex           int
	lastExternalIndex      int
	lastInternalIndex      int
	derivationPathByScript map[string]string
}

// NewAccount returns an empty Account instance
func NewAccount(positiveAccountIndex int) (*Account, error) {
	if err := validateAccountIndex(positiveAccountIndex); err != nil {
		return nil, err
	}

	return &Account{
		accountIndex:           positiveAccountIndex,
		derivationPathByScript: map[string]string{},
		lastExternalIndex:      0,
		lastInternalIndex:      0,
	}, nil
}

// Index returns the index of the current account
func (a *Account) Index() int {
	return a.accountIndex
}

// LastExternalIndex returns the last address index of external chain (0)
func (a *Account) LastExternalIndex() int {
	return a.lastExternalIndex
}

// LastInternalIndex returns the last address index of internal chain (1)
func (a *Account) LastInternalIndex() int {
	return a.lastInternalIndex
}

// DerivationPathByScript returns the derivation path that generates the
// provided output script
func (a *Account) DerivationPathByScript(outputScript string) (string, bool) {
	derivationPath, ok := a.derivationPathByScript[outputScript]
	return derivationPath, ok
}

// NextExternalIndex increments the last external index by one and returns the new last
func (a *Account) nextExternalIndex() int {
	// restart from 0 if index has reached the its max value
	if a.lastExternalIndex == hdkeychain.HardenedKeyStart-1 {
		a.lastExternalIndex = 0
	} else {
		a.lastExternalIndex++
	}
	return a.lastExternalIndex
}

// NextInternalIndex increments the last internal index by one and returns the new last
func (a *Account) nextInternalIndex() int {
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
