package main

import (
	"context"
	"fmt"
	"io"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
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
				Name:  "num-of-addresses",
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
		Name:  "split",
		Usage: "split fee fragmenter funds and make them deposits of the fee account",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "num-fragments",
				Usage: "Number of fragmented utxos to generate from fee fragmenter account balance",
			},
		},
		Action: feeFragmenterSplitFundsAction,
	}
	feeFragmenterWithdrawCmd = &cli.Command{
		Name:  "withdraw",
		Usage: "withdraw funds from fee fragmenter account",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "receivers",
				Usage: "list of withdrawal receivers as  {aseet, amount, address}",
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
		context.Background(), &daemonv2.GetFeeFragmenterBalanceRequest{},
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

	numOfAddresses := ctx.Int64("num-of-addresses")
	resp, err := client.DeriveFeeFragmenterAddresses(
		context.Background(), &daemonv2.DeriveFeeFragmenterAddressesRequest{
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
		context.Background(), &daemonv2.ListFeeFragmenterAddressesRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}

func feeFragmenterSplitFundsAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	numFragments := ctx.Int("num-fragments")

	stream, err := client.FeeFragmenterSplitFunds(
		context.Background(), &daemonv2.FeeFragmenterSplitFundsRequest{
			MaxFragments: uint32(numFragments),
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

	receivers := ctx.StringSlice("receivers")
	password := ctx.String("password")
	mSatsPerByte := ctx.Uint64("millisatsperbyte")
	outputs, err := parseOutputs(receivers)
	if err != nil {
		return err
	}

	reply, err := client.WithdrawFeeFragmenter(
		context.Background(), &daemonv2.WithdrawFeeFragmenterRequest{
			Outputs:          outputs,
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
