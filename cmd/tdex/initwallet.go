package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex-daemon/v1"

	"github.com/urfave/cli/v2"
)

var initwallet = cli.Command{
	Name:  "init",
	Usage: "initialize the daemon and its internal wallet",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "password",
			Usage: "the password used to encrypt the mnemonic",
		},
		&cli.StringFlag{
			Name:  "seed",
			Usage: "the mnemonic seed of the daemon wallet",
		},
		&cli.BoolFlag{
			Name:  "restore",
			Value: false,
			Usage: "whether restore existing funds for the wallet",
		},
	},
	Action: initWalletAction,
}

func initWalletAction(ctx *cli.Context) error {
	client, cleanup, err := getUnlockerClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	state, err := getState()
	if err != nil {
		return err
	}

	req := &daemonv1.InitWalletRequest{}

	password := ctx.String("password")
	seed := ctx.String("seed")
	restore := ctx.Bool("restore")

	if len(password) > 0 && len(seed) > 0 {
		req = &daemonv1.InitWalletRequest{
			WalletPassword: []byte(password),
			SeedMnemonic:   strings.Split(seed, " "),
			Restore:        restore,
		}
	}

	stream, err := client.InitWallet(
		context.Background(),
		req,
	)
	if err != nil {
		return err
	}

	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		account := reply.GetAccount()
		status := reply.GetStatus()
		data := reply.GetData()
		if account >= 0 {
			fmt.Println(data, account, status)
			continue
		}

		if macaroon, err := hex.DecodeString(data); err == nil {
			fmt.Println("admin.macaroon", data)
			// In case the CLI has been configured with a tdexdconnect URL,
			// the macaroon is written to a file in the CLI's datadir and the
			// macaroons_path is updated in the config file.
			// To know that, let's check if the TLS certificate file is inside the
			// CLI's datadir. This suggests that the 'connect' command was used.
			tlsCertPath := state["tls_cert_path"]
			if ok, _ := filepath.Match(tdexDataDir, filepath.Dir(tlsCertPath)); ok {
				macPath := filepath.Join(tdexDataDir, "admin.macaroon")
				if err := ioutil.WriteFile(macPath, macaroon, 0644); err != nil {
					return fmt.Errorf("failed to write macaroon to file: %s", err)
				}
				if err := setState(
					map[string]string{"macaroons_path": macPath},
				); err != nil {
					return fmt.Errorf(
						"an error occured while setting 'macaroons_path' in config: %s.\n"+
							"Please run 'tdex config set macaroons_path %s'", err, macPath,
					)
				}
			}
		} else {
			fmt.Println(data, status)
		}
	}

	fmt.Println()
	fmt.Println("Wallet is initialized. You can unlock")
	return nil
}
