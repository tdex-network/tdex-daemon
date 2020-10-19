package application

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/btcsuite/btcutil"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
)

type mockedWallet struct {
	mnemonic          []string
	encryptedMnemonic string
	password          string
}

var (
	dryWallet = &mockedWallet{
		mnemonic: []string{
			"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
			"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
			"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
		},
		password:          "pass",
		encryptedMnemonic: "dVoBFte1oeRkPl8Vf8DzBP3PRnzPA3fxtyvDHXFGYAS9MP8V2Sc9nHcQW4PrMkQNnf2uGrDg81dFgBrwqv1n3frXxRBKhp83fSsTm4xqj8+jdwTI3nouFmi1W/O4UqpHdQ62EYoabJQtKpptWO11TFJzw8WF02pfS6git8YjLR4xrnfp2LkOEjSU9CI82ZasF46WZFKcpeUJTAsxU/03ONpAdwwEsC96f1KAvh8tqaO0yLDOcmPf8a5B82jefgncCRrt32kCpbpIE4YiCFrqqdUHXKH+",
	}
	dryLockedWallet = &mockedWallet{
		mnemonic:          []string{},
		password:          "pass",
		encryptedMnemonic: "dVoBFte1oeRkPl8Vf8DzBP3PRnzPA3fxtyvDHXFGYAS9MP8V2Sc9nHcQW4PrMkQNnf2uGrDg81dFgBrwqv1n3frXxRBKhp83fSsTm4xqj8+jdwTI3nouFmi1W/O4UqpHdQ62EYoabJQtKpptWO11TFJzw8WF02pfS6git8YjLR4xrnfp2LkOEjSU9CI82ZasF46WZFKcpeUJTAsxU/03ONpAdwwEsC96f1KAvh8tqaO0yLDOcmPf8a5B82jefgncCRrt32kCpbpIE4YiCFrqqdUHXKH+",
	}
	emptyWallet = &mockedWallet{
		mnemonic: []string{
			"curtain", "summer", "juice", "thought", "release", "velvet", "dress", "fantasy",
			"price", "hard", "core", "friend", "reopen", "myth", "giant", "consider",
			"seminar", "ladder", "thought", "spell", "state", "home", "diamond", "gold",
		},
		password:          "Sup3rS3cr3tP4ssw0rd!",
		encryptedMnemonic: "um8H1ulZShOz+zSZLUBjWVysVVqq8LGOneKte6fCSVHRDYsP6FG40W+NZ9IHCwSeigrGyr0rGazoNqIJy9Q9CaLMs2MA5yQVw1g19OuagZqXAsPrGY75FNKgcAYRRieSICC/ZnlwzPqZVxFGNIPza4bYe8JIflekPHKJ2y8kY4A6JThq4hWzVa7Icw7E4MautmpNYq9ic5ERcYL5lizamXYZ0u8KiRQr6bMW36d4jdgaIfizhbVxylBCtncriR4yOhSYB3Vi20YrzTorBwaDu1xcD5m552Bp6MKbcQ==",
	}
	usedWallet = &mockedWallet{
		mnemonic: []string{
			"trophy", "situate", "mobile", "royal", "disease", "obvious", "ramp", "buddy",
			"turn", "robust", "trust", "company", "wheel", "adult", "produce", "spawn",
			"afford", "inspire", "topic", "farm", "sword", "embark", "body", "runway",
		},
		password:          "Sup3rS3cr3tP4ssw0rd!",
		encryptedMnemonic: "46OIUILJEmvmdb/BbaTOEjMM743D5TnfqLBhl9c+E/PSG+7miMCpP3maRNttCP3RF/jdJnbzG6KkAbcKGXJROpF9tSGV5oizjp07lRG85fQH8OSJajn515sclXlKjX2aaB76b3Vt3a94pIzeZrQ2g5c8voupYnL0TDAjLd1Iltl5ApKLuPf5WfEJtvZ5Klb4rF+cLlvIjPtdqFHIwjotB8fR0LGr9yw1hfduDOWe+DPyNCkgbtKBKe0qWjBnnng88eMdlD8bsanuEkoiDlyHDnIvZ+JwgYOOUw==",
	}
)

var ctx = context.Background()

func newTestWallet(w *mockedWallet) (*walletService, func()) {
	dbManager, err := dbbadger.NewDbManager("testdb")
	if err != nil {
		panic(err)
	}
	vaultRepo := inmemory.NewVaultRepositoryImpl()
	if w != nil {
		vaultRepo = newMockedVaultRepositoryImpl(*w)
	}

	explorerSvc := explorer.NewService(config.GetString(config.ExplorerEndpointKey))
	return newWalletService(
			vaultRepo,
			dbbadger.NewUnspentRepositoryImpl(dbManager),
			crawler.NewService(explorerSvc, []crawler.Observable{}, func(err error) {}),
			explorerSvc,
		), func() {
			recover()
			dbManager.Store.Close()
			os.RemoveAll("testdb")
		}
}

func TestNewWalletService(t *testing.T) {
	ws, close := newTestWallet(nil)
	defer close()
	assert.Equal(t, false, ws.walletInitialized)
	assert.Equal(t, false, ws.walletIsSyncing)
}

func TestGenSeed(t *testing.T) {
	walletSvc, close := newTestWallet(nil)
	defer close()

	seed, err := walletSvc.GenSeed(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 24, len(seed))
}

func TestInitWalletWrongSeed(t *testing.T) {
	walletSvc, close := newTestWallet(nil)
	defer close()

	wrongSeed := []string{"test"}
	err := walletSvc.InitWallet(ctx, wrongSeed, "pass")
	assert.Error(t, err)
}

func TestInitEmptyWallet(t *testing.T) {
	walletSvc, close := newTestWallet(emptyWallet)
	defer close()
	// If the vault repository is not empty when the wallet service is
	// instantiated, this behaves like it  it was shut down and restarted again.
	// Therefore, the service restores its previous state and "marks" the wallet
	// as initialized by setting the internal walletInitialized bool field to
	// true. InitWallet, on its side, does not perform any operation if the
	// wallet looks already initialized.
	// In this test and in the next one, the walletInitialized field is manually
	// set to false because a mocked Vault repository is used that would cause
	// the bool field to be set to true when at service instantiation.
	walletSvc.walletInitialized = false

	w, _ := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: emptyWallet.mnemonic,
	})
	firstWalletAccountAddr, _, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
		DerivationPath: "1'/0/0",
		Network:        config.GetNetwork(),
	})

	err := walletSvc.InitWallet(ctx, emptyWallet.mnemonic, emptyWallet.password)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, walletSvc.walletInitialized)

	if err := walletSvc.UnlockWallet(ctx, emptyWallet.password); err != nil {
		t.Fatal(err)
	}
	addr, _, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, firstWalletAccountAddr, addr)
}

func TestInitUsedWallet(t *testing.T) {
	walletSvc, close := newTestWallet(usedWallet)
	defer close()
	walletSvc.walletInitialized = false

	w, _ := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: usedWallet.mnemonic,
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

	err := walletSvc.InitWallet(ctx, usedWallet.mnemonic, usedWallet.password)
	if err != nil {
		t.Fatal(err)
	}
	if err := walletSvc.UnlockWallet(ctx, usedWallet.password); err != nil {
		t.Fatal(err)
	}
	addr, _, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, firstWalletAccountAddr, addr)
}

func TestWalletUnlock(t *testing.T) {
	walletSvc, close := newTestWallet(dryLockedWallet)
	defer close()

	address, blindingKey, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	assert.Equal(t, domain.ErrMustBeUnlocked, err)

	err = walletSvc.UnlockWallet(ctx, dryLockedWallet.password)
	if err != nil {
		t.Fatal(err)
	}

	address, blindingKey, err = walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, len(address) > 0)
	assert.Equal(t, true, len(blindingKey) > 0)
}

func TestWalletChangePass(t *testing.T) {
	walletSvc, close := newTestWallet(dryLockedWallet)
	defer close()

	err := walletSvc.ChangePassword(ctx, "wrongPass", "newPass")
	assert.Equal(t, domain.ErrInvalidPassphrase, err)

	err = walletSvc.ChangePassword(ctx, dryLockedWallet.password, "newPass")
	assert.NoError(t, err)

	err = walletSvc.UnlockWallet(ctx, dryLockedWallet.password)
	assert.Equal(t, wallet.ErrInvalidPassphrase, err)
}

func TestWalletBalance(t *testing.T) {
	walletSvc, close := newTestWallet(dryWallet)
	defer close()

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
		true,
		int(balance[config.GetString(config.BaseAssetKey)].ConfirmedBalance) >= 100000000,
	)
}

// Mocked vault repository with harcoded mnemonic, passphrase and encrypted
// mnemonic. It can be initialized either locked or unlocked
type mockedVaultRepository struct {
	vault *domain.Vault
}

func newMockedVaultRepositoryImpl(w mockedWallet) domain.VaultRepository {
	return &mockedVaultRepository{
		&domain.Vault{
			Mnemonic:               w.mnemonic,
			EncryptedMnemonic:      w.encryptedMnemonic,
			PassphraseHash:         btcutil.Hash160([]byte(w.password)),
			Accounts:               map[int]*domain.Account{},
			AccountAndKeyByAddress: map[string]domain.AccountAndKey{},
		},
	}
}

func (r *mockedVaultRepository) GetOrCreateVault(ctx context.Context, mnemonic []string, passphrase string) (*domain.Vault, error) {
	return r.vault, nil
}

func (r *mockedVaultRepository) UpdateVault(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	updateFn func(v *domain.Vault) (*domain.Vault, error),
) error {
	v, err := updateFn(r.vault)
	if err != nil {
		return err
	}
	r.vault = v
	return nil
}

func (r *mockedVaultRepository) GetAccountByIndex(ctx context.Context, accountIndex int) (*domain.Account, error) {
	return r.vault.AccountByIndex(accountIndex)
}

func (r *mockedVaultRepository) GetAccountByAddress(ctx context.Context, addr string) (*domain.Account, int, error) {
	return r.vault.AccountByAddress(addr)
}

func (r *mockedVaultRepository) GetAllDerivedAddressesAndBlindingKeysForAccount(
	ctx context.Context,
	accountIndex int,
) ([]string, [][]byte, error) {
	return r.vault.AllDerivedAddressesAndBlindingKeysForAccount(accountIndex)
}

func (r *mockedVaultRepository) GetDerivationPathByScript(ctx context.Context, accountIndex int, scripts []string) (map[string]string, error) {
	a, _ := r.GetAccountByIndex(ctx, accountIndex)
	m := map[string]string{}
	for _, script := range scripts {
		m[script] = a.DerivationPathByScript[script]
	}
	return m, nil
}
