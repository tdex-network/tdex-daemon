package grpcinterface

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

// bakeMacaroon creates a new macaroon with newest version and the given
// permissions then returns it binary serialized.
func bakeMacaroon(
	ctx context.Context, svc *macaroons.Service, permissions []bakery.Op,
) ([]byte, error) {
	mac, err := svc.NewMacaroon(
		ctx, macaroons.DefaultRootKeyID, permissions...,
	)
	if err != nil {
		return nil, err
	}

	return mac.M().MarshalBinary()
}

// genMacaroons generates four macaroon files; one admin-level, one for
// updating the strategy of a market, one for updating its price  and one
// read-only. Admin and read-only can also be used to generate more granular
// macaroons.
func genMacaroons(
	ctx context.Context, svc *macaroons.Service, datadir string,
) error {
	adminMacFile := filepath.Join(datadir, AdminMacaroonFile)
	roMacFile := filepath.Join(datadir, ReadOnlyMacaroonFile)
	marketMacFile := filepath.Join(datadir, MarketMacaroonFile)
	priceMacFile := filepath.Join(datadir, PriceMacaroonFile)
	if pathExists(adminMacFile) || pathExists(roMacFile) ||
		pathExists(marketMacFile) || pathExists(priceMacFile) {
		return nil
	}

	// Let's create the datadir if it doesn't exist.
	if err := makeDirectoryIfNotExists(datadir); err != nil {
		return err
	}

	for macFilename, macPermissions := range Macaroons {
		mktMacBytes, err := bakeMacaroon(ctx, svc, macPermissions)
		if err != nil {
			return err
		}
		macFile := filepath.Join(datadir, macFilename)
		perms := fs.FileMode(0644)
		if macFilename == AdminMacaroonFile {
			perms = 0600
		}
		if err := os.WriteFile(macFile, mktMacBytes, perms); err != nil {
			os.Remove(macFile)
			return err
		}
	}

	return nil
}

func makeDirectoryIfNotExists(path string) error {
	if pathExists(path) {
		return nil
	}
	return os.MkdirAll(path, os.ModeDir|0755)
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
