package macaroons

/*
Modified from https://github.com/lightningnetwork/lnd/blob/master/macaroons/auth.go
Original Copyright 2017 Olaoluwa Osuntokun. All Rights Reserved. See LICENSE for licensing terms.
*/

import (
	"bytes"
	"context"
	"github.com/btcsuite/btcwallet/snacl"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons/kvdb"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestStore(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "macaroonstore-")
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := kvdb.Create(
		kvdb.BoltBackendName, path.Join(tempDir, "weks.db"), true,
	)
	if err != nil {
		t.Fatalf("Error opening store DB: %v", err)
	}

	store, err := NewRootKeyStorage(db)
	if err != nil {
		db.Close()
		t.Fatalf("Error creating root key store: %v", err)
	}
	defer store.Close()

	_, _, err = store.RootKey(context.TODO())
	if err != ErrStoreLocked {
		t.Fatalf("Received %v instead of ErrStoreLocked", err)
	}

	_, err = store.Get(context.TODO(), nil)
	if err != ErrStoreLocked {
		t.Fatalf("Received %v instead of ErrStoreLocked", err)
	}

	pw := []byte("weks")
	err = store.CreateUnlock(&pw)
	if err != nil {
		t.Fatalf("Error creating store encryption key: %v", err)
	}

	// Check ErrContextRootKeyID is returned when no root key ID found in
	// context.
	_, _, err = store.RootKey(context.TODO())
	if err != ErrContextRootKeyID {
		t.Fatalf("Received %v instead of ErrContextRootKeyID", err)
	}

	// Check ErrMissingRootKeyID is returned when empty root key ID is used.
	emptyKeyID := []byte{}
	badCtx := ContextWithRootKeyID(context.TODO(), emptyKeyID)
	_, _, err = store.RootKey(badCtx)
	if err != ErrMissingRootKeyID {
		t.Fatalf("Received %v instead of ErrMissingRootKeyID", err)
	}

	// Create a context with illegal root key ID value.
	encryptedKeyID := []byte("enckey")
	badCtx = ContextWithRootKeyID(context.TODO(), encryptedKeyID)
	_, _, err = store.RootKey(badCtx)
	if err != ErrKeyValueForbidden {
		t.Fatalf("Received %v instead of ErrKeyValueForbidden", err)
	}

	// Create a context with root key ID value.
	ctx := ContextWithRootKeyID(
		context.TODO(), DefaultRootKeyID,
	)
	key, id, err := store.RootKey(ctx)
	if err != nil {
		t.Fatalf("Error getting root key from store: %v", err)
	}

	rootID := id
	if !bytes.Equal(rootID, DefaultRootKeyID) {
		t.Fatalf("Root key ID doesn't match: expected %v, got %v",
			DefaultRootKeyID, rootID)
	}

	key2, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Error getting key with ID %s: %v", string(id), err)
	}
	if !bytes.Equal(key, key2) {
		t.Fatalf("Root key doesn't match: expected %v, got %v",
			key, key2)
	}

	badpw := []byte("badweks")
	err = store.CreateUnlock(&badpw)
	if err != ErrAlreadyUnlocked {
		t.Fatalf("Received %v instead of ErrAlreadyUnlocked", err)
	}

	store.Close()

	// Between here and the re-opening of the store, it's possible to get
	// a double-close, but that's not such a big deal since the tests will
	// fail anyway in that case.
	db, err = kvdb.Create(
		kvdb.BoltBackendName, path.Join(tempDir, "weks.db"), true,
	)
	if err != nil {
		t.Fatalf("Error opening store DB: %v", err)
	}

	store, err = NewRootKeyStorage(db)
	if err != nil {
		db.Close()
		t.Fatalf("Error creating root key store: %v", err)
	}

	err = store.CreateUnlock(&badpw)
	if err != snacl.ErrInvalidPassword {
		t.Fatalf("Received %v instead of ErrInvalidPassword", err)
	}

	err = store.CreateUnlock(nil)
	if err != ErrPasswordRequired {
		t.Fatalf("Received %v instead of ErrPasswordRequired", err)
	}

	_, _, err = store.RootKey(ctx)
	if err != ErrStoreLocked {
		t.Fatalf("Received %v instead of ErrStoreLocked", err)
	}

	_, err = store.Get(ctx, nil)
	if err != ErrStoreLocked {
		t.Fatalf("Received %v instead of ErrStoreLocked", err)
	}

	err = store.CreateUnlock(&pw)
	if err != nil {
		t.Fatalf("Error unlocking root key store: %v", err)
	}

	key, err = store.Get(ctx, rootID)
	if err != nil {
		t.Fatalf("Error getting key with ID %s: %v",
			string(rootID), err)
	}
	if !bytes.Equal(key, key2) {
		t.Fatalf("Root key doesn't match: expected %v, got %v",
			key2, key)
	}

	key, id, err = store.RootKey(ctx)
	if err != nil {
		t.Fatalf("Error getting root key from store: %v", err)
	}
	if !bytes.Equal(key, key2) {
		t.Fatalf("Root key doesn't match: expected %v, got %v",
			key2, key)
	}
	if !bytes.Equal(rootID, id) {
		t.Fatalf("Root ID doesn't match: expected %v, got %v",
			rootID, id)
	}
}
