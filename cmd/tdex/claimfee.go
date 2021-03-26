package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"

	"github.com/urfave/cli/v2"
)

var claimfee = cli.Command{
	Name:  "claimfee",
	Usage: "claim deposits for the fee account",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "outpoints",
			Usage: "list of outpoints referring to utxos [{\"hash\": <string>, \"index\": <number>}]",
		},
	},
	Action: claimFeeAction,
}

func claimFeeAction(ctx *cli.Context) error {
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
		context.Background(), &pboperator.ClaimFeeDepositRequest{
			Outpoints: outpoints,
		},
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("fee account is funded")

	return nil
}

func parseOutpoints(str string) ([]*pboperator.TxOutpoint, error) {
	var outpoints []*pboperator.TxOutpoint
	if err := json.Unmarshal([]byte(str), &outpoints); err != nil {
		return nil, errors.New("unable to parse provided outpoints")
	}
	return outpoints, nil
}
