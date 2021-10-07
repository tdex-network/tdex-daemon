package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	pbwallet "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/walletunlocker"

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

	req := &pbwallet.InitWalletRequest{}

	password := ctx.String("password")
	seed := ctx.String("seed")
	restore := ctx.Bool("restore")

	if len(password) > 0 && len(seed) > 0 {
		req = &pbwallet.InitWalletRequest{
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
			fmt.Println("restore account", account, status)
			continue
		}

		if _, err := hex.DecodeString(data); err == nil {
			fmt.Println("admin.macaroon", data)
		} else {
			fmt.Println(data, status)
		}
	}

	fmt.Println()
	fmt.Println("Wallet is initialized. You can unlock")
	return nil
}
