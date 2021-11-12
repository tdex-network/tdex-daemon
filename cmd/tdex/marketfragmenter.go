package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"github.com/urfave/cli/v2"
)

var (
	marketFragmenterAccount = cli.Command{
		Name:  "marketfragmenter",
		Usage: "manage the market fragmenter account of the daemon's wallet",
		Subcommands: []*cli.Command{
			marketFragmenterBalanceCmd, marketFragmenterDepositCmd,
			marketFragmenterListAddressesCmd, marketFragmenterSplitFundsCmd,
			marketFragmenterWithdrawCmd,
		},
	}

	marketFragmenterBalanceCmd = &cli.Command{
		Name:   "balance",
		Usage:  "check the balance of the market fragmenter account",
		Action: marketFragmenterBalanceAction,
	}
	marketFragmenterDepositCmd = &cli.Command{
		Name:  "deposit",
		Usage: "generate some address(es) to receive funds",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "num_of_addresses",
				Usage: "the number of addresses to generate",
			},
		},
		Action: marketFragmenterDepositAction,
	}
	marketFragmenterListAddressesCmd = &cli.Command{
		Name:   "addresses",
		Usage:  "list all the derived deposit addresses of the market fragmenter account",
		Action: marketFragmenterListAddressesAction,
	}
	marketFragmenterSplitFundsCmd = &cli.Command{
		Name:   "split",
		Usage:  "split market fragmenter funds and make them deposits of some market",
		Action: marketFragmenterSplitFundsAction,
	}
	marketFragmenterWithdrawCmd = &cli.Command{
		Name:  "withdraw",
		Usage: "withdraw all the market fragmenter funds",
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
		},
		Action: marketFragmenterWithdrawAction,
	}
)

func marketFragmenterBalanceAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.GetMarketFragmenterBalance(
		context.Background(), &pb.GetMarketFragmenterBalanceRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}

func marketFragmenterDepositAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	numOfAddresses := ctx.Int64("num_of_addresses")
	resp, err := client.GetMarketFragmenterAddress(
		context.Background(), &pb.GetMarketFragmenterAddressRequest{
			NumOfAddresses: numOfAddresses,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}

func marketFragmenterListAddressesAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.ListMarketFragmenterAddresses(
		context.Background(), &pb.ListMarketFragmenterAddressesRequest{},
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

func marketFragmenterSplitFundsAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	stream, err := client.MarketFragmenterSplitFunds(
		context.Background(), &pb.MarketFragmenterSplitFundsRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
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

func marketFragmenterWithdrawAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	addr := ctx.String("address")
	mSatsPerByte := ctx.Uint64("millisatsperbyte")

	reply, err := client.WithdrawMarketFragmenter(
		context.Background(), &pb.WithdrawMarketFragmenterRequest{
			Address:          addr,
			MillisatsPerByte: mSatsPerByte,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}
