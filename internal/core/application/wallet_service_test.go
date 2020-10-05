package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
)

var ctx = context.Background()

func newTestWallet() *walletService {
	explorerSvc := explorer.NewService(config.GetString(config.ExplorerEndpointKey))
	return newWalletService(
		inmemory.NewVaultRepositoryImpl(),
		inmemory.NewUnspentRepositoryImpl(),
		crawler.NewService(explorerSvc, []crawler.Observable{}, func(err error) {}),
		explorerSvc,
	)
}
func TestNewWalletService(t *testing.T) {
	ws := newTestWallet()
	assert.Equal(t, false, ws.walletInitialized)
	assert.Equal(t, false, ws.walletIsSyncing)
}

func TestGenSeed(t *testing.T) {
	walletSvc := newTestWallet()

	seed, err := walletSvc.GenSeed(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(seed), 24)
}

func TestInitWalletWrongSeed(t *testing.T) {
	wrongSeed := []string{"test"}
	walletSvc := newTestWallet()
	err := walletSvc.InitWallet(ctx, wrongSeed, "pass")
	assert.Error(t, err)
}

func TestInitEmptyWallet(t *testing.T) {
	mnemonic := []string{
		"curtain", "summer", "juice", "thought", "release", "velvet", "dress", "fantasy",
		"price", "hard", "core", "friend", "reopen", "myth", "giant", "consider",
		"seminar", "ladder", "thought", "spell", "state", "home", "diamond", "gold",
	}
	passphrase := "Sup3rS3cr3tP4ssw0rd!"

	w, _ := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	})
	firstWalletAccountAddr, _, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
		DerivationPath: "1'/0/0",
		Network:        config.GetNetwork(),
	})

	walletSvc := newTestWallet()
	ctx := context.Background()
	err := walletSvc.InitWallet(ctx, mnemonic, passphrase)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, walletSvc.walletInitialized)

	if err := walletSvc.UnlockWallet(ctx, passphrase); err != nil {
		t.Fatal(err)
	}
	addr, _, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, firstWalletAccountAddr, addr)
}

func TestInitUsedWallet(t *testing.T) {
	mnemonic := []string{
		"trophy", "situate", "mobile", "royal", "disease", "obvious", "ramp", "buddy",
		"turn", "robust", "trust", "company", "wheel", "adult", "produce", "spawn",
		"afford", "inspire", "topic", "farm", "sword", "embark", "body", "runway",
	}
	passphrase := "Sup3rS3cr3tP4ssw0rd!"

	walletSvc := newTestWallet()
	ctx := context.Background()

	w, _ := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	})
	mockedLastDerivedAddr, _, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
		DerivationPath: "1'/0/15",
		Network:        config.GetNetwork(),
	})
	if _, err := walletSvc.explorerService.Faucet(mockedLastDerivedAddr); err != nil {
		t.Fatal(err)
	}
	firstWalletAccountAddr, _, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
		DerivationPath: "1'/0/16",
		Network:        config.GetNetwork(),
	})

	err := walletSvc.InitWallet(ctx, mnemonic, passphrase)
	if err != nil {
		t.Fatal(err)
	}
	if err := walletSvc.UnlockWallet(ctx, passphrase); err != nil {
		t.Fatal(err)
	}
	addr, _, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, firstWalletAccountAddr, addr)
}

func TestWalletLockUnlock(t *testing.T) {
	walletSvc := newTestWallet()
	seed, err := walletSvc.GenSeed(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = walletSvc.InitWallet(ctx, seed, "pass")
	if err != nil {
		t.Fatal(err)
	}

	address, blindingKey, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	assert.Error(t, err, "wallet must be unlocked to perform this operation")

	err = walletSvc.UnlockWallet(ctx, "pass")
	if err != nil {
		t.Fatal(err)
	}

	address, blindingKey, err = walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(address) > 0, true)
	assert.Equal(t, len(blindingKey) > 0, true)
}

func TestWalletChangePass(t *testing.T) {
	walletSvc := newTestWallet()
	seed, err := walletSvc.GenSeed(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = walletSvc.InitWallet(ctx, seed, "pass")
	if err != nil {
		t.Fatal(err)
	}

	err = walletSvc.ChangePassword(
		ctx,
		"wrongPass",
		"newPass")
	assert.Error(t, err, "passphrase is not valid")

	err = walletSvc.ChangePassword(
		ctx,
		"pass",
		"newPass")
	assert.NoError(t, err)

	err = walletSvc.UnlockWallet(ctx, "newPass")
	assert.NoError(t, err)
}

func TestWalletBalance(t *testing.T) {
	walletSvc := newTestWallet()
	seed, err := walletSvc.GenSeed(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = walletSvc.InitWallet(ctx, seed, "pass")
	if err != nil {
		t.Fatal(err)
	}

	err = walletSvc.UnlockWallet(ctx, "pass")
	if err != nil {
		t.Fatal(err)
	}

	address, _, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = walletSvc.explorerService.Faucet(address)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	balance, err := walletSvc.GetWalletBalance(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(
		t,
		100000000,
		int(balance[config.GetString(config.BaseAssetKey)].ConfirmedBalance),
	)
}
