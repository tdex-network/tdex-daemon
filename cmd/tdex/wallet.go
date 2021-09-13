package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/wallet"
	"github.com/urfave/cli/v2"
)

var walletAccount = cli.Command{
	Name:  "wallet",
	Usage: "send or receive funds of the daemon wallet account",
	Subcommands: []*cli.Command{
		{
			Name:   "balance",
			Usage:  "check the balance of the wallet account",
			Action: walletBalanceAction,
		},
		{
			Name:   "receive",
			Usage:  "generate an address to receive funds",
			Action: walletReceiveAction,
		},
		{
			Name:   "send",
			Usage:  "send some funds to one or more addresses",
			Action: walletSendAction,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "asset",
					Usage: "the hash of the asset to send",
				},
				&cli.Float64Flag{
					Name:  "amount",
					Usage: "the amount (in BTC) to send",
				},
				&cli.StringFlag{
					Name:  "address",
					Usage: "the address of the receiver of the funds",
				},
				&cli.IntFlag{
					Name:  "millisatsperbyte",
					Usage: "the mSat/byte to pay for the transaction",
					Value: 100,
				},
			},
		},
	},
}

func walletBalanceAction(ctx *cli.Context) error {
	client, cleanup, err := getWalletClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.WalletBalance(context.Background(), &pb.WalletBalanceRequest{})
	if err != nil {
		return err
	}

	balance := reply.GetBalance()
	if balance == nil {
		fmt.Println("{}")
		return nil
	}

	resStr, _ := json.MarshalIndent(balance, "", "   ")
	fmt.Println(string(resStr))

	return nil
}

func walletReceiveAction(ctx *cli.Context) error {
	client, cleanup, err := getWalletClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.WalletAddress(context.Background(), &pb.WalletAddressRequest{})
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}

func walletSendAction(ctx *cli.Context) error {
	client, cleanup, err := getWalletClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	asset := ctx.String("asset")
	amount := int64((ctx.Float64("amount")) * math.Pow10(8))
	addr := ctx.String("address")
	mSatsPerByte := ctx.Int64("millisatsperbyte")

	out := &pb.TxOut{
		Asset:   asset,
		Value:   amount,
		Address: addr,
	}

	reply, err := client.SendToMany(context.Background(), &pb.SendToManyRequest{
		Outputs:         []*pb.TxOut{out},
		MillisatPerByte: mSatsPerByte,
		Push:            true,
	})
	if err != nil {
		return err
	}

	res := map[string]string{
		"txid": hex.EncodeToString(reply.GetTxid()),
	}

	resStr, _ := json.MarshalIndent(res, "", "\t")

	fmt.Println(string(resStr))

	return nil
}
