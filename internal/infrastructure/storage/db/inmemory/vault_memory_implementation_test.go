package inmemory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func TestAll(t *testing.T) {
	db := newMockDb()
	vaultRepository := NewVaultRepositoryImpl(db)

	var addr string

	config.Set(config.MnemonicKey,
		"leave dice fine decrease dune ribbon ocean earn lunar account silver"+
			" admit cheap fringe disorder trade because trade steak clock grace video jacket equal")

	if err := vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			a, _, _, err := v.DeriveNextExternalAddressForAccount(
				domain.FeeAccount,
			)
			if err != nil {
				return nil, err
			}

			addr = a

			return v, nil
		},
	); err != nil {
		t.Fatal(err)
	}

	vault, err := vaultRepository.GetOrCreateVault(
		ctx,
		nil,
		"",
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, vault.Accounts[domain.FeeAccount].LastExternalIndex)

	account, err := vaultRepository.GetAccountByIndex(ctx, domain.FeeAccount)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, account.LastExternalIndex)

	accnt, accntIndex, err := vaultRepository.GetAccountByAddress(ctx, addr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, accnt.LastExternalIndex)
	assert.Equal(t, domain.FeeAccount, accntIndex)

	addresses, _, err := vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(
			ctx,
			domain.FeeAccount,
		)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, addr, addresses[0])

	var script string
	var path string
	for k, v := range account.DerivationPathByScript {
		script = k
		path = v
	}

	pathByScript, err := vaultRepository.GetDerivationPathByScript(
		ctx,
		domain.FeeAccount,
		[]string{script},
	)
	assert.Equal(t, path, pathByScript[script])
}
