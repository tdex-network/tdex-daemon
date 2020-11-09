package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/btcsuite/btcutil"
	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/go-elements/network"
	"google.golang.org/grpc"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbwallet "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
)

var (
	networkFlag = cli.StringFlag{
		Name:  "network, n",
		Usage: "the network tdexd is running on: liquid or regtest",
		Value: network.Regtest.Name,
	}

	rpcFlag = cli.StringFlag{
		Name:  "rpcserver",
		Usage: "tdexd daemon address host:port",
		Value: "localhost:9000",
	}

	macaroonFlag = cli.StringFlag{
		Name:  "macaroon",
		Usage: "hex encoded admin macaroon",
		Value: "",
	}

	// maxMsgRecvSize is the largest message our client will receive. We
	// set this to 200MiB atm.
	maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 200)

	tdexDataDir = btcutil.AppDataDir("tdex-operator", false)
	statePath   = path.Join(tdexDataDir, "state.json")
)

func main() {
	app := cli.NewApp()

	app.Version = "0.0.1" //TODO use goreleaser for setting version
	app.Name = "tdex operator CLI"
	app.Usage = "Command line interface for tdexd daemon operators"
	app.Flags = []cli.Flag{
		&networkFlag,
		&rpcFlag,
		&macaroonFlag,
	}
	app.Commands = append(
		app.Commands,
		&initwallet,
		&unlockwallet,
		&depositfee,
		&depositmarket,
		&market,
		&openmarket,
		&closemarket,
		&updatestrategy,
	)

	err := app.Run(os.Args)
	if err != nil {
		fatal(err)
	}

	if _, err := os.Stat(tdexDataDir); os.IsNotExist(err) {
		os.Mkdir(tdexDataDir, os.ModeDir|0755)
	}
}

func getMarketFromState() (string, string, error) {
	state, err := getState()
	if err != nil {
		return "", "", errors.New("a market must be selected")
	}
	baseAsset := state["base_asset"].(string)
	quoteAsset := state["quote_asset"].(string)

	return baseAsset, quoteAsset, nil
}

func setMarketIntoState(baseAsset, quoteAsset string) error {
	return setState(map[string]string{
		"base_asset":  baseAsset,
		"quote_asset": quoteAsset,
	})
}

func getState() (map[string]interface{}, error) {
	data := map[string]interface{}{}

	file, err := ioutil.ReadFile(statePath)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(file, &data)

	return data, nil
}

func setState(data map[string]string) error {
	jsonString, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(statePath, jsonString, 0755)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}

func getOperatorClient(ctx *cli.Context) (pboperator.OperatorClient, func(),
	error) {

	rpcServer := ctx.String("rpcserver")

	var macaroonHex = ""

	conn, err := getClientConn(rpcServer, macaroonHex)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }

	return pboperator.NewOperatorClient(conn), cleanup, nil
}

func getWalletClient(ctx *cli.Context) (pbwallet.WalletClient, func(),
	error) {

	rpcServer := ctx.String("rpcserver")

	var macaroonHex = ""

	conn, err := getClientConn(rpcServer, macaroonHex)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }

	return pbwallet.NewWalletClient(conn), cleanup, nil
}

func getClientConn(address, macaroonHex string) (*grpc.ClientConn,
	error) {

	opts := []grpc.DialOption{grpc.WithDefaultCallOptions(maxMsgRecvSize), grpc.WithInsecure()}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to RPC server: %v",
			err)
	}

	return conn, nil
}

type invalidUsageError struct {
	ctx     *cli.Context
	command string
}

func (e *invalidUsageError) Error() string {
	return fmt.Sprintf("invalid usage of command %s", e.command)
}

func fatal(err error) {
	var e *invalidUsageError
	if errors.As(err, &e) {
		_ = cli.ShowCommandHelp(e.ctx, e.command)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "[tdex] %v\n", err)
	}
	os.Exit(1)
}
