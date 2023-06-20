package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	t.Skip()

	v091DataDir := "v091-datadir"
	v1OceanDataDir := "v1-oceandatadir"
	v1TdexdDataDir := "v1-datadir"
	v091VaultPassword := "ciaociao"
	err := migrate(v091DataDir, v1OceanDataDir, v1TdexdDataDir, v091VaultPassword)
	require.NoError(t, err)
}
