package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/tdex-network/tdex-daemon/pkg/trade"
	tradeclient "github.com/tdex-network/tdex-daemon/pkg/trade/client"
	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	"github.com/vulpemventures/go-elements/network"
	"golang.org/x/sync/errgroup"
)

var (
	daemonDatadir = "test/e2e/.tdex-daemon"
	daemonEnv     = []string{fmt.Sprintf("TDEX_DATA_DIR_PATH=%s", daemonDatadir)}
	daemon        = fmt.Sprintf("./build/tdexd-%s-%s", runtime.GOOS, runtime.GOARCH)

	cliDatadir = "test/e2e/.tdex-operator"
	cliEnv     = []string{fmt.Sprintf("TDEX_OPERATOR_DATADIR=%s", cliDatadir)}
	cli        = fmt.Sprintf("./build/tdex-%s-%s", runtime.GOOS, runtime.GOARCH)

	feeder           = fmt.Sprintf("./build/feederd-%s-%s", runtime.GOOS, runtime.GOARCH)
	feederConfigJSON = "test/e2e/config.json"

	explorerUrl    = "http://localhost:3001"
	explorerSvc, _ = esplora.NewService(explorerUrl, 15000)

	password                 = "password"
	feeDepositAmount         = 0.00005
	marketBaseDepositAmount  = 0.2
	marketQuoteDepositAmount = float64(10000)
	numOfConcurrentTrades    = 4

	lbtcAsset = network.Regtest.AssetID
	usdtAsset string
)

func main() {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("Recover from panic", rec)
		}
		clear()
	}()

	if err := setupUSDTAsset(); err != nil {
		fmt.Println(err)
		return
	}

	// build and start the daemon
	runCommand("make", "build")

	daemonEnv = append(daemonEnv, []string{
		"TDEX_NETWORK=regtest",
		"TDEX_EXPLORER_ENDPOINT=http://127.0.0.1:3001",
		"TDEX_LOG_LEVEL=5",
		"TDEX_BASE_ASSET=5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
		"TDEX_FEE_ACCOUNT_BALANCE_THRESHOLD=1000",
	}...)
	stopDaemon, err := runCommandDetached(os.Stdout, os.Stderr, daemon, daemonEnv)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stopDaemon()

	// build the CLI binary
	runCommand("make", "build-cli")

	// generate a new seed
	if _, err := runCommandWithEnv(cliEnv, cli, "config", "init", "--network", "regtest", "--explorer_url", explorerUrl); err != nil {
		fmt.Println(err)
		return
	}

	seed, err := runCommandWithEnv(cliEnv, cli, "genseed")
	if err != nil {
		fmt.Println(err)
		return
	}

	// init daemon with generated seed
	if _, err := runCommandWithEnv(cliEnv, cli, "init", "--seed", seed, "--password", password); err != nil {
		fmt.Println(err)
		return
	}

	// unlock with password
	if _, err := runCommandWithEnv(cliEnv, cli, "unlock", "--password", password); err != nil {
		fmt.Println(err)
		return
	}

	// deposit some funds to pay for future trades' network fees
	out, err := runCommandWithEnv(cliEnv, cli, "depositfee", "--num_of_addresses", "10")
	if err != nil {
		fmt.Println(err)
		return
	}
	feeAddresses := feeAddressesFromStdout(out)
	feeOutpoints := fundFeeAccount(feeAddresses)
	if _, err := runCommandWithEnv(cliEnv, cli, "claimfee", "--outpoints", feeOutpoints); err != nil {
		fmt.Println(err)
		return
	}

	// create a LBTC/USDT market on regtest and deposit funds
	out, err = runCommandWithEnv(cliEnv, cli, "depositmarket", "--num_of_addresses", "10")
	if err != nil {
		fmt.Println(err)
		return
	}
	marketAddresses := marketAddressesFromStdout(out)

	if _, err := runCommandWithEnv(cliEnv, cli, "config", "set", "base_asset", lbtcAsset); err != nil {
		fmt.Println(err)
		return
	}
	if _, err := runCommandWithEnv(cliEnv, cli, "config", "set", "quote_asset", usdtAsset); err != nil {
		fmt.Println(err)
		return
	}

	outpoints := fundMarketAccount(marketAddresses)
	if _, err := runCommandWithEnv(cliEnv, cli, "claimmarket", "--outpoints", outpoints); err != nil {
		fmt.Println(err)
		return
	}

	// before opening the market, let's set the strategy to pluggable)
	// and start the feeder
	if _, err := runCommandWithEnv(cliEnv, cli, "strategy", "--pluggable"); err != nil {
		fmt.Println(err)
		return
	}

	if err := setupFeeder(); err != nil {
		fmt.Println("feeder setup failed", err)
		return
	}
	feederEnv := []string{
		"FEEDER_LOG_LEVEL=5",
		"FEEDER_CONFIG_PATH=" + feederConfigJSON,
	}
	stopFeeder, err := runCommandDetached(nil, nil, feeder, feederEnv)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stopFeeder()

	time.Sleep(5 * time.Second)

	if _, err := runCommandWithEnv(cliEnv, cli, "open"); err != nil {
		fmt.Println(err)
		return
	}

	client, _ := setupTraderClient()

	wallets := make([]*trade.Wallet, 0, numOfConcurrentTrades)
	assets := make([]string, 0, numOfConcurrentTrades)
	for i := 0; i < numOfConcurrentTrades; i++ {
		w, _ := trade.NewRandomWallet(&network.Regtest)
		faucetAmount, asset := 0.0004, lbtcAsset // 0.0004 LBTC
		if i%2 != 0 {
			faucetAmount, asset = 20, usdtAsset // 20 USDT
		}
		if _, err := explorerSvc.Faucet(w.Address(), faucetAmount, asset); err != nil {
			fmt.Println(err)
			return
		}

		wallets = append(wallets, w)
		assets = append(assets, asset)
	}

	time.Sleep(10 * time.Second)

	// start trading against the market
	var eg errgroup.Group

	for i := 0; i < numOfConcurrentTrades; i++ {
		wallet := wallets[i]
		asset := assets[i]
		eg.Go(func() error {
			return tradeOnMarket(client, wallet, asset)
		})

		// TODO: our goal is to decrese this wating time. At the moment, decreasing
		// it would result in making trades that double spend some inputs. This is
		// caused by the daemon that apparently, selects the same unspents in case
		// of concurrent proposals, ie 2 requests arriving nearly at the same time.
		// Ideally, this should be removed.
		time.Sleep(2 * time.Second)
	}

	if err := eg.Wait(); err != nil {
		fmt.Printf("error occoured while trading against LBTC/USDT market: \n%s\n", err)
		return
	}

	return
}

func tradeOnMarket(client *trade.Trade, w *trade.Wallet, asset string) error {
	defer time.Sleep(200 * time.Millisecond)

	if asset == usdtAsset {
		_, err := client.BuyAndComplete(trade.BuyOrSellAndCompleteOpts{
			Market: trademarket.Market{
				BaseAsset:  lbtcAsset,
				QuoteAsset: usdtAsset,
			},
			Asset:       asset,
			Amount:      1000000000, // 10 USDT
			PrivateKey:  w.PrivateKey(),
			BlindingKey: w.BlindingKey(),
		})
		return err
	}

	_, err := client.SellAndComplete(trade.BuyOrSellAndCompleteOpts{
		Market: trademarket.Market{
			BaseAsset:  lbtcAsset,
			QuoteAsset: usdtAsset,
		},
		Asset:       asset,
		Amount:      20000, // 0.0002 LBTC
		PrivateKey:  w.PrivateKey(),
		BlindingKey: w.BlindingKey(),
	})
	return err
}

func clear() {
	// remove builds
	runCommand("rm", "-rf", "build")
	// remove datadirs
	runCommand("rm", "-rf", daemonDatadir)
	runCommand("rm", "-rf", cliDatadir)
	// remove feeder config file
	runCommand("rm", "-f", feederConfigJSON)
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

func runCommandWithEnv(env []string, name string, arg ...string) (string, error) {
	outb := new(strings.Builder)
	errb := new(strings.Builder)
	cmd := newCommand(outb, errb, name, arg...)
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		return "", err
	}
	if errMsg := errb.String(); len(errMsg) > 0 {
		return "", fmt.Errorf(errMsg)
	}

	return strings.Trim(outb.String(), "\n"), nil
}

func runCommandDetached(out, err io.Writer, name string, env []string) (func(), error) {
	cmd := newCommand(out, err, name)
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return func() {
		cmd.Process.Signal(syscall.SIGINT)
		cmd.Wait()
	}, nil
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

func feeAddressesFromStdout(out string) []string {
	res := make(map[string]interface{})
	json.Unmarshal([]byte(out), &res)

	addresses := make([]string, 0)
	for _, i := range res["address_with_blinding_key"].([]interface{}) {
		o := i.(map[string]interface{})
		addresses = append(addresses, o["address"].(string))
	}
	return addresses
}

func fundFeeAccount(addresses []string) string {
	// fund every address with 5000 sats
	for _, addr := range addresses {
		explorerSvc.Faucet(addr, feeDepositAmount, "")
	}
	time.Sleep(10 * time.Second)

	utxos, _ := explorerSvc.GetUnspentsForAddresses(addresses, nil)

	funds := make([]map[string]interface{}, 0, len(utxos))
	for _, u := range utxos {
		funds = append(funds, map[string]interface{}{
			"hash":  u.Hash(),
			"index": u.Index(),
		})
	}

	buf, _ := json.Marshal(funds)
	return string(buf)
}

func marketAddressesFromStdout(out string) []string {
	res := make(map[string]interface{})
	json.Unmarshal([]byte(out), &res)

	addresses := make([]string, 0)
	for _, i := range res["addresses"].([]interface{}) {
		addresses = append(addresses, i.(string))
	}
	return addresses
}

func fundMarketAccount(addresses []string) string {
	numAddr := len(addresses)

	for _, addr := range addresses[:numAddr/2] {
		explorerSvc.Faucet(addr, marketBaseDepositAmount, "")
	}

	for _, addr := range addresses[numAddr/2:] {
		explorerSvc.Faucet(addr, marketQuoteDepositAmount, usdtAsset)
	}

	time.Sleep(10 * time.Second)

	utxos, _ := explorerSvc.GetUnspentsForAddresses(addresses, nil)
	funds := make([]map[string]interface{}, 0, len(utxos))
	for _, u := range utxos {
		funds = append(funds, map[string]interface{}{
			"hash":  u.Hash(),
			"index": u.Index(),
		})
	}

	buf, _ := json.Marshal(funds)
	return string(buf)
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
	usdtAsset = asset
	return nil
}

func setupFeeder() error {
	configMap := map[string]interface{}{
		"daemon_endpoint":    "127.0.0.1:9000",
		"kraken_ws_endpoint": "ws.kraken.com",
		"markets": []map[string]interface{}{
			{
				"base_asset":    lbtcAsset,
				"quote_asset":   usdtAsset,
				"kraken_ticker": "XBT/USDT",
				"interval":      500,
			},
		},
	}
	configJSON, _ := json.Marshal(configMap)
	ioutil.WriteFile(feederConfigJSON, configJSON, 0777)

	// get feeder's latest relase

	cmd := "curl -s https://api.github.com/repos/Tdex-network/tdex-feeder/releases/latest | grep '\"tag_name\":' | sed -E 's/.*\"([^\"]+)\".*/\\1/'"
	latestVersion, err := runCommand("bash", "-c", cmd)
	if err != nil {
		return err
	}

	feederUrl := fmt.Sprintf(
		"https://github.com/TDex-network/tdex-feeder/releases/download/%s/feederd-%s-%s-%s",
		latestVersion, latestVersion, runtime.GOOS, runtime.GOARCH,
	)

	if _, err := runCommand("curl", "-sL", "-o", feeder, feederUrl); err != nil {
		return err
	}

	if _, err := runCommand("chmod", "+x", feeder); err != nil {
		return err
	}

	return nil
}

func setupTraderClient() (*trade.Trade, error) {
	client, _ := tradeclient.NewTradeClient("localhost", 9945)
	return trade.NewTrade(trade.NewTradeOpts{
		Chain:           "regtest",
		ExplorerService: explorerSvc,
		Client:          client,
	})
}

func getNodeAddress() (string, error) {
	client := esplora.NewHTTPClient(15 * time.Second)
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
