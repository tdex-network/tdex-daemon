package main

import (
	"errors"
	"fmt"
	"os"

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
		&openmarket,
		&closemarket,
		&updatestrategy,
	)

	err := app.Run(os.Args)
	if err != nil {
		fatal(err)
	}
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
