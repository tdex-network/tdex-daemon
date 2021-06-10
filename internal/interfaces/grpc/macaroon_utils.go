package grpcinterface

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/permissions"
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

	// First, we'll generate a macaroon that only allows the caller to
	// access invoice related calls. This is useful for merchants and other
	// services to allow an isolated instance that can only query and
	// modify invoices.
	mktMacBytes, err := bakeMacaroon(ctx, svc, permissions.MarketPermissions())
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(marketMacFile, mktMacBytes, 0644); err != nil {
		os.Remove(MarketMacaroonFile)
		return err
	}

	priceMacBytes, err := bakeMacaroon(ctx, svc, permissions.PricePermissions())
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(priceMacFile, priceMacBytes, 0644); err != nil {
		os.Remove(PriceMacaroonFile)
		return err
	}

	// Generate the read-only macaroon and write it to a file.
	roBytes, err := bakeMacaroon(ctx, svc, permissions.ReadOnlyPermissions())
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(roMacFile, roBytes, 0644); err != nil {
		os.Remove(ReadOnlyMacaroonFile)
		return err
	}

	// Generate the admin macaroon and write it to a file.
	admBytes, err := bakeMacaroon(ctx, svc, permissions.AdminPermissions())
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(adminMacFile, admBytes, 0600); err != nil {
		os.Remove(AdminMacaroonFile)
		return err
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
