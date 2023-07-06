package mapper

import (
	"encoding/base64"
	"fmt"

	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0-v1/v1-domain"
	v0domain "github.com/tdex-network/tdex-daemon/old-v0"
)

const (
	nameSpaceFormat = "bip84-account%d"
)

func (m *mapperService) FromV091VaultToV1Wallet(
	vault v0domain.Vault, password string,
) (*v1domain.Wallet, error) {
	accounts := make(map[string]*v1domain.Account)
	accountsByLabel := make(map[string]string)
	highestAccountIndex := 0
	masterKey, err := vault.MasterKey(password)
	if err != nil {
		return nil, err
	}

	for _, v := range vault.Accounts {
		label, err := m.getLabel(v.AccountIndex)
		if err != nil {
			return nil, err
		}

		namespace := fmt.Sprintf(nameSpaceFormat, v.AccountIndex)
		accountsByLabel[label] = namespace

		if v.AccountIndex > highestAccountIndex {
			highestAccountIndex = v.AccountIndex
		}

		xpub, err := v0domain.Xpub(uint32(v.AccountIndex), masterKey)
		if err != nil {
			return nil, err
		}

		accounts[namespace] = &v1domain.Account{
			AccountInfo: v1domain.AccountInfo{
				Namespace: namespace,
				Label:     label,
				Xpub:      xpub,
				DerivationPath: fmt.Sprintf(
					"%s/%d'", v0domain.RootPath, v.AccountIndex,
				),
			},
			Index:                  uint32(v.AccountIndex),
			BirthdayBlock:          0,
			NextExternalIndex:      uint(v.LastExternalIndex + 1),
			NextInternalIndex:      uint(v.LastInternalIndex + 1),
			DerivationPathByScript: v.DerivationPathByScript,
		}
	}

	encryptedMnemonic, _ := base64.StdEncoding.DecodeString(
		vault.EncryptedMnemonic,
	)

	return &v1domain.Wallet{
		EncryptedMnemonic:   encryptedMnemonic,
		PasswordHash:        vault.PassphraseHash,
		BirthdayBlockHeight: 0,
		RootPath:            v0domain.RootPath,
		NetworkName:         vault.Network.Name,
		Accounts:            accounts,
		AccountsByLabel:     accountsByLabel,
		NextAccountIndex:    uint32(highestAccountIndex + 1),
	}, nil
}

func (m *mapperService) getLabel(accountIndex int) (string, error) {
	switch accountIndex {
	case v0domain.FeeAccount:
		return "fee_account", nil
	case v0domain.WalletAccount:
		return "personal_account", nil
	case v0domain.FeeFragmenterAccount:
		return "fee_fragmenter_account", nil
	case v0domain.MarketFragmenterAccount:
		return "market_fragmenter_account", nil
	default:
		market, err := m.v0RepoManager.MarketRepository().
			GetMarketByAccount(accountIndex)
		if err != nil {
			return "", err
		}

		return market.AccountName(), nil
	}
}
