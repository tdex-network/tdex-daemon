package v091domain

import (
	"github.com/vulpemventures/go-elements/network"
)

type AccountAndKey struct {
	AccountIndex int
	BlindingKey  []byte
}

type Vault struct {
	EncryptedMnemonic      string
	PassphraseHash         []byte
	Accounts               map[int]*Account
	AccountAndKeyByAddress map[string]AccountAndKey
	Network                *network.Network
}

type Account struct {
	AccountIndex           int
	LastExternalIndex      int
	LastInternalIndex      int
	DerivationPathByScript map[string]string
}

type AddressInfo struct {
	AccountIndex   int
	Address        string
	BlindingKey    []byte
	DerivationPath string
	Script         string
}
