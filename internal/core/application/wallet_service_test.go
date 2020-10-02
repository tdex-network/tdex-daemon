package application

import (
	"context"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"testing"
	"time"
)

var ctx = context.Background()
var explorerSvc explorer.Service
var unspentRepository domain.UnspentRepository
var crawlerSvc crawler.Service
var vaultRepository domain.VaultRepository

func setUp() WalletService {
	unspentRepository = inmemory.NewUnspentRepositoryImpl()
	vaultRepository = inmemory.NewVaultRepositoryImpl()
	explorerSvc = explorer.NewService(config.GetString(config.ExplorerEndpointKey))
	walletSvc := NewWalletService(
		vaultRepository,
		unspentRepository,
		explorerSvc)

	return walletSvc
}

func startCrawler(t *testing.T) {
	marketRepository := inmemory.NewMarketRepositoryImpl()
	tradeRepository := inmemory.NewTradeRepositoryImpl()
	observables := make([]crawler.Observable, 0)

	errorHandler := func(err error) {
		t.Log(err)
	}
	crawlerSvc = crawler.NewService(explorerSvc, observables, errorHandler)
	operatorSvc := NewOperatorService(
		marketRepository,
		vaultRepository,
		tradeRepository,
		unspentRepository,
		explorerSvc,
		crawlerSvc,
	)
	operatorSvc.ObserveBlockchain()
}

func TestGenSeed(t *testing.T) {
	walletSvc := setUp()

	seed, err := walletSvc.GenSeed(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(seed), 24)
}

func TestInitWallet(t *testing.T) {
	walletSvc := setUp()
	seed, err := walletSvc.GenSeed(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = walletSvc.InitWallet(ctx, seed, "pass")
	assert.NoError(t, err)
}

func TestInitWalletWrongSeed(t *testing.T) {
	wrongSeed := []string{"test"}
	walletSvc := setUp()
	err := walletSvc.InitWallet(ctx, wrongSeed, "pass")
	assert.Error(t, err)
}

func TestWalletLockUnlock(t *testing.T) {
	walletSvc := setUp()
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
	walletSvc := setUp()
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
	walletSvc := setUp()
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

	address, blindingKey, err := walletSvc.GenerateAddressAndBlindingKey(ctx)
	if err != nil {
		t.Fatal(err)
	}

	bk, err := hex.DecodeString(blindingKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = explorerSvc.Faucet(address)
	if err != nil {
		t.Fatal(err)
	}

	startCrawler(t)

	crawlerSvc.AddObservable(&crawler.AddressObservable{
		AccountIndex: domain.WalletAccount,
		Address:      address,
		BlindingKey:  bk,
	})

	time.Sleep(10 * time.Second)

	balance, err := walletSvc.GetWalletBalance(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(
		t,
		uint64(100000000),
		balance[config.GetString(config.BaseAssetKey)].ConfirmedBalance,
	)
}
