package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/credentials"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"gopkg.in/macaroon.v2"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	// maxMsgRecvSize is the largest message our client will receive. We
	// set this to 200MiB atm.
	maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 200)

	tdexDataDir = btcutil.AppDataDir("tdex-operator", false)
	statePath   = filepath.Join(tdexDataDir, "state.json")

	initialState = map[string]string{
		"network":        defaultNetwork,
		"rpcserver":      defaultRPCServer,
		"no_macaroons":   strconv.FormatBool(defaultNoMacaroonsAuth),
		"tls_cert_path":  defaultTLSCertPath,
		"macaroons_path": defaultMacaroonsPath,
	}
)

func init() {
	dataDir := cleanAndExpandPath(os.Getenv("TDEX_OPERATOR_DATADIR"))
	if len(dataDir) <= 0 {
		return
	}

	tdexDataDir = dataDir
	statePath = filepath.Join(tdexDataDir, "state.json")
}

func main() {
	app := cli.NewApp()

	app.Version = formatVersion()
	app.Name = "tdex operator CLI"
	app.Usage = "Command line interface for tdexd daemon operators"
	app.Commands = append(
		app.Commands,
		&cliConfig,
		&genseed,
		&initwallet,
		&unlockwallet,
		&lockwallet,
		&status,
		&changepassword,
		&getwalletinfo,
		&walletAccount,
		&feeAccount,
		&feeFragmenterAccount,
		&marketAccount,
		&marketFragmenterAccount,
		&listmarkets,
		&listtrades,
		&listutxos,
		&reloadtxos,
		&addwebhook,
		&removewebhook,
		&listwebhooks,
		&listdeposits,
		&listwithdrawals,
		&contentType,
	)

	app.Before = func(ctx *cli.Context) error {
		if _, err := os.Stat(tdexDataDir); os.IsNotExist(err) {
			os.Mkdir(tdexDataDir, os.ModeDir|0755)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fatal(err)
	}
}

func getState() (map[string]string, error) {
	file, err := ioutil.ReadFile(statePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := setState(initialState); err != nil {
			return nil, err
		}
		return initialState, nil
	}

	data := map[string]string{}
	json.Unmarshal(file, &data)

	return data, nil
}

func setState(data map[string]string) error {
	file, err := os.OpenFile(statePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	currentData, err := getState()
	if err != nil {
		return err
	}

	noMacaroons, ok := data[noMacaroonsKey]
	if ok {
		noMac, err := strconv.ParseBool(noMacaroons)
		if err != nil {
			return fmt.Errorf("invalid bool value for %s: %s", noMacaroonsKey, err)
		}
		if noMac {
			data[macaroonsPathKey] = ""
		}
	}
	noTls, ok := data[noTlsKey]
	if ok {
		notls, err := strconv.ParseBool(noTls)
		if err != nil {
			return fmt.Errorf("invalid bool value for %s: %s", noTlsKey, err)
		}
		if notls {
			data[tlsCertPathKey] = ""
		}
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

func formatVersion() string {
	return fmt.Sprintf(
		"\nVersion: %s\nCommit: %s\nDate: %s",
		version, commit, date,
	)
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

func getTransportClient(ctx *cli.Context) (tdexv1.TransportServiceClient, func(), error) {
	conn, err := getClientConn(false)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { conn.Close() }

	return tdexv1.NewTransportServiceClient(conn), cleanup, nil
}

func getOperatorClient(ctx *cli.Context) (daemonv2.OperatorServiceClient, func(), error) {
	conn, err := getClientConn(false)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { conn.Close() }

	return daemonv2.NewOperatorServiceClient(conn), cleanup, nil
}

func getWalletClient(ctx *cli.Context) (daemonv2.WalletServiceClient, func(), error) {
	conn, err := getClientConn(true)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }

	return daemonv2.NewWalletServiceClient(conn), cleanup, nil
}

func getClientConn(skipMacaroon bool) (*grpc.ClientConn, error) {
	state, err := getState()
	if err != nil {
		return nil, err
	}
	address, ok := state["rpcserver"]
	if !ok {
		return nil, errors.New("set rpcserver with `config set rpcserver`")
	}

	opts := []grpc.DialOption{grpc.WithDefaultCallOptions(maxMsgRecvSize)}

	noTls, _ := strconv.ParseBool(state[noTlsKey])
	if noTls {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		certPath, ok := state["tls_cert_path"]
		if !ok {
			return nil, fmt.Errorf(
				"TLS certificate filepath is missing. Try " +
					"'tdex config set tls_cert_path path/to/tls/certificate'",
			)
		}

		tlsCreds, err := credentials.NewClientTLSFromFile(certPath, "")
		if err != nil {
			return nil, fmt.Errorf("could not read TLS certificate:  %s", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
	}

	noMacaroons, _ := strconv.ParseBool(state["no_macaroons"])
	if !noMacaroons {
		// Load macaroons and add credentials to dialer
		if !skipMacaroon {
			macPath, ok := state["macaroons_path"]
			if !ok {
				return nil, fmt.Errorf(
					"macaroons filepath is missing. Try " +
						"'tdex config set macaroons_path path/to/macaroon",
				)
			}
			macBytes, err := ioutil.ReadFile(macPath)
			if err != nil {
				return nil, fmt.Errorf("could not read macaroon %s: %s", macPath, err)
			}
			mac := &macaroon.Macaroon{}
			if err := mac.UnmarshalBinary(macBytes); err != nil {
				return nil, fmt.Errorf("could not parse macaroon %s: %s", macPath, err)
			}
			macCreds := macaroons.NewMacaroonCredential(mac, !noTls)
			opts = append(opts, grpc.WithPerRPCCredentials(macCreds))
		}
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to RPC server: %v",
			err)
	}

	return conn, nil
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
// This function is taken from https://github.com/btcsuite/btcd
func cleanAndExpandPath(path string) string {
	if path == "" {
		return ""
	}

	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		var homeDir string
		u, err := user.Current()
		if err == nil {
			homeDir = u.HomeDir
		} else {
			homeDir = os.Getenv("HOME")
		}

		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but the variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

func printDeprecatedWarn(newCmd string) {
	colorYellow := "\033[33m"
	fmt.Println(fmt.Sprintf(
		"%sWarning: this command is deprecated and will be removed in the next "+
			"version.\nInstead, use the new command '%s'", string(colorYellow), newCmd,
	))
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
