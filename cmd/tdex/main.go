package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/btcsuite/btcutil"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbwallet "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
)

var (
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
	app.Commands = append(
		app.Commands,
		&config,
		&genseed,
		&initwallet,
		&unlockwallet,
		&depositfee,
		&depositmarket,
		&listmarket,
		&listswaps,
		&openmarket,
		&closemarket,
		&updatestrategy,
		&updateprice,
	)

	err := app.Run(os.Args)
	if err != nil {
		fatal(err)
	}
}

func getState() (map[string]string, error) {
	data := map[string]string{}

	file, err := ioutil.ReadFile(statePath)
	if err != nil {
		return nil, errors.New("get config state error: try 'config init'")
	}
	json.Unmarshal(file, &data)

	return data, nil
}

func setState(data map[string]string) error {

	if _, err := os.Stat(tdexDataDir); os.IsNotExist(err) {
		os.Mkdir(tdexDataDir, os.ModeDir|0755)
	}

	file, err := os.OpenFile(statePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	currentData, err := getState()
	if err != nil {
		fmt.Println(err)
		return err
	}

	mergedData := merge(currentData, data)

	jsonString, err := json.Marshal(mergedData)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(statePath, jsonString, 0755)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}

func merge(maps ...map[string]string) map[string]string {
	merge := make(map[string]string, 0)
	for _, m := range maps {
		for k, v := range m {
			merge[k] = v
		}
	}
	return merge
}

/*
Modified from https://github.com/lightninglabs/pool/blob/master/cmd/pool/main.go
Original Copyright 2017 Oliver Gugger. All Rights Reserved.
*/
func printRespJSON(resp interface{}) {
	jsonMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		OrigName:     true,
		Indent:       "\t", // Matches indentation of printJSON.
	}

	jsonStr, err := jsonMarshaler.MarshalToString(resp.(proto.Message))
	if err != nil {
		fmt.Println("unable to decode response: ", err)
		return
	}

	fmt.Println(jsonStr)
}

func getOperatorClient(ctx *cli.Context) (pboperator.OperatorClient, func(),
	error) {

	conn, err := getClientConn()
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }

	return pboperator.NewOperatorClient(conn), cleanup, nil
}

func getWalletClient(ctx *cli.Context) (pbwallet.WalletClient, func(),
	error) {

	conn, err := getClientConn()
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }

	return pbwallet.NewWalletClient(conn), cleanup, nil
}

func getClientConn() (*grpc.ClientConn,
	error) {

	state, err := getState()
	if err != nil {
		return nil, err
	}
	address, ok := state["rpcserver"]
	if !ok {
		return nil, errors.New("set rpcserver with `config set rpcserver`")
	}
	//macaroon := state["macaroon"]

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
