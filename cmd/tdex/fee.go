package main

import (
	"context"
	"encoding/hex"
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
			feeBalanceCmd, feeDepositCmd, feeClaimCmd, feeWithdrawCmd,
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
				Usage: "the amount in Satoshi to wtìithdraw",
			},
			&cli.StringFlag{
				Name:  "address",
				Usage: "the address of the receiver of the funds",
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

	reply, err := client.BalanceFeeAccount(context.Background(), &pb.BalanceFeeAccountRequest{})
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
	resp, err := client.DepositFeeAccount(
		context.Background(), &pb.DepositFeeAccountRequest{
			NumOfAddresses: numOfAddresses,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

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

	if _, err := client.ClaimFeeDeposit(
		context.Background(), &pb.ClaimFeeDepositRequest{
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

	reply, err := client.WithdrawFee(context.Background(), &pb.WithdrawFeeRequest{
		Amount:           amount,
		Address:          addr,
		MillisatsPerByte: mSatsPerByte,
		Push:             true,
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
