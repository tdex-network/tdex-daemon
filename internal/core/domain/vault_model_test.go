package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/vulpemventures/go-elements/network"
)

func TestFailingNewVault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mnemonic      []string
		passphrase    string
		network       *network.Network
		expectedError error
	}{
		{
			nil,
			"passphrase",
			&network.Regtest,
			domain.ErrVaultNullMnemonicOrPassphrase,
		},
		{
			[]string{
				"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
				"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
				"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
			},
			"",
			&network.Regtest,
			domain.ErrVaultNullMnemonicOrPassphrase,
		},
		{
			[]string{
				"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
				"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
				"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
			},
			"passphrase",
			nil,
			domain.ErrVaultNullNetwork,
		},
	}

	for _, tt := range tests {
		v, err := domain.NewVault(tt.mnemonic, tt.passphrase, tt.network)
		require.Nil(t, v)
		require.EqualError(t, err, tt.expectedError.Error())
	}
}
