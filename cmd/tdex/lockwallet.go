package main

import (
	"context"
	"fmt"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"

	"github.com/urfave/cli/v2"
)

var lockwallet = cli.Command{
	Name:  "lock",
	Usage: "lock the daemon wallet",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "password",
			Usage:    "the (un)locking password",
			Value:    "",
			Required: true,
		},
	},
	Action: lockWalletAction,
}

func lockWalletAction(ctx *cli.Context) error {
	client, cleanup, err := getWalletClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	_, err = client.LockWallet(
		context.Background(), &daemonv2.LockWalletRequest{
			Password: ctx.String("password"),
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Wallet is locked")
	return nil
}
