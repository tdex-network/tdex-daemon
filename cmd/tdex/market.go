package main

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"github.com/urfave/cli/v2"
)

var (
	marketAccount = cli.Command{
		Name:  "market",
		Usage: "manage a market account of the daemon's wallet",
		Subcommands: []*cli.Command{
			marketNewCmd, marketInfoCmd, marketBalanceCmd, marketListAddressesCmd,
			marketDepositCmd, marketClaimCmd, marketWithdrawCmd,
			marketOpenCmd, marketCloseCmd, marketDropCmd,
			marketUpdateFixedFeeCmd, marketUpdatePercentageFeeCmd, marketReportFeeCmd,
			marketUpdateStrategyCmd, marketUpdatePriceCmd,
		},
	}

	marketNewCmd = &cli.Command{
		Name:  "new",
		Usage: "create a new market",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "base_asset",
				Usage: "the base asset hash of an existent market",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "quote_asset",
				Usage: "the base asset hash of an existent market",
				Value: "",
			},
		},
		Action: newMarketAction,
	}
	marketInfoCmd = &cli.Command{
		Name:   "info",
		Usage:  "get info about the current market",
		Action: marketInfoAction,
	}
	marketBalanceCmd = &cli.Command{
		Name:   "balance",
		Usage:  "check the balance of a market",
		Action: marketBalanceAction,
	}
	marketDepositCmd = &cli.Command{
		Name:  "deposit",
		Usage: "generate some address(es) to deposit funds for a market",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "num_of_addresses",
				Usage: "the number of addresses to generate for the market",
			},
			&cli.BoolFlag{
				Name: "fragment",
				Usage: "send funds to an ephemeral wallet to be split into multiple " +
					"fragments and deposited into the market account",
			},
			&cli.StringFlag{
				Name: "recover_funds_to_address",
				Usage: "specify an address where to send the funds owned by the " +
					"fragmenter to abort the process",
			},
		},
		Action: marketDepositAction,
	}
	marketListAddressesCmd = &cli.Command{
		Name:   "addresses",
		Usage:  "list all the derived deposit addresses of a market",
		Action: marketListAddressesAction,
	}
	marketClaimCmd = &cli.Command{
		Name:  "claim",
		Usage: "claim deposits for a market",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "outpoints",
				Usage: "list of outpoints referring to utxos [{\"hash\": <string>, \"index\": <number>}]",
			},
		},
		Action: marketClaimAction,
	}
	marketWithdrawCmd = &cli.Command{
		Name:  "withdraw",
		Usage: "withdraw some funds to an address",
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:  "base_amount",
				Usage: "the amount in Satoshi of base asset to withdraw from the market.",
			},
			&cli.Uint64Flag{
				Name:  "quote_amount",
				Usage: "the amount in Satoshi of quote asset to withdraw from the market.",
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
		Action: marketWithdrawAction,
	}
	marketOpenCmd = &cli.Command{
		Name:   "open",
		Usage:  "open a market",
		Action: marketOpenAction,
	}
	marketCloseCmd = &cli.Command{
		Name:   "close",
		Usage:  "close a market",
		Action: marketCloseAction,
	}
	marketDropCmd = &cli.Command{
		Name:   "drop",
		Usage:  "drop a market",
		Action: marketDropAction,
	}
	marketUpdateFixedFeeCmd = &cli.Command{
		Name:  "fixedfee",
		Usage: "updates the current market fixed fee",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:  "base_fee",
				Usage: "set the fixed fee for base asset",
				Value: -1,
			},
			&cli.Int64Flag{
				Name:  "quote_fee",
				Usage: "set the fixed fee for quote asset",
				Value: -1,
			},
		},
		Action: marketUpdateFixedFeeAction,
	}
	marketUpdatePercentageFeeCmd = &cli.Command{
		Name:  "percentagefee",
		Usage: "updates the current market percentage fee",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:  "basis_point",
				Usage: "set the fee basis point",
			},
		},
		Action: marketUpdatePercentageFeeAction,
	}
	marketReportFeeCmd = &cli.Command{
		Name:  "reportfee",
		Usage: "get a report of the fees collected for the trades of a market.",
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:  "page",
				Usage: "the number of the page to be listed. If omitted, the entire list is returned",
			},
			&cli.Uint64Flag{
				Name:  "page_size",
				Usage: "the size of the page",
				Value: 10,
			},
		},
		Action: marketReportFeeAction,
	}
	marketUpdateStrategyCmd = &cli.Command{
		Name:  "strategy",
		Usage: "updates the current market making strategy, either automated or pluggable market making",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "pluggable",
				Usage: "set the strategy as pluggable",
				Value: false,
			},
		},
		Action: marketUpdateStrategyAction,
	}

	marketUpdatePriceCmd = &cli.Command{
		Name:  "price",
		Usage: "updates the price of a market",
		Flags: []cli.Flag{
			&cli.Float64Flag{
				Name:     "base_price",
				Usage:    "the base price, or the amount of quote asset needed to buy 1 BTC of base asset",
				Required: true,
			},
			&cli.Float64Flag{
				Name:     "quote_price",
				Usage:    "the quote price, or the amount of base asset need to buy 1 BTC of quote asset",
				Required: true,
			},
		},
		Action: marketUpdatePriceAction,
	}
)

func newMarketAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset := ctx.String("base_asset")
	quoteAsset := ctx.String("quote_asset")

	if _, err := client.NewMarket(
		context.Background(), &pb.NewMarketRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market created")
	return nil
}

func marketInfoAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	resp, err := client.GetMarketInfo(
		context.Background(), &pb.GetMarketInfoRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

func marketBalanceAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	resp, err := client.GetMarketBalance(
		context.Background(), &pb.GetMarketBalanceRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

func marketDepositAction(ctx *cli.Context) error {
	if withFragmenter := ctx.Bool("fragment"); withFragmenter {
		return marketFragmentDepositAction(ctx)
	}

	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	numOfAddresses := ctx.Int64("num_of_addresses")
	resp, err := client.GetMarketAddress(
		context.Background(),
		&pb.GetMarketAddressRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			NumOfAddresses: numOfAddresses,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

func marketFragmentDepositAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex marketfragmenter split")
	return nil
}

func marketListAddressesAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	resp, err := client.ListMarketAddresses(
		context.Background(), &pb.ListMarketAddressesRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	list := resp.GetAddressWithBlindingKey()
	if list == nil {
		fmt.Println("[]")
		return nil
	}

	listStr, _ := json.MarshalIndent(list, "", "   ")
	fmt.Println(string(listStr))
	return nil
}

func marketClaimAction(ctx *cli.Context) error {
	outpoints, err := parseOutpoints(ctx.String("outpoints"))
	if err != nil {
		return err
	}

	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	if _, err := client.ClaimMarketDeposits(
		context.Background(), &pb.ClaimMarketDepositsRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			Outpoints: outpoints,
		},
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market is funded")
	return nil
}

func marketWithdrawAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}
	baseAmount := ctx.Uint64("base_amount")
	quoteAmount := ctx.Uint64("quote_amount")
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
		Address:          addr,
		MillisatsPerByte: mSatsPerByte,
	})
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}

func marketOpenAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	_, err = client.OpenMarket(
		context.Background(), &pb.OpenMarketRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market is open")
	return nil
}

func marketCloseAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	_, err = client.CloseMarket(
		context.Background(), &pb.CloseMarketRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market is closed")
	return nil
}

func marketDropAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	_, err = client.DropMarket(
		context.Background(), &pb.DropMarketRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market is dropped")
	return nil
}

func marketUpdateFixedFeeAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	baseFee := ctx.Int64("base_fee")
	quoteFee := ctx.Int64("quote_fee")
	req := &pb.UpdateMarketFixedFeeRequest{
		Market: &pbtypes.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		Fixed: &pbtypes.Fixed{
			BaseFee:  baseFee,
			QuoteFee: quoteFee,
		},
	}

	if _, err := client.UpdateMarketFixedFee(
		context.Background(), req,
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market fees have been updated")
	return nil
}

func marketUpdatePercentageFeeAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	basisPoint := ctx.Int64("basis_point")
	req := &pb.UpdateMarketPercentageFeeRequest{
		Market: &pbtypes.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		BasisPoint: basisPoint,
	}

	if _, err := client.UpdateMarketPercentageFee(
		context.Background(), req,
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market fees have been updated")
	return nil
}

func marketReportFeeAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	pageNumber := ctx.Int64("page")
	pageSize := ctx.Int64("page_size")
	var page *pb.Page
	if pageNumber > 0 {
		page = &pb.Page{
			PageNumber: pageNumber,
			PageSize:   pageSize,
		}
	}

	reply, err := client.GetMarketCollectedSwapFees(
		context.Background(), &pb.GetMarketCollectedSwapFeesRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			Page: page,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}

func marketUpdateStrategyAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	strategy := pb.StrategyType_BALANCED
	if ctx.Bool("pluggable") {
		strategy = pb.StrategyType_PLUGGABLE
	}

	_, err = client.UpdateMarketStrategy(
		context.Background(), &pb.UpdateMarketStrategyRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			StrategyType: strategy,
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("strategy has been updated")
	return nil
}

func marketUpdatePriceAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	_, err = client.UpdateMarketPrice(
		context.Background(), &pb.UpdateMarketPriceRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			Price: &pbtypes.Price{
				BasePrice:  float32(ctx.Float64("base_price")),
				QuotePrice: float32(ctx.Float64("quote_price")),
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("price has been updated")
	return nil
}
