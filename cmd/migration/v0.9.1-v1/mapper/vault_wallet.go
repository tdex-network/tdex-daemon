package mapper

import (
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"
	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-domain"
)

func FromV091VaultToV1Wallet(vault v091domain.Vault) *v1domain.Wallet {
	accounts := make(map[string]*v1domain.Account)
	highestAccountIndex := 0
	for _, v := range vault.Accounts {
		if v.AccountIndex > highestAccountIndex {
			highestAccountIndex = v.AccountIndex
		}
		accounts[""] = &v1domain.Account{
			AccountInfo: v1domain.AccountInfo{
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

	return &v1domain.Wallet{
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
