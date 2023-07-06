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
	"github.com/tdex-network/tdex-daemon/examples"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/vulpemventures/go-elements/network"
)

var (
	composePath = "resources/compose/docker-compose.yml"
	volumesPath = "resources/volumes"
	// feederConfigJSON = fmt.Sprintf("%s/feederd/config.json", volumesPath)

	daemonAddr     = "localhost:9945"
	explorerAddr   = "http://localhost:3001"
	explorerSvc, _ = esplora.NewService(explorerAddr, 15000)

	password                   = "password"
	feeFragmenterDepositAmount = 0.001
	marketBaseDepositAmount    = 1.0
	marketQuoteDepositAmount   = float64(25000)
	// numOfConcurrentTrades      = 4

	lbtc = network.Regtest.AssetID
	usdt string
)

func main() {
	log.RegisterExitHandler(clear)

	if err := makeDirectoryIfNotExists(volumesPath); err != nil {
		log.WithError(err).Fatal("failed to create volume dir")
	}

	log.Info("starting oceand and tdexd services...")
	// docker-compose logs are sent to stderr therefore we cannot check for errors :(
	//nolint
	runCommand(
		"docker-compose", "-f", composePath, "up", "-d", "oceand", "tdexd",
	)
	log.Infof("done\n\n")

	log.Info("minting USDT asset...")
	if err := setupUSDTAsset(); err != nil {
		log.WithError(err).Fatal("failed to mint USDT asset")
	}
	log.Infof("asset: %s", usdt)
	log.Infof("done\n\n")

	time.Sleep(5 * time.Second)

	log.Info("configuring tdex CLI...")
	if _, err := runCLICommand("config", "init", "--no-tls", "--no-macaroons"); err != nil {
		log.WithError(err).Fatal("failed to config tdex CLI")
	}
	log.Infof("done\n\n")

	log.Info("initializing wallet...")
	seed, err := runCLICommand("genseed")
	if err != nil {
		log.WithError(err).Fatal("failed to retrieve mnemonic seed")
	}
	log.Infof("mnemonic: %s", seed)

	if _, err := runCLICommand("init", "--seed", seed, "--password", password); err != nil {
		log.WithError(err).Fatal("failed to initialize wallet")
	}
	log.Infof("done\n\n")

	// unlock with password
	log.Info("unlocking wallet...")
	if _, err := runCLICommand("unlock", "--password", password); err != nil {
		log.WithError(err).Fatal("failed to unlock wallet")
	}
	log.Infof("done\n\n")

	// deposit to the fee account via the fragmeneter one
	log.Infof("funding feefragmenter account with %f LBTC...\n", feeFragmenterDepositAmount)
	out, err := runCLICommand("feefragmenter", "deposit")
	if err != nil {
		log.WithError(err).Fatal("failed to derive addresses from feefragmenter account")
	}

	feeAddresses := addressesFromStdout(out)
	if err := fundFeeFragmenterAccount(feeAddresses); err != nil {
		log.WithError(err).Fatal("failed to fund feefragmenter account")
	}
	log.Infof("done\n\n")

	log.Info("splitting and depositing funds to fee account...")
	if _, err := runCLICommand("feefragmenter", "split"); err != nil {
		log.WithError(err).Fatal("failed to split and deposit feefragmnenter account funds to fee one")
	}
	log.Infof("done\n\n")

	// mint a block
	log.Info("minting a block...")
	if err := mintBlock(); err != nil {
		log.WithError(err).Fatal("failed to mint block")
	}
	log.Infof("done\n\n")

	// create a LBTC/USDT market
	log.Info("creating new market...")
	if _, err = runCLICommand(
		"market", "new", "--base-asset", lbtc, "--quote-asset", usdt,
		"--base-asset-precision", "8", "--quote-asset-precision", "8",
	); err != nil {
		log.WithError(err).Fatal("failed to create new market")
	}

	if _, err := runCLICommand("config", "set", "base_asset", lbtc); err != nil {
		log.WithError(err).Fatal("failed to configure market base asset")
	}
	if _, err := runCLICommand("config", "set", "quote_asset", usdt); err != nil {
		log.WithError(err).Fatal("failed to configure market quote asset")
	}
	log.Infof("done\n\n")

	// deposit funds to the market via the fragmenter one
	log.Infof("funding marketfragmenter account with %f LBTC and %f USDT...\n", marketBaseDepositAmount, marketQuoteDepositAmount)
	out, err = runCLICommand("marketfragmenter", "deposit")
	if err != nil {
		log.WithError(err).Fatal("failed to derive addresses from marketfragmenter account")
	}

	marketAddresses := addressesFromStdout(out)
	if err := fundMarketFragmenterAccount(marketAddresses); err != nil {
		log.WithError(err).Fatal("failed to fund marketfragmenter account")
	}
	log.Infof("done\n\n")

	log.Info("splitting and depositing funds to market account...")
	if _, err := runCLICommand("marketfragmenter", "split"); err != nil {
		log.WithError(err).Fatal("failed to split and deposit marketfragmenter account funds to market one")
	}
	log.Infof("done\n\n")

	// mint a block
	log.Info("minting a block...")
	if err := mintBlock(); err != nil {
		log.WithError(err).Fatal("failed to mint block")
	}
	log.Infof("done\n\n")

	// setup market fees
	log.Info("setting trading fees for the market...")
	if _, err := runCLICommand(
		"market", "percentagefee", "--base-fee", "100", "--quote-fee", "50",
	); err != nil {
		log.WithError(err).Fatal("failed to set market percentage fee")
	}
	if _, err := runCLICommand(
		"market", "fixedfee", "--base-fee", "500", "--quote-fee", "1000000",
	); err != nil {
		log.WithError(err).Fatal("failed to set market fixed fee")
	}
	log.Infof("done\n\n")

	// before opening the market, let's set its strategy to pluggable and also
	// start the feeder service.
	// log.Info("switching to pluggable market strategy...")
	// if _, err := runCLICommand("market", "strategy", "--pluggable"); err != nil {
	// 	log.WithError(err).Fatal("failed to update market strategy")
	// }
	// log.Infof("done\n\n")

	// TODO: restore using feeder once it supports tdex-daemon/v2 proto.
	// For now let's manually set a price for the market 1 LBTC = 20k USDT.

	// log.Info("starting feeder...")
	// if err := setupFeeder(); err != nil {
	// 	log.WithError(err).Fatal("failed to start feeder service")
	// }
	// time.Sleep(7 * time.Second)
	// log.Infof("done\n\n")

	// TODO: remove this step and restore usage of price feed
	log.Info("setting market price (TODO: remove this step)...")
	if _, err := runCLICommand(
		"market", "price", "--base-price", "0.00004", "--quote-price", "25000",
	); err != nil {
		log.WithError(err).Fatal("failed to update market price")
	}
	log.Infof("done\n\n")

	// open the market
	log.Info("opening market...")
	if _, err := runCLICommand("market", "open"); err != nil {
		log.WithError(err).Fatal("failed to open market")
	}
	log.Infof("done\n\n")

	time.Sleep(15 * time.Second)

	// trade on market
	log.Infof("trading on market:\n\n")
	if err = examples.SellExample(daemonAddr, explorerAddr); err != nil {
		log.WithError(err).Error("failed to sell lbtc for usd")
	}

	// mint a block
	log.Infof("minting a block...")
	if err := mintBlock(); err != nil {
		log.WithError(err).Fatal("failed to mint block")
	}
	log.Infof("done\n\n")

	if err = examples.BuyExample(daemonAddr, explorerAddr); err != nil {
		log.WithError(err).Fatal("failed to buy lbtc for usd")
	}

	// mint a block
	log.Info("minting a block...")
	if err := mintBlock(); err != nil {
		log.WithError(err).Fatal("failed to mint block")
	}
	log.Infof("done\n\n")

	// log.Info("setting up traders' wallets...")
	// client, _ := setupTraderClient()

	// wallets := make([]*trade.Wallet, 0, numOfConcurrentTrades)
	// assets := make([]string, 0, numOfConcurrentTrades)
	// for i := 0; i < numOfConcurrentTrades; i++ {
	// 	w, _ := trade.NewRandomWallet(&network.Regtest)
	// 	faucetAmount, asset := 0.0004, lbtc // 0.0004 LBTC
	// 	if i%2 != 0 {
	// 		faucetAmount, asset = 20, usdt // 20 USDT
	// 	}
	// 	if _, err := explorerSvc.Faucet(w.Address(), faucetAmount, asset); err != nil {
	// 		log.WithError(err).Fatal("failed to fund traders' wallets")
	// 	}

	// 	wallets = append(wallets, w)
	// 	assets = append(assets, asset)
	// }

	// time.Sleep(7 * time.Second)
	// log.Infof("done\n\n")

	// // start trading against the market
	// log.Info("start trading on market...")
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
	// 	log.WithError(err).Fatal("failed to trade on LBTC/USDT market")
	// }
	// for i := 0; i < numOfConcurrentTrades; i++ {
	// 	if err := tradeOnMarket(client, wallets[i], assets[i]); err != nil {
	// 		log.WithError(err).Fatal("failed to trade on LBTC/USDT market")
	// 	}
	// 	if err := mintBlock(); err != nil {
	// 		log.WithError(err).Fatal("failed to mint new block")
	// 	}
	// }
	log.Info("test succeeded")
	log.Exit(0)
}

func clear() {
	// stop all services
	//nolint
	runCommand("docker-compose", "-f", composePath, "down")
	// remove volumes
	//nolint
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
	//nolint
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
	time.Sleep(10 * time.Second)
	return nil
}

func fundMarketFragmenterAccount(addresses []string) error {
	for _, addr := range addresses {
		//nolint
		explorerSvc.Faucet(addr, marketBaseDepositAmount, "")
		//nolint
		explorerSvc.Faucet(addr, marketQuoteDepositAmount, usdt)
	}
	time.Sleep(10 * time.Second)
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

func getNodeAddress() (string, error) {
	client := esplora.NewHTTPClient(17 * time.Second)
	url := fmt.Sprintf("%s/getnewaddress", explorerAddr)
	status, resp, err := client.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf(resp)
	}

	res := map[string]string{}
	//nolint
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
