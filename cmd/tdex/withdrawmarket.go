package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"github.com/urfave/cli/v2"
)

var withdrawmarket = cli.Command{
	Name:  "withdrawmarket",
	Usage: "withdraw funds from some market.",
	Flags: []cli.Flag{
		&cli.Float64Flag{
			Name:  "base_amount",
			Usage: "the amount in BTC of base asset to withdraw from the market.",
		},
		&cli.Float64Flag{
			Name:  "quote_amount",
			Usage: "the amount in BTC of quote asset to withdraw from the market.",
		},
		&cli.StringFlag{
			Name:  "address",
			Usage: "the address where to send the withdrew amount(s).",
		},
		&cli.Int64Flag{
			Name:  "millisatsperbyte",
			Usage: "the mSat/byte to pay for the transaction",
			Value: 100,
		},
	},
	Action: withdrawMarketAction,
}

func withdrawMarketAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}
	baseAmount := uint64(ctx.Float64("base_amount") * math.Pow10(8))
	quoteAmount := uint64(ctx.Float64("quote_amount") * math.Pow10(8))
	addr := ctx.String("address")
	mSatsPerByte := ctx.Int64("millisatsperbyte")

	reply, err := client.WithdrawMarket(context.Background(), &pb.WithdrawMarketRequest{
		Market: &pbtypes.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		BalanceToWithdraw: &pbtypes.Balance{
			BaseAmount:  baseAmount,
			QuoteAmount: quoteAmount,
		},
		Address:         addr,
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
