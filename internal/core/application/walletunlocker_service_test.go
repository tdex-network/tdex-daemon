package application_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/network"
)

var (
	regtest                    = &network.Regtest
	marketBaseAsset            = regtest.AssetID
	marketFee           int64  = 25
	feeBalanceThreshold uint64 = 5000
	restore                    = true
	passphrase                 = "passphrase"
	mnemonic                   = []string{
		"curious", "alien", "peanut", "protect", "capable", "charge", "recipe", "hub",
		"volume", "deal", "math", "make", "suggest", "bleak", "seat", "swim",
		"into", "save", "hint", "wood", "pioneer", "ball", "decline", "universe",
	}
	mnemonicStr       = strings.Join(mnemonic, " ")
	encryptedMnemonic = "RF0mJIJzHqokOgSD8fbHSy9YpN56qTZAaMxRQD6DSH27Q1Y7npNy4wuznaBbJL3s6j7HkmBOEGLpj8gf9PHo4cbv+6uIStF8DE0wTNxSu8AJKPQDbYi/lx59mhIkisL77Zx2cZQKFrFvTGHw5En8Zt8eKgFSnrM1goZZbsU9oe5C6MRK8zLdmVau9ipTN3nhTFMfTR1KsQ5OLhXWpjIdezrdb1LmN/7I/CU3Ts81/+R5fefzaa4vB+3g02TgPJmcvr1Yg53gjfwBpUVtrK4naQ=="
	ctx               = context.Background()
)

func TestMain(m *testing.M) {
	domain.EncrypterManager = mockCryptoHandler{
		encrypt: func(_, _ string) (string, error) {
			return encryptedMnemonic, nil
		},
		decrypt: func(_, _ string) (string, error) {
			return mnemonicStr, nil
		},
	}
	domain.MnemonicStoreManager = newSimpleMnemonicStore(nil)

	mockedPsetParser := &mockPsetParser{}
	mockedPsetParser.On("GetTxID", mock.AnythingOfType("string")).Return(randomHex(32), nil)
	mockedPsetParser.On("GetTxHex", mock.AnythingOfType("string")).Return(randomHex(1000), nil)
	domain.PsetParserManager = mockedPsetParser

	mockedSwapRequest := newMockedSwapRequest()
	mockedSwapAccept := newMockedSwapAccept()
	mockedSwapComplete := newMockedSwapComplete()
	mockedSwapFail := newMockedSwapFail()

	mockedSwapParser := &mockSwapParser{}
	mockedSwapParser.
		On("SerializeRequest", mock.Anything).Return(randomBytes(100), nil)
	mockedSwapParser.
		On("SerializeAccept", mock.Anything).Return(mockedSwapAccept.GetId(), randomBytes(100), nil)
	mockedSwapParser.
		On("SerializeComplete", mock.Anything, mock.Anything).Return(mockedSwapComplete.GetId(), randomBytes(100), nil)
	mockedSwapParser.
		On("SerializeFail", mock.Anything, mock.Anything, mock.Anything).Return(mockedSwapFail.GetId(), nil)
	mockedSwapParser.
		On("DeserializeRequest", mock.Anything).Return(mockedSwapRequest, nil)
	mockedSwapParser.
		On("DeserializeAccept", mock.Anything).Return(mockedSwapAccept, nil)
	mockedSwapParser.
		On("DeserializeComplete", mock.Anything).Return(mockedSwapComplete, nil)
	mockedSwapParser.
		On("DeserializeFail", mock.Anything).Return(mockedSwapFail, nil)
	domain.SwapParserManager = mockedSwapParser

	mockedTransactionManager := &mockTransactionManager{}
	mockedTransactionManager.
		On("ExtractUnspents", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil, nil)
	application.TransactionManager = mockedTransactionManager

	os.Exit(m.Run())
}

func TestInitWallet(t *testing.T) {
	t.Run("wallet_from_scratch", func(t *testing.T) {
		walletSvc := newWalletUnlockerService()
		require.NotNil(t, walletSvc)

		chReplies := make(chan *application.InitWalletReply)
		chErr := make(chan error, 1)

		go walletSvc.InitWallet(
			ctx,
			mnemonic,
			passphrase,
			!restore,
			chReplies,
			chErr,
		)

		replies, err := listenToReplies(chReplies, chErr)
		require.NoError(t, err)
		require.Len(t, replies, 0)
	})

	t.Run("wallet_from_restart", func(t *testing.T) {
		walletSvc, err := newWalletUnlockerServiceRestart()
		require.NoError(t, err)
		require.NotNil(t, walletSvc)

		// No need to call InitWallet when restarting a wallet service!
		// This is the only case where we can call directly Unlock because the
		// service doesn't make any async ops.
		err = walletSvc.UnlockWallet(ctx, passphrase)
		require.NoError(t, err)
	})

	t.Run("wallet_from_restore", func(t *testing.T) {
		walletSvc := newWalletUnlockerServiceRestore()
		require.NotNil(t, walletSvc)

		chReplies := make(chan *application.InitWalletReply)
		chErr := make(chan error, 1)

		go walletSvc.InitWallet(
			ctx,
			mnemonic,
			passphrase,
			restore,
			chReplies,
			chErr,
		)

		replies, err := listenToReplies(chReplies, chErr)
		require.NoError(t, err)
		require.Greater(t, len(replies), 0)
	})
}

func newWalletUnlockerService() application.WalletUnlockerService {
	repoManager, explorerSvc, bcListener := newServices()

	return application.NewWalletUnlockerService(
		repoManager,
		explorerSvc,
		bcListener,
		regtest,
		marketFee,
		marketBaseAsset,
	)
}

// When the wallet service is instatiated, it automatically takes care of
// restoring the utxo set if it finds a vault in the Vault repository.
// This function creates a new vault in the repo before passing the db manager
// down to the wallet service to simulate the described situation.
func newWalletUnlockerServiceRestart() (application.WalletUnlockerService, error) {
	repoManager, explorerSvc, bcListener := newServices()
	v, err := repoManager.VaultRepository().GetOrCreateVault(
		ctx, mnemonic, passphrase, regtest,
	)
	if err != nil {
		return nil, err
	}

	info := v.AllDerivedAddressesInfo()
	addresses, keys := info.AddressesAndKeys()
	explorerSvc.(*mockExplorer).
		On("GetUnspentsForAddresses", addresses, keys).
		Return(randomUtxos(addresses), nil)

	return application.NewWalletUnlockerService(
		repoManager,
		explorerSvc,
		bcListener,
		regtest,
		marketFee,
		marketBaseAsset,
	), nil
}

// Restoring a wallet is an operation that depends almost entirely on the
// explorer service. This function mocks explorer's responses in order to
// emulate an already used wallet with some used Fee account's addresses.
func newWalletUnlockerServiceRestore() application.WalletUnlockerService {
	repoManager, explorerSvc, bcListener := newServices()

	v, _ := domain.NewVault(mnemonic, passphrase, regtest)
	accountIndexes := []int{
		domain.FeeAccount,
		domain.WalletAccount,
		domain.MarketAccountStart,
	}
	usedAddresses := make([]string, 0)
	usedKeys := make([][]byte, 0)
	unusedAddresses := make([]string, 0)
	unusedKeys := make([][]byte, 0)
	for i := range accountIndexes {
		accountIndex := accountIndexes[i]
		for j := 0; j < 22; j++ {
			v.Unlock(passphrase)
			extInfo, _ := v.DeriveNextExternalAddressForAccount(accountIndex)
			v.Unlock(passphrase)
			inInfo, _ := v.DeriveNextInternalAddressForAccount(accountIndex)
			if j < 2 && accountIndex < domain.WalletAccount {
				usedAddresses = append(usedAddresses, extInfo.Address)
				usedAddresses = append(usedAddresses, inInfo.Address)
				usedKeys = append(usedKeys, extInfo.BlindingKey)
				usedKeys = append(usedKeys, inInfo.BlindingKey)
			} else {
				unusedAddresses = append(unusedAddresses, extInfo.Address)
				unusedAddresses = append(unusedAddresses, inInfo.Address)
				unusedKeys = append(unusedKeys, extInfo.BlindingKey)
				unusedKeys = append(unusedKeys, inInfo.BlindingKey)
			}
		}
	}

	for i, addr := range usedAddresses {
		key := usedKeys[i]
		explorerSvc.(*mockExplorer).
			On("GetTransactionsForAddress", addr, key).
			Return(randomTxs(addr), nil)
		explorerSvc.(*mockExplorer).
			On("GetUnspents", addr, [][]byte{key}).
			Return(randomUtxos([]string{addr}), nil)
	}

	for i, addr := range unusedAddresses {
		key := unusedKeys[i]
		explorerSvc.(*mockExplorer).
			On("GetTransactionsForAddress", addr, key).
			Return(nil, nil)
	}

	return application.NewWalletUnlockerService(
		repoManager,
		explorerSvc,
		bcListener,
		regtest,
		marketFee,
		marketBaseAsset,
	)
}

func newServices() (
	ports.RepoManager,
	explorer.Service,
	application.BlockchainListener,
) {
	repoManager, _ := dbbadger.NewRepoManager("", nil)
	explorerSvc := &mockExplorer{}
	crawlerSvc := crawler.NewService(crawler.Opts{
		ExplorerSvc:        explorerSvc,
		ExplorerLimit:      10,
		ExplorerTokenBurst: 1,
		CrawlerInterval:    1000,
	})
	bcListener := application.NewBlockchainListener(
		crawlerSvc,
		repoManager,
		nil,
		marketBaseAsset,
		regtest,
	)
	return repoManager, explorerSvc, bcListener
}

func listenToReplies(
	chReplies chan *application.InitWalletReply,
	chErr chan error,
) ([]*application.InitWalletReply, error) {
	replies := make([]*application.InitWalletReply, 0)
	for {
		select {
		case err, ok := <-chErr:
			if ok {
				return nil, err
			}
		case reply, ok := <-chReplies:
			if !ok {
				return replies, nil
			}
			replies = append(replies, reply)
		}
	}
}

func randomUtxos(addresses []string) []explorer.Utxo {
	uLen := len(addresses)
	utxos := make([]explorer.Utxo, uLen, uLen)
	for i, addr := range addresses {
		script, _ := address.ToOutputScript(addr)
		utxos[i] = esplora.NewWitnessUtxo(
			randomHex(32),           //hash
			randomVout(),            // index
			randomValue(),           // value
			randomHex(32),           // asset
			randomValueCommitment(), // valuecommitment
			randomAssetCommitment(), // assetcommitment
			randomBytes(32),         // valueblinder
			randomBytes(32),         // assetblinder
			script,
			randomBytes(32),  // nonce
			randomBytes(100), // rangeproof
			randomBytes(100), // surjectionproof
			true,             // confirmed
		)
	}
	return utxos
}

func randomTxs(addr string) []explorer.Transaction {
	return []explorer.Transaction{&mockTransaction{addr}}
}

func randomValueCommitment() string {
	c := randomBytes(32)
	c[0] = 9
	return hex.EncodeToString(c)
}

func randomAssetCommitment() string {
	c := randomBytes(32)
	c[0] = 10
	return hex.EncodeToString(c)
}

func randomId() string {
	return uuid.New().String()
}

func randomHex(len int) string {
	return hex.EncodeToString(randomBytes(len))
}

func randomVout() uint32 {
	return uint32(randomIntInRange(0, 15))
}

func randomValue() uint64 {
	return uint64(randomIntInRange(1000000, 10000000000))
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	rand.Read(b)
	return b
}

func randomIntInRange(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(int(n.Int64())) + min
}
