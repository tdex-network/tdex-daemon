package main

import (
	"context"
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
	client, cleanup, err := getWalletClient(ctx)
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

	m := make(map[int]struct{})
	var reply *pbwallet.InitWalletReply
	var prevReply *pbwallet.InitWalletReply
	for {
		prevReply = reply
		reply, err = stream.Recv()
		if err == io.EOF {
			if prevReply != nil {
				fmt.Println("restore account", prevReply.GetAccount(), prevReply.GetStatus())
			}
			break
		}
		if err != nil {
			return err
		}

		status := reply.GetStatus()
		data := reply.GetData()
		if strings.Contains(data, "addresses") {
			fmt.Println(data, status)
			continue
		}

		prevAccount := prevReply.GetAccount()
		prevStatus := prevReply.GetStatus()
		account := reply.GetAccount()
		if status == pbwallet.InitWalletReply_PROCESSING {
			if prevStatus == pbwallet.InitWalletReply_DONE && account != prevAccount {
				fmt.Println("restore account", prevAccount, prevStatus)
			}

			if _, ok := m[int(account)]; !ok {
				fmt.Println("restore account", account, status)
				m[int(account)] = struct{}{}
			}
		}
	}

	fmt.Println()
	fmt.Println("Wallet is initialized. You can unlock")
	return nil
}
