package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/vulpemventures/go-elements/payment"
	"google.golang.org/grpc"

	pkgtrade "github.com/tdex-network/tdex-daemon/pkg/trade"
	tradeclient "github.com/tdex-network/tdex-daemon/pkg/trade/client"
	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	pbwallet "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
)

const walletPassword = "Sup3rS3cr3tP4ssw0rd!"

func TestGrpcMain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	startDaemon()
	t.Cleanup(func() {
		// give the daemon the time to process last requests
		time.Sleep(1 * time.Second)
		stopDaemon()
	})

	time.Sleep(1 * time.Second)

	if err := initWallet(); err != nil {
		t.Fatal(err)
	}

	if err := initFee(); err != nil {
		t.Fatal(err)
	}

	if err := initMarketAccounts(); err != nil {
		t.Fatal(err)
	}

	clientWallet, err := newSingleKeyWallet()
	if err != nil {
		t.Fatal(err)
	}

	Parallelize(
		func() {
			for i := 0; i < 3; i++ {
				market, err := getTradableMarket()
				if err != nil {
					t.Fatal(err)
				}

				if _, err := tradeLBTCPerUSDTFixedLBTC(clientWallet, market); err != nil {
					t.Fatal(err)
				}
				time.Sleep(7 * time.Second)

				if _, err := tradeLBTCPerUSDTFixedUSDT(clientWallet, market); err != nil {
					t.Fatal(err)
				}
				time.Sleep(7 * time.Second)
			}
		},
		func() {
			for i := 0; i < 5; i++ {
				if err := initMarketAccounts(); err != nil {
					t.Fatal(err)
				}
			}
		},
	)
}

func startDaemon() {
	go main()
}

func stopDaemon() {
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	time.Sleep(2 * time.Second)
	os.RemoveAll(config.GetString(config.DataDirPathKey))
}

func initWallet() error {
	client, err := newWalletClient()
	if err != nil {
		return err
	}

	seedReply, err := client.GenSeed(context.Background(), &pbwallet.GenSeedRequest{})
	if err != nil {
		return err
	}

	if _, err := client.InitWallet(context.Background(), &pbwallet.InitWalletRequest{
		WalletPassword: []byte(walletPassword),
		SeedMnemonic:   seedReply.GetSeedMnemonic(),
	}); err != nil {
		return err
	}

	time.Sleep(20 * time.Second)

	if _, err := client.UnlockWallet(context.Background(), &pbwallet.UnlockWalletRequest{
		WalletPassword: []byte(walletPassword),
	}); err != nil {
		return err
	}

	return nil
}

func initFee() error {
	client, err := newOperatorClient()
	if err != nil {
		return err
	}
	explorerSvc := explorer.NewService(config.GetString(config.ExplorerEndpointKey))
	ctx := context.Background()

	// get an address for funding the fee account
	depositFeeReply, err := client.DepositFeeAccount(ctx, &pboperator.DepositFeeAccountRequest{})
	if err != nil {
		return err
	}
	if _, err := explorerSvc.Faucet(
		depositFeeReply.GetAddressWithBlindingKey()[0].GetAddress(),
	); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	return nil
}

func initMarketAccounts() error {
	client, err := newOperatorClient()
	if err != nil {
		return err
	}
	explorerSvc := explorer.NewService(config.GetString(config.ExplorerEndpointKey))
	ctx := context.Background()

	// create a new market
	depositMarketReply, err := client.DepositMarket(ctx, &pboperator.DepositMarketRequest{})
	if err != nil {
		return err
	}

	addr := depositMarketReply.GetAddresses()[0]

	// and fund it with 1 LBTC and 6500 USDT
	if _, err := explorerSvc.Faucet(addr); err != nil {
		return err
	}
	_, usdt, err := explorerSvc.Mint(addr, 6500)
	if err != nil {
		return err
	}
	lbtc := config.GetNetwork().AssetID

	time.Sleep(8 * time.Second)

	// ...finally, open the market
	if _, err := client.OpenMarket(ctx, &pboperator.OpenMarketRequest{
		Market: &pbtypes.Market{
			BaseAsset:  lbtc,
			QuoteAsset: usdt,
		},
	}); err != nil {
		return err
	}

	return nil
}

func getTradableMarket() (m trademarket.Market, err error) {
	client, err := tradeclient.NewTradeClient("localhost", config.GetInt(config.TraderListeningPortKey))
	if err != nil {
		return
	}

	// get trading market from the list of all those tradable
	marketsReply, err := client.Markets()
	if err != nil {
		return
	}
	if len(marketsReply.GetMarkets()) <= 0 {
		err = errors.New("no open markets found")
		return
	}

	m = trademarket.Market{
		BaseAsset:  marketsReply.GetMarkets()[0].GetMarket().GetBaseAsset(),
		QuoteAsset: marketsReply.GetMarkets()[0].GetMarket().GetQuoteAsset(),
	}

	return
}

func tradeLBTCPerUSDTFixedLBTC(w wallet, market trademarket.Market) (string, error) {
	balances, err := w.getWalletBalance()
	if err != nil {
		return "", err
	}

	if balances[config.GetNetwork().AssetID] == 0 {
		if _, err := w.explorer.Faucet(w.address); err != nil {
			return "", err
		}
		time.Sleep(5 * time.Second)
	}

	client, err := tradeclient.NewTradeClient("localhost", config.GetInt(config.TraderListeningPortKey))
	if err != nil {
		return "", err
	}
	tr, err := pkgtrade.NewTrade(trade.NewTradeOpts{
		Chain:       "regtest",
		ExplorerURL: config.GetString(config.ExplorerEndpointKey),
		Client:      client,
	})
	if err != nil {
		return "", err
	}

	// SELL LBTC specifing the will of sending 0.1 LBTC
	return tr.SellAndComplete(pkgtrade.BuyOrSellAndCompleteOpts{
		Market:      market,
		TradeType:   int(tradetype.Sell),
		Amount:      10000000,
		Asset:       market.BaseAsset,
		PrivateKey:  w.signingKey,
		BlindingKey: w.blindingKey,
	})
}

func tradeLBTCPerUSDTFixedUSDT(w wallet, market trademarket.Market) (string, error) {
	client, err := tradeclient.NewTradeClient("localhost", config.GetInt(config.TraderListeningPortKey))
	if err != nil {
		return "", err
	}
	tr, err := pkgtrade.NewTrade(trade.NewTradeOpts{
		Chain:       "regtest",
		ExplorerURL: config.GetString(config.ExplorerEndpointKey),
		Client:      client,
	})
	if err != nil {
		return "", err
	}

	// SELL LBTC specifing the will of buying 400 USDT
	return tr.SellAndComplete(pkgtrade.BuyOrSellAndCompleteOpts{
		Market:      market,
		TradeType:   int(tradetype.Sell),
		Amount:      40000000000,
		Asset:       market.QuoteAsset,
		PrivateKey:  w.signingKey,
		BlindingKey: w.blindingKey,
	})
}

func newWalletClient() (pbwallet.WalletClient, error) {
	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "localhost", config.GetInt(config.OperatorListeningPortKey)), opts...)
	if err != nil {
		return nil, err
	}
	return pbwallet.NewWalletClient(conn), nil
}

func newOperatorClient() (pboperator.OperatorClient, error) {
	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "localhost", config.GetInt(config.OperatorListeningPortKey)), opts...)
	if err != nil {
		return nil, err
	}
	return pboperator.NewOperatorClient(conn), nil
}

type wallet struct {
	signingKey  []byte
	blindingKey []byte
	address     string
	explorer    explorer.Service
}

func newSingleKeyWallet() (w wallet, err error) {
	prvkey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return
	}
	blindkey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return
	}

	p2wpkh := payment.FromPublicKey(prvkey.PubKey(), config.GetNetwork(), blindkey.PubKey())
	ctAddress, err := p2wpkh.ConfidentialWitnessPubKeyHash()
	if err != nil {
		return
	}

	explorerSvc := explorer.NewService(config.GetString(config.ExplorerEndpointKey))

	w = wallet{
		signingKey:  prvkey.Serialize(),
		blindingKey: blindkey.Serialize(),
		address:     ctAddress,
		explorer:    explorerSvc,
	}
	return
}

func (w wallet) getWalletBalance() (map[string]uint64, error) {
	unspents, err := w.explorer.GetUnspents(w.address, [][]byte{w.blindingKey})
	if err != nil {
		return nil, err
	}
	balances := map[string]uint64{}
	for _, u := range unspents {
		balances[u.Asset()] += u.Value()
	}
	return balances, nil
}

// Parallelize parallelizes the function calls
func Parallelize(functions ...func()) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(functions))

	defer waitGroup.Wait()

	for _, function := range functions {
		go func(copy func()) {
			defer waitGroup.Done()
			copy()
		}(function)
	}
}
