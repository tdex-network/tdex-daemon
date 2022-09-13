package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"
	"github.com/urfave/cli/v2"
)

var (
	feeFragmenterAccount = cli.Command{
		Name:  "feefragmenter",
		Usage: "manage the fee fragmenter account of the daemon's wallet",
		Subcommands: []*cli.Command{
			feeFragmenterBalanceCmd, feeFragmenterDepositCmd,
			feeFragmenterListAddressesCmd, feeFragmenterSplitFundsCmd,
			feeFragmenterWithdrawCmd,
		},
	}

	feeFragmenterBalanceCmd = &cli.Command{
		Name:   "balance",
		Usage:  "check the balance of the fee fragmenter account",
		Action: feeFragmenterBalanceAction,
	}
	feeFragmenterDepositCmd = &cli.Command{
		Name:  "deposit",
		Usage: "generate some address(es) to receive funds",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "num_of_addresses",
				Usage: "the number of addresses to generate",
			},
		},
		Action: feeFragmenterDepositAction,
	}
	feeFragmenterListAddressesCmd = &cli.Command{
		Name:   "addresses",
		Usage:  "list all the derived deposit addresses of the fee fragmenter account",
		Action: feeFragmenterListAddressesAction,
	}
	feeFragmenterSplitFundsCmd = &cli.Command{
		Name:   "split",
		Usage:  "split fee fragmenter funds and make them deposits of the fee acount",
		Action: feeFragmenterSplitFundsAction,
	}
	feeFragmenterWithdrawCmd = &cli.Command{
		Name:  "withdraw",
		Usage: "withdraw all the fee fragmenter funds",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "address",
				Usage: "the address of the receiver of the funds",
			},
			&cli.Uint64Flag{
				Name:  "millisatsperbyte",
				Usage: "the mSat/byte to pay for the transaction",
				Value: 100,
			},
			&cli.StringFlag{
				Name:     "password",
				Usage:    "the wallet unlocking password as security measure",
				Required: true,
			},
		},
		Action: feeFragmenterWithdrawAction,
	}
)

func feeFragmenterBalanceAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.GetFeeFragmenterBalance(
		context.Background(), &daemonv1.GetFeeFragmenterBalanceRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}

func feeFragmenterDepositAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	numOfAddresses := ctx.Int64("num_of_addresses")
	resp, err := client.GetFeeFragmenterAddress(
		context.Background(), &daemonv1.GetFeeFragmenterAddressRequest{
			NumOfAddresses: numOfAddresses,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}

func feeFragmenterListAddressesAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.ListFeeFragmenterAddresses(
		context.Background(), &daemonv1.ListFeeFragmenterAddressesRequest{},
	)
	if err != nil {
		return err
	}

	list := reply.GetAddressWithBlindingKey()
	if list == nil {
		fmt.Println("[]")
		return nil
	}

	listStr, _ := json.MarshalIndent(list, "", "   ")
	fmt.Println(string(listStr))

	return nil
}

func feeFragmenterSplitFundsAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	maxFragments := ctx.Uint64("max_fragments")

	stream, err := client.FeeFragmenterSplitFunds(
		context.Background(), &daemonv1.FeeFragmenterSplitFundsRequest{
			MaxFragments: uint32(maxFragments),
		})
	if err != nil {
		return err
	}

	for {
		fmt.Println()

		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		fmt.Println(reply.GetMessage())
	}

	return nil
}

func feeFragmenterWithdrawAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	addr := ctx.String("address")
	password := ctx.String("password")
	mSatsPerByte := ctx.Uint64("millisatsperbyte")

	reply, err := client.WithdrawFeeFragmenter(
		context.Background(), &daemonv1.WithdrawFeeFragmenterRequest{
			Address:          addr,
			MillisatsPerByte: mSatsPerByte,
			Password:         password,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}
