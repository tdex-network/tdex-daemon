package main

import (
	"context"
	"fmt"
	"io"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"github.com/urfave/cli/v2"
)

var (
	fragmenter = cli.Command{
		Name: "fragmenter",
		Usage: "make use of the fragmenter to send all funds to be split and " +
			"deposited into the Fee account to an ephemeral wallet ",
		Subcommands: []*cli.Command{
			feeFragmenterCmd, mktFragmenterCmd,
		},
	}

	feeFragmenterCmd = &cli.Command{
		Name:  "fee",
		Usage: "manage the fragmenter for the Fee account",
		Subcommands: []*cli.Command{
			feeFragmenterAddressCmd, feeFragmenterFragmentDepositsCmd,
			feeFragmenterRecoverCmd,
		},
	}
	mktFragmenterCmd = &cli.Command{
		Name:  "market",
		Usage: "manage the fragmenter for a market account",
		Subcommands: []*cli.Command{
			mktFragmenterAddressCmd, mktFragmenterFragmentDepositsCmd,
			mktFragmenterRecoverCmd,
		},
	}

	feeFragmenterAddressCmd = &cli.Command{
		Name: "getaddress",
		Usage: "get the address of the ephemeral wallet where to send funds for " +
			"the fee account. Make sure to send only LBTC to this address",
		Action: feeFragmenterAddressAction,
	}
	feeFragmenterFragmentDepositsCmd = &cli.Command{
		Name: "fragmentdeposits",
		Usage: "make the fragmenter fetch its funds and split them to fund the " +
			"Fee account",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "max_fragments",
				Usage: "specify the max number of fragments the fragment can create",
				Value: 50,
			},
		},
		Action: feeFragmenterFragmentDepositsAction,
	}
	feeFragmenterRecoverCmd = &cli.Command{
		Name: "recover",
		Usage: "abort the fragmentation and send all funds owned by the " +
			"ephemeral wallet to an address of yours",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "address",
				Usage: "specify an address where to send the funds owned by the " +
					"fragmenter to abort the process",
				Required: true,
			},
		},
		Action: feeFragmenterRecoverAction,
	}

	mktFragmenterAddressCmd = &cli.Command{
		Name: "getaddress",
		Usage: "get the address of the ephemeral wallet where to send funds for " +
			"a market account. Make sure to send funds of just the market asset pair " +
			"to this address",
		Action: mktFragmenterAddressAction,
	}
	mktFragmenterFragmentDepositsCmd = &cli.Command{
		Name: "fragmentdeposits",
		Usage: "make the fragmenter fetch its funds and split them to fund the " +
			"market account",
		Action: mktFragmenterFragmentDepositsAction,
	}
	mktFragmenterRecoverCmd = &cli.Command{
		Name: "recover",
		Usage: "abort the fragmentation and send all funds owned by the " +
			"ephemeral wallet to an address of yours",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "address",
				Usage: "specify an address where to send the funds owned by the " +
					"fragmenter to abort the process",
				Required: true,
			},
		},
		Action: mktFragmenterRecoverAction,
	}
)

func feeFragmenterAddressAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.GetFeeFragmenterAddress(
		context.Background(), &pb.GetFeeFragmenterAddressRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}

func feeFragmenterFragmentDepositsAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	maxFragments := ctx.Uint("max_fragments")

	stream, err := client.FragmentFeeDeposits(context.Background(), &pb.FragmentFeeDepositsRequest{
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

func feeFragmenterRecoverAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	recoverAddress := ctx.String("address")
	if len(recoverAddress) <= 0 {
		return fmt.Errorf("recover address is missing")
	}

	stream, err := client.FragmentFeeDeposits(context.Background(), &pb.FragmentFeeDepositsRequest{
		RecoverAddress: recoverAddress,
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

func mktFragmenterAddressAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	reply, err := client.GetMarketFragmenterAddress(
		context.Background(), &pb.GetMarketFragmenterAddressRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}

func mktFragmenterFragmentDepositsAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	stream, err := client.FragmentMarketDeposits(
		context.Background(), &pb.FragmentMarketDepositsRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
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

func mktFragmenterRecoverAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}
	recoverAddress := ctx.String("address")
	if len(recoverAddress) <= 0 {
		return fmt.Errorf("recover address is missing")
	}

	stream, err := client.FragmentMarketDeposits(
		context.Background(), &pb.FragmentMarketDepositsRequest{
			RecoverAddress: recoverAddress,
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
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
