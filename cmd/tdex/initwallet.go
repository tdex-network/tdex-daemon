package main

import (
	"context"
	"fmt"
	"strings"

	pbwallet "github.com/tdex-network/tdex-protobuf/generated/go/wallet"

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
	},
	Action: initWalletAction,
}

func initWalletAction(ctx *cli.Context) error {
	client, cleanup, err := getWalletClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	req := &pbwallet.InitWalletRequest{}

	password := ctx.String("password")
	seed := ctx.String("seed")

	if len(password) > 0 && len(seed) > 0 {
		req = &pbwallet.InitWalletRequest{
			WalletPassword: []byte(password),
			SeedMnemonic:   strings.Split(seed, " "),
		}
	}

	stream, err := client.InitWallet(
		context.Background(),
		req,
	)
	if err != nil {
		return err
	}

	_, err = stream.Recv()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Wallet is initialized. You can unlock")
	return nil
}
