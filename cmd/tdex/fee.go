package main

import (
	"context"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	"github.com/urfave/cli/v2"
)

var (
	feeAccount = cli.Command{
		Name:  "fee",
		Usage: "manage the fee account of the daemon's wallet",
		Subcommands: []*cli.Command{
			feeBalanceCmd, feeDepositCmd, feeListAddressesCmd, feeWithdrawCmd,
		},
	}

	feeBalanceCmd = &cli.Command{
		Name:   "balance",
		Usage:  "check the balance of the fee account",
		Action: feeBalanceAction,
	}
	feeDepositCmd = &cli.Command{
		Name:  "deposit",
		Usage: "generate some address(es) to receive funds",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "num-of-addresses",
				Usage: "the number of addresses to generate",
			},
		},
		Action: feeDepositAction,
	}
	feeListAddressesCmd = &cli.Command{
		Name:   "addresses",
		Usage:  "list all the derived deposit addresses of the fee account",
		Action: feeListAddressesAction,
	}
	feeWithdrawCmd = &cli.Command{
		Name:  "withdraw",
		Usage: "withdraw funds from fee account",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "receivers",
				Usage:    "list of withdrawal receivers as {asset, amount, address}",
				Required: true,
			},
			&cli.Uint64Flag{
				Name:  "milli-sats-per-byte",
				Usage: "the mSat/byte to pay for the transaction",
				Value: 100,
			},
			&cli.StringFlag{
				Name:     "password",
				Usage:    "the wallet unlocking password as security measure",
				Required: true,
			},
		},
		Action: feeWithdrawAction,
	}
)

func feeBalanceAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.GetFeeBalance(context.Background(), &daemonv2.GetFeeBalanceRequest{})
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}

func feeDepositAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	numOfAddresses := ctx.Int64("num-of-addresses")
	reply, err := client.DeriveFeeAddresses(
		context.Background(), &daemonv2.DeriveFeeAddressesRequest{
			NumOfAddresses: numOfAddresses,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}

func feeListAddressesAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.ListFeeAddresses(
		context.Background(), &daemonv2.ListFeeAddressesRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}

func feeWithdrawAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	receivers := ctx.StringSlice("receivers")
	password := ctx.String("password")
	mSatsPerByte := ctx.Uint64("milli-sats-per-byte")
	outputs, err := parseOutputs(receivers)
	if err != nil {
		return err
	}

	reply, err := client.WithdrawFee(context.Background(), &daemonv2.WithdrawFeeRequest{
		Outputs:          outputs,
		MillisatsPerByte: mSatsPerByte,
		Password:         password,
	})
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}
