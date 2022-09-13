package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"
	"github.com/urfave/cli/v2"
)

var (
	feeAccount = cli.Command{
		Name:  "fee",
		Usage: "manage the fee account of the daemon's wallet",
		Subcommands: []*cli.Command{
			feeBalanceCmd, feeDepositCmd, feeListAddressesCmd, feeClaimCmd,
			feeWithdrawCmd,
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
				Name:  "num_of_addresses",
				Usage: "the number of addresses to generate",
			},
			&cli.BoolFlag{
				Name: "fragment",
				Usage: "send funds to an ephemeral wallet to be split into multiple " +
					"fragments and deposited into the Fee account",
			},
			&cli.UintFlag{
				Name: "max_fragments",
				Usage: "specify the max number of fragments the fragmenter can " +
					"create when splitting its funds",
				Value: 50,
			},
			&cli.StringFlag{
				Name: "recover_funds_to_address",
				Usage: "specify an address where to send the funds owned by the " +
					"fragmenter to abort the process",
			},
		},
		Action: feeDepositAction,
	}
	feeListAddressesCmd = &cli.Command{
		Name:   "addresses",
		Usage:  "list all the derived deposit addresses of the fee account",
		Action: feeListAddressesAction,
	}
	feeClaimCmd = &cli.Command{
		Name:  "claim",
		Usage: "claim deposits for the fee account",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "outpoints",
				Usage: "list of outpoints referring to utxos [{\"hash\": <string>, \"index\": <number>}]",
			},
		},
		Action: feeClaimAction,
	}
	feeWithdrawCmd = &cli.Command{
		Name:  "withdraw",
		Usage: "withdraw some funds to an address",
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:  "amount",
				Usage: "the amount in Satoshi to withdraw",
			},
			&cli.StringFlag{
				Name:  "address",
				Usage: "the address of the receiver of the funds",
			},
			&cli.StringFlag{
				Name:  "asset",
				Usage: "the asset of the funds to withdraw",
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
		Action: feeWithdrawAction,
	}
)

func feeBalanceAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.GetFeeBalance(context.Background(), &daemonv1.GetFeeBalanceRequest{})
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}

func feeDepositAction(ctx *cli.Context) error {
	if withFragmenter := ctx.Bool("fragment"); withFragmenter {
		return feeFragmentDepositAction(ctx)
	}

	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	numOfAddresses := ctx.Int64("num_of_addresses")
	resp, err := client.GetFeeAddress(
		context.Background(), &daemonv1.GetFeeAddressRequest{
			NumOfAddresses: numOfAddresses,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}

func feeFragmentDepositAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex feefragmenter split")
	return nil
}

func feeListAddressesAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.ListFeeAddresses(
		context.Background(), &daemonv1.ListFeeAddressesRequest{},
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

func feeClaimAction(ctx *cli.Context) error {
	outpoints, err := parseOutpoints(ctx.String("outpoints"))
	if err != nil {
		return err
	}

	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	if _, err := client.ClaimFeeDeposits(
		context.Background(), &daemonv1.ClaimFeeDepositsRequest{
			Outpoints: outpoints,
		},
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("fee account is funded")

	return nil
}

func parseOutpoints(str string) ([]*daemonv1.Outpoint, error) {
	var outpoints []*daemonv1.Outpoint
	if err := json.Unmarshal([]byte(str), &outpoints); err != nil {
		return nil, errors.New("unable to parse provided outpoints")
	}
	return outpoints, nil
}

func feeWithdrawAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	amount := ctx.Uint64("amount")
	addr := ctx.String("address")
	password := ctx.String("password")
	mSatsPerByte := ctx.Uint64("millisatsperbyte")
	asset := ctx.String("asset")

	reply, err := client.WithdrawFee(context.Background(), &daemonv1.WithdrawFeeRequest{
		Amount:           amount,
		Address:          addr,
		Asset:            asset,
		MillisatsPerByte: mSatsPerByte,
		Password:         password,
	})
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}
