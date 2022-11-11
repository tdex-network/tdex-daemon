package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	tradeclient "github.com/tdex-network/tdex-daemon/pkg/trade/client"
	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	"github.com/vulpemventures/go-elements/network"
)

var (
	composePath = "resources/compose/docker-compose.yml"
	volumesPath = "resources/volumes"
	// feederConfigJSON = fmt.Sprintf("%s/feederd/config.json", volumesPath)

	explorerUrl    = "http://localhost:3001"
	explorerSvc, _ = esplora.NewService(explorerUrl, 15000)

	password                   = "password"
	feeFragmenterDepositAmount = 0.001
	marketBaseDepositAmount    = 0.5
	marketQuoteDepositAmount   = float64(10000)
	numOfConcurrentTrades      = 4

	lbtc = network.Regtest.AssetID
	usdt string
)

func main() {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("Recover from panic", rec)
		}
		clear()
	}()

	if err := makeDirectoryIfNotExists(volumesPath); err != nil {
		log.WithError(err).Error("failed to create volume dir")
		return
	}

	log.Info("starting ocean and tdex services...")
	// docker-compose logs are sent to stderr therefore we cannot check for errors :(
	runCommand(
		"docker-compose", "-f", composePath, "up", "-d", "oceand", "tdexd",
	)
	log.Infof("done\n\n")

	log.Info("minting USDT asset...")
	if err := setupUSDTAsset(); err != nil {
		log.WithError(err).Error("failed to mint USDT asset")
		return
	}
	log.Infof("asset: %s\n\n", usdt)

	log.Info("configuring tdex CLI...")
	if _, err := runCLICommand("config", "init", "--no_tls", "--no_macaroons"); err != nil {
		log.WithError(err).Error("failed to config tdex CLI")
		return
	}
	log.Infof("done\n\n")

	log.Info("asking for new mnemonic seed...")
	seed, err := runCLICommand("genseed")
	if err != nil {
		log.WithError(err).Error("failed to retrieve mnemonic seed")
		return
	}
	log.Infof("mnemonic: %s\n\n", seed)

	// init daemon with generated seed
	log.Info("initializing wallet...")
	if _, err := runCLICommand("init", "--seed", seed, "--password", password); err != nil {
		log.WithError(err).Error("failed to initialize wallet")
		return
	}
	log.Infof("done\n\n")

	// unlock with password
	log.Info("unlocking wallet...")
	if _, err := runCLICommand("unlock", "--password", password); err != nil {
		log.WithError(err).Error("failed to unlock wallet")
		return
	}
	log.Infof("done\n\n")

	// deposit some funds to the fee fragmenter account so they can easily be
	// split into several fragments of 5000 sats and deposited to the fee account
	log.Infof("funding feefragmenter account with %f LBTC...\n", feeFragmenterDepositAmount)
	out, err := runCLICommand("feefragmenter", "deposit")
	if err != nil {
		log.WithError(err).Error("failed to derive addresses from feefragmenter account")
		return
	}

	feeAddresses := addressesFromStdout(out)
	if err := fundFeeFragmenterAccount(feeAddresses); err != nil {
		log.WithError(err).Error("failed to fund feefragmenter account")
		return
	}
	log.Infof("done\n\n")

	log.Info("splitting and depositing funds to fee account...")
	if _, err := runCLICommand("feefragmenter", "split"); err != nil {
		log.WithError(err).Error("failed to split and deposit feefragmnenter account funds to fee one")
		return
	}

	if err := mintBlock(); err != nil {
		log.WithError(err).Error("failed to mint new block")
		return
	}
	log.Infof("done\n\n")

	// create a LBTC/USDT market on regtest and deposit funds
	log.Info("creating new market...")
	if _, err = runCLICommand(
		"market", "new", "--base_asset", lbtc, "--quote_asset", usdt,
	); err != nil {
		log.WithError(err).Error("failed to create new market")
		return
	}

	if _, err := runCLICommand("config", "set", "base_asset", lbtc); err != nil {
		log.WithError(err).Error("failed to configure market base asset")
		return
	}
	if _, err := runCLICommand("config", "set", "quote_asset", usdt); err != nil {
		log.WithError(err).Error("failed to configure market quote asset")
		return
	}
	log.Infof("done\n\n")

	log.Infof("funding marketfragmenter account with %f LBTC and %f USDT...\n", marketBaseDepositAmount, marketQuoteDepositAmount)
	out, err = runCLICommand("marketfragmenter", "deposit")
	if err != nil {
		log.WithError(err).Error("failed to derive addresses from marketfragmenter account")
		return
	}

	marketAddresses := addressesFromStdout(out)
	if err := fundMarketFragmenterAccount(marketAddresses); err != nil {
		log.WithError(err).Error("failed to fund marketfragmenter account")
		return
	}
	log.Infof("done\n\n")

	log.Info("splitting and depositing funds to market account...")
	if _, err := runCLICommand("marketfragmenter", "split"); err != nil {
		log.WithError(err).Error("failed to split and deposit marketfragmenter account funds to market one")
		return
	}

	if err := mintBlock(); err != nil {
		log.WithError(err).Error("failed to mint new block")
		return
	}
	log.Infof("done\n\n")

	// before opening the market, let's set its strategy to pluggable and also
	// start the feeder service.
	log.Info("switching to pluggable market strategy...")
	if _, err := runCLICommand("market", "strategy", "--pluggable"); err != nil {
		log.WithError(err).Error("failed to update market strategy")
		return
	}
	log.Infof("done\n\n")

	// TODO: restore using feeder once it supports tdex-daemon/v2 proto.
	// For now let's manually set a price for the market 1 LBTC = 20k USDT.

	// log.Info("starting feeder...")
	// if err := setupFeeder(); err != nil {
	// 	log.WithError(err).Error("failed to start feeder service")
	// 	return
	// }
	// time.Sleep(7 * time.Second)
	// log.Infof("done\n\n")
	if _, err := runCLICommand(
		"market", "price", "--base_price", "0.00005", "--quote_price", "20000",
	); err != nil {
		log.WithError(err).Error("failed to update market price")
		return
	}
	log.Infof("done\n\n")

	log.Info("opening market...")
	if _, err := runCLICommand("market", "open"); err != nil {
		log.WithError(err).Error("failed to open market")
		return
	}
	log.Infof("done\n\n")

	log.Info("setting up traders' wallets...")
	client, _ := setupTraderClient()

	wallets := make([]*trade.Wallet, 0, numOfConcurrentTrades)
	assets := make([]string, 0, numOfConcurrentTrades)
	for i := 0; i < numOfConcurrentTrades; i++ {
		w, _ := trade.NewRandomWallet(&network.Regtest)
		faucetAmount, asset := 0.0004, lbtc // 0.0004 LBTC
		if i%2 != 0 {
			faucetAmount, asset = 20, usdt // 20 USDT
		}
		if _, err := explorerSvc.Faucet(w.Address(), faucetAmount, asset); err != nil {
			log.WithError(err).Error("failed to fund traders' wallets")
			return
		}

		wallets = append(wallets, w)
		assets = append(assets, asset)
	}

	time.Sleep(7 * time.Second)
	log.Infof("done\n\n")

	// start trading against the market
	log.Info("start trading on market...")
	// TODO: restore concurrent trades when fixing issues in Ocean wallet on
	// getting non-wallet utxos. For now consecutive trades are made instead.

	// var eg errgroup.Group
	// for i := 0; i < numOfConcurrentTrades; i++ {
	// 	wallet := wallets[i]
	// 	asset := assets[i]
	// 	eg.Go(func() error {
	// 		if err := tradeOnMarket(client, wallet, asset); err != nil {
	// 			return err
	// 		}
	// 		return mintBlock()
	// 	})
	// }

	// if err := eg.Wait(); err != nil {
	// 	log.WithError(err).Error("failed to trade on LBTC/USDT market")
	// 	return
	// }
	for i := 0; i < numOfConcurrentTrades; i++ {
		if err := tradeOnMarket(client, wallets[i], assets[i]); err != nil {
			log.WithError(err).Error("failed to trade on LBTC/USDT market")
			return
		}
		if err := mintBlock(); err != nil {
			log.WithError(err).Error("failed to mint new block")
			return
		}
	}
	log.Info("done.\n\n")
	log.Info("test succeeded")
}

func tradeOnMarket(client *trade.Trade, w *trade.Wallet, asset string) error {
	if asset == usdt {
		txid, err := client.BuyAndComplete(trade.BuyOrSellAndCompleteOpts{
			Market: trademarket.Market{
				BaseAsset:  lbtc,
				QuoteAsset: usdt,
			},
			Asset:       asset,
			Amount:      1000000000, // 10.00 USDT
			PrivateKey:  w.PrivateKey(),
			BlindingKey: w.BlindingKey(),
		})
		if err != nil {
			return fmt.Errorf("failed to buy LBTC: %s", err)
		}
		log.Infof("bought LBTC for USDT: %s", txid)
		return nil
	}

	txid, err := client.SellAndComplete(trade.BuyOrSellAndCompleteOpts{
		Market: trademarket.Market{
			BaseAsset:  lbtc,
			QuoteAsset: usdt,
		},
		Asset:       asset,
		Amount:      20000, // 0.0002 LBTC
		PrivateKey:  w.PrivateKey(),
		BlindingKey: w.BlindingKey(),
	})
	if err != nil {
		return fmt.Errorf("failed to sell LBTC: %s", err)
	}
	log.Infof("sold LBTC for USDT: %s", txid)
	return nil
}

func clear() {
	// stop all services
	runCommand("docker-compose", "-f", composePath, "down")
	// remove volumes
	runCommand("rm", "-rf", volumesPath)
}

func runCLICommand(arg ...string) (string, error) {
	args := append([]string{"exec", "tdexd", "tdex"}, arg...)
	return runCommand("docker", args...)
}

func runCommand(name string, arg ...string) (string, error) {
	outb := new(strings.Builder)
	errb := new(strings.Builder)
	cmd := newCommand(outb, errb, name, arg...)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	if errMsg := errb.String(); len(errMsg) > 0 {
		return "", fmt.Errorf(errMsg)
	}

	return strings.Trim(outb.String(), "\n"), nil
}

func newCommand(out, err io.Writer, name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	if out != nil {
		cmd.Stdout = out
	}
	if err != nil {
		cmd.Stderr = err
	}
	return cmd
}

func addressesFromStdout(out string) []string {
	res := make(map[string]interface{})
	json.Unmarshal([]byte(out), &res)

	addresses := make([]string, 0)
	for _, i := range res["addresses"].([]interface{}) {
		addresses = append(addresses, i.(string))
	}
	return addresses
}

func fundFeeFragmenterAccount(addresses []string) error {
	// fund every address with 5000 sats
	for _, addr := range addresses {
		if _, err := explorerSvc.Faucet(addr, feeFragmenterDepositAmount, ""); err != nil {
			return err
		}
	}
	time.Sleep(7 * time.Second)
	return nil
}

func fundMarketFragmenterAccount(addresses []string) error {
	for _, addr := range addresses {
		explorerSvc.Faucet(addr, marketBaseDepositAmount, "")
		explorerSvc.Faucet(addr, marketQuoteDepositAmount, usdt)
	}
	time.Sleep(7 * time.Second)
	return nil
}

func setupUSDTAsset() error {
	// little trick to let nigiri fauceting an issued asset
	addr, err := getNodeAddress()
	if err != nil {
		return err
	}

	_, asset, err := explorerSvc.Mint(addr, 100000)
	if err != nil {
		return err
	}
	usdt = asset
	return nil
}

func mintBlock() error {
	// to mint a new block let's just faucet an address of the elements node's
	// wallet
	addr, err := getNodeAddress()
	if err != nil {
		return err
	}

	if _, err := explorerSvc.Faucet(addr, 1, ""); err != nil {
		return err
	}
	return nil
}

// func setupFeeder() error {
// 	if err := makeDirectoryIfNotExists(
// 		filepath.Dir(feederConfigJSON),
// 	); err != nil {
// 		return err
// 	}

// 	configMap := map[string]interface{}{
// 		"price_feeder": "kraken",
// 		"interval":     1000,
// 		"targets": map[string]string{
// 			"rpc_address":    "tdexd:9000",
// 			"tls_cert_path":  "",
// 			"macaroons_path": "",
// 		},
// 		"well_known_markets": map[string]interface{}{
// 			"kraken": []map[string]interface{}{
// 				{
// 					"base_asset":  lbtc,
// 					"quote_asset": usdt,
// 					"ticker":      "XBT/USDT",
// 				},
// 			},
// 		},
// 	}
// 	configJSON, _ := json.Marshal(configMap)
// 	ioutil.WriteFile(feederConfigJSON, configJSON, 0777)

// 	runCommand("docker-compose", "-f", composePath, "up", "-d", "feederd")
// 	return nil
// }

func setupTraderClient() (*trade.Trade, error) {
	client, _ := tradeclient.NewTradeClient("localhost", 9945)
	return trade.NewTrade(trade.NewTradeOpts{
		Chain:           "regtest",
		ExplorerService: explorerSvc,
		Client:          client,
	})
}

func getNodeAddress() (string, error) {
	client := esplora.NewHTTPClient(17 * time.Second)
	url := fmt.Sprintf("%s/getnewaddress", explorerUrl)
	status, resp, err := client.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf(resp)
	}

	res := map[string]string{}
	json.Unmarshal([]byte(resp), &res)
	return res["address"], nil
}

func makeDirectoryIfNotExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModeDir|0755); err != nil {
			return err
		}
	}
	return nil
}
