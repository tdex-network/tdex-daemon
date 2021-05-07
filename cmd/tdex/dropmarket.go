package main

import (
	"context"
	"fmt"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"

	"github.com/urfave/cli/v2"
)

var dropmarket = cli.Command{
	Name:  "dropmarket",
	Usage: "drop a market",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "account_index",
			Usage: "the account index of the market to drop",
		},
	},
	Action: dropMarketAction,
}

func dropMarketAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	accountIndex := ctx.Uint64("account_index")

	_, err = client.DropMarket(
		context.Background(), &pboperator.DropMarketRequest{
			AccountIndex: accountIndex,
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market is dropped")
	return nil
}
