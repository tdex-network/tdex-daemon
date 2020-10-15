package dbbadger

import (
	"github.com/magiconair/properties/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"testing"
)

func TestGetOrCreateVault(t *testing.T) {
	before()
	defer after()

	var addr string

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

	assert.Equal(t, vault.Accounts[domain.FeeAccount].LastExternalIndex, 1)

	account, err := vaultRepository.GetAccountByIndex(ctx, domain.FeeAccount)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, account.LastExternalIndex, 1)

	accnt, accntIndex, err := vaultRepository.GetAccountByAddress(ctx, addr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, accnt.LastExternalIndex, 1)
	assert.Equal(t, accntIndex, domain.FeeAccount)

	addresses, _, err := vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(
			ctx,
			domain.FeeAccount,
		)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, addresses[0], addr)

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
	assert.Equal(t, pathByScript[script], path)
}
