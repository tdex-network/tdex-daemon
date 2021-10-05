package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
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

	reply, err := client.GetFeeBalance(context.Background(), &pb.GetFeeBalanceRequest{})
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

	numOfAddresses := ctx.Int64("num_of_addresses")
	resp, err := client.GetFeeAddress(
		context.Background(), &pb.GetFeeAddressRequest{
			NumOfAddresses: numOfAddresses,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}

func feeListAddressesAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.ListFeeAddresses(
		context.Background(), &pb.ListFeeAddressesRequest{},
	)
	if err != nil {
		return err
	}

	list := reply.GetAddressWithBlinidngKey()
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
		context.Background(), &pb.ClaimFeeDepositsRequest{
			Outpoints: outpoints,
		},
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("fee account is funded")

	return nil
}

func parseOutpoints(str string) ([]*pb.TxOutpoint, error) {
	var outpoints []*pb.TxOutpoint
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
	mSatsPerByte := ctx.Uint64("millisatsperbyte")
	asset := ctx.String("asset")

	reply, err := client.WithdrawFee(context.Background(), &pb.WithdrawFeeRequest{
		Amount:           amount,
		Address:          addr,
		Asset:            asset,
		MillisatsPerByte: mSatsPerByte,
	})
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}
