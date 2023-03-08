package grpchandler

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"
	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type operatorHandler struct {
	tradeAddress    string
	operatorAddress string
	operatorSvc     application.OperatorService
}

// NewOperatorHandler is a constructor function returning an protobuf OperatorServer.
func NewOperatorHandler(
	operatorSvc application.OperatorService,
	tradeAddress string,
	operatorAddress string,
) daemonv1.OperatorServiceServer {
	return newOperatorHandler(operatorSvc, tradeAddress, operatorAddress)
}

func newOperatorHandler(
	operatorSvc application.OperatorService,
	tradeAddress string,
	operatorAddress string,
) *operatorHandler {
	return &operatorHandler{
		operatorSvc:     operatorSvc,
		tradeAddress:    tradeAddress,
		operatorAddress: operatorAddress,
	}
}

func (o operatorHandler) GetInfo(
	ctx context.Context, req *daemonv1.GetInfoRequest,
) (*daemonv1.GetInfoResponse, error) {
	return o.getInfo(ctx, req)
}

func (o operatorHandler) GetFeeAddress(
	ctx context.Context, req *daemonv1.GetFeeAddressRequest,
) (*daemonv1.GetFeeAddressResponse, error) {
	return o.getFeeAddress(ctx, req)
}

func (o operatorHandler) ListFeeAddresses(
	ctx context.Context, req *daemonv1.ListFeeAddressesRequest,
) (*daemonv1.ListFeeAddressesResponse, error) {
	return o.listFeeAddresses(ctx, req)
}

func (o operatorHandler) GetFeeBalance(
	ctx context.Context, req *daemonv1.GetFeeBalanceRequest,
) (*daemonv1.GetFeeBalanceResponse, error) {
	return o.getFeeBalance(ctx, req)
}

func (o operatorHandler) ClaimFeeDeposits(
	ctx context.Context, req *daemonv1.ClaimFeeDepositsRequest,
) (*daemonv1.ClaimFeeDepositsResponse, error) {
	return o.claimFeeDeposits(ctx, req)
}

func (o operatorHandler) WithdrawFee(
	ctx context.Context, req *daemonv1.WithdrawFeeRequest,
) (*daemonv1.WithdrawFeeResponse, error) {
	return o.withdrawFee(ctx, req)
}

func (o operatorHandler) NewMarket(
	ctx context.Context, req *daemonv1.NewMarketRequest,
) (*daemonv1.NewMarketResponse, error) {
	return o.newMarket(ctx, req)
}

func (o operatorHandler) GetMarketInfo(
	ctx context.Context, req *daemonv1.GetMarketInfoRequest,
) (*daemonv1.GetMarketInfoResponse, error) {
	return o.getMarketInfo(ctx, req)
}

func (o operatorHandler) GetMarketAddress(
	ctx context.Context, req *daemonv1.GetMarketAddressRequest,
) (*daemonv1.GetMarketAddressResponse, error) {
	return o.getMarketAddress(ctx, req)
}

func (o operatorHandler) ListMarketAddresses(
	ctx context.Context, req *daemonv1.ListMarketAddressesRequest,
) (*daemonv1.ListMarketAddressesResponse, error) {
	return o.listMarketAddresses(ctx, req)
}

func (o operatorHandler) GetMarketBalance(
	ctx context.Context, req *daemonv1.GetMarketBalanceRequest,
) (*daemonv1.GetMarketBalanceResponse, error) {
	return o.getMarketBalance(ctx, req)
}

func (o operatorHandler) ClaimMarketDeposits(
	ctx context.Context, req *daemonv1.ClaimMarketDepositsRequest,
) (*daemonv1.ClaimMarketDepositsResponse, error) {
	return o.claimMarketDeposits(ctx, req)
}

func (o operatorHandler) OpenMarket(
	ctx context.Context, req *daemonv1.OpenMarketRequest,
) (*daemonv1.OpenMarketResponse, error) {
	return o.openMarket(ctx, req)
}

func (o operatorHandler) CloseMarket(
	ctx context.Context, req *daemonv1.CloseMarketRequest,
) (*daemonv1.CloseMarketResponse, error) {
	return o.closeMarket(ctx, req)
}

func (o operatorHandler) DropMarket(
	ctx context.Context, req *daemonv1.DropMarketRequest,
) (*daemonv1.DropMarketResponse, error) {
	return o.dropMarket(ctx, req)
}

func (o operatorHandler) GetMarketCollectedSwapFees(
	ctx context.Context, req *daemonv1.GetMarketCollectedSwapFeesRequest,
) (*daemonv1.GetMarketCollectedSwapFeesResponse, error) {
	return o.getMarketCollectedSwapFees(ctx, req)
}

func (o operatorHandler) WithdrawMarket(
	ctx context.Context, req *daemonv1.WithdrawMarketRequest,
) (*daemonv1.WithdrawMarketResponse, error) {
	return o.withdrawMarket(ctx, req)
}

func (o operatorHandler) UpdateMarketPercentageFee(
	ctx context.Context, req *daemonv1.UpdateMarketPercentageFeeRequest,
) (*daemonv1.UpdateMarketPercentageFeeResponse, error) {
	return o.updateMarketPercentageFee(ctx, req)
}

func (o operatorHandler) UpdateMarketFixedFee(
	ctx context.Context, req *daemonv1.UpdateMarketFixedFeeRequest,
) (*daemonv1.UpdateMarketFixedFeeResponse, error) {
	return o.updateMarketFixedFee(ctx, req)
}

func (o operatorHandler) UpdateMarketAssetsPrecision(
	ctx context.Context, req *daemonv1.UpdateMarketAssetsPrecisionRequest,
) (*daemonv1.UpdateMarketAssetsPrecisionResponse, error) {
	return o.updateMarketAssetsPrecision(ctx, req)
}

func (o operatorHandler) UpdateMarketPrice(
	ctx context.Context, req *daemonv1.UpdateMarketPriceRequest,
) (*daemonv1.UpdateMarketPriceResponse, error) {
	return o.updateMarketPrice(ctx, req)
}

func (o operatorHandler) UpdateMarketStrategy(
	ctx context.Context, req *daemonv1.UpdateMarketStrategyRequest,
) (*daemonv1.UpdateMarketStrategyResponse, error) {
	return o.updateMarketStrategy(ctx, req)
}

func (o operatorHandler) GetFeeFragmenterAddress(
	ctx context.Context, req *daemonv1.GetFeeFragmenterAddressRequest,
) (*daemonv1.GetFeeFragmenterAddressResponse, error) {
	return o.getFeeFragmenterAddress(ctx, req)
}

func (o operatorHandler) ListFeeFragmenterAddresses(
	ctx context.Context, req *daemonv1.ListFeeFragmenterAddressesRequest,
) (*daemonv1.ListFeeFragmenterAddressesResponse, error) {
	return o.listFeeFragmenterAddresses(ctx, req)
}

func (o operatorHandler) GetFeeFragmenterBalance(
	ctx context.Context,
	req *daemonv1.GetFeeFragmenterBalanceRequest,
) (*daemonv1.GetFeeFragmenterBalanceResponse, error) {
	return o.getFeeFragmenterBalance(ctx, req)
}

func (o operatorHandler) FeeFragmenterSplitFunds(
	req *daemonv1.FeeFragmenterSplitFundsRequest, stream daemonv1.OperatorService_FeeFragmenterSplitFundsServer,
) error {
	return o.feeFragmenterSplitFunds(req, stream)
}

func (o operatorHandler) WithdrawFeeFragmenter(
	ctx context.Context, req *daemonv1.WithdrawFeeFragmenterRequest,
) (*daemonv1.WithdrawFeeFragmenterResponse, error) {
	return o.withdrawFeeFragmenter(ctx, req)
}

func (o operatorHandler) GetMarketFragmenterAddress(
	ctx context.Context, req *daemonv1.GetMarketFragmenterAddressRequest,
) (*daemonv1.GetMarketFragmenterAddressResponse, error) {
	return o.getMarketFragmenterAddress(ctx, req)
}

func (o operatorHandler) ListMarketFragmenterAddresses(
	ctx context.Context, req *daemonv1.ListMarketFragmenterAddressesRequest,
) (*daemonv1.ListMarketFragmenterAddressesResponse, error) {
	return o.listMarketFragmenterAddresses(ctx, req)
}

func (o operatorHandler) GetMarketFragmenterBalance(
	ctx context.Context,
	req *daemonv1.GetMarketFragmenterBalanceRequest,
) (*daemonv1.GetMarketFragmenterBalanceResponse, error) {
	return o.getMarketFragmenterBalance(ctx, req)
}

func (o operatorHandler) MarketFragmenterSplitFunds(
	req *daemonv1.MarketFragmenterSplitFundsRequest, stream daemonv1.OperatorService_MarketFragmenterSplitFundsServer,
) error {
	return o.marketFragmenterSplitFunds(req, stream)
}

func (o operatorHandler) WithdrawMarketFragmenter(
	ctx context.Context, req *daemonv1.WithdrawMarketFragmenterRequest,
) (*daemonv1.WithdrawMarketFragmenterResponse, error) {
	return o.withdrawMarketFragmenter(ctx, req)
}

func (o operatorHandler) ListMarkets(
	ctx context.Context, req *daemonv1.ListMarketsRequest,
) (*daemonv1.ListMarketsResponse, error) {
	return o.listMarkets(ctx, req)
}

func (o operatorHandler) ListTrades(
	ctx context.Context, req *daemonv1.ListTradesRequest,
) (*daemonv1.ListTradesResponse, error) {
	return o.listTrades(ctx, req)
}

func (o operatorHandler) ReloadUtxos(
	ctx context.Context, rew *daemonv1.ReloadUtxosRequest,
) (*daemonv1.ReloadUtxosResponse, error) {
	if err := o.operatorSvc.ReloadUtxos(ctx); err != nil {
		return nil, err
	}
	return &daemonv1.ReloadUtxosResponse{}, nil
}

func (o operatorHandler) ListUtxos(
	ctx context.Context, req *daemonv1.ListUtxosRequest,
) (*daemonv1.ListUtxosResponse, error) {
	return o.listUtxos(ctx, req)
}

func (o operatorHandler) AddWebhook(
	ctx context.Context, req *daemonv1.AddWebhookRequest,
) (*daemonv1.AddWebhookResponse, error) {
	return o.addWebhook(ctx, req)
}

func (o operatorHandler) RemoveWebhook(
	ctx context.Context, req *daemonv1.RemoveWebhookRequest,
) (*daemonv1.RemoveWebhookResponse, error) {
	return o.removeWebhook(ctx, req)
}
func (o operatorHandler) ListWebhooks(
	ctx context.Context, req *daemonv1.ListWebhooksRequest,
) (*daemonv1.ListWebhooksResponse, error) {
	return o.listWebhooks(ctx, req)
}

func (o operatorHandler) ListDeposits(
	ctx context.Context, req *daemonv1.ListDepositsRequest,
) (*daemonv1.ListDepositsResponse, error) {
	return o.listDeposits(ctx, req)
}

func (o operatorHandler) ListWithdrawals(
	ctx context.Context, req *daemonv1.ListWithdrawalsRequest,
) (*daemonv1.ListWithdrawalsResponse, error) {
	return o.listWithdrawals(ctx, req)
}

func (o operatorHandler) GetMarketReport(
	ctx context.Context,
	req *daemonv1.GetMarketReportRequest,
) (*daemonv1.GetMarketReportResponse, error) {
	return o.getMarketReport(ctx, req)
}

func (o operatorHandler) ListProtoServices(
	ctx context.Context,
	req *daemonv1.ListProtoServicesRequest,
) (*daemonv1.ListProtoServicesResponse, error) {
	tlsConf := &tls.Config{InsecureSkipVerify: true} // nolint:gosec
	creds := credentials.NewTLS(tlsConf)
	conn, err := grpc.Dial(
		o.operatorAddress,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	allServices := make([]string, 0)
	services, err := listServices(conn)
	if err != nil {
		return nil, err
	}
	allServices = append(allServices, services...)

	if o.operatorAddress != o.tradeAddress {
		conn, err = grpc.Dial(
			o.tradeAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		services, err = listServices(conn)
		if err != nil {
			return nil, err
		}
		allServices = append(allServices, services...)
	}

	return &daemonv1.ListProtoServicesResponse{
		Services: allServices,
	}, nil
}

func listServices(conn *grpc.ClientConn) ([]string, error) {
	stub := rpb.NewServerReflectionClient(conn)
	client, err := stub.ServerReflectionInfo(context.Background())
	if err != nil {
		return nil, err
	}
	defer client.CloseSend()

	if err := client.Send(&rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_ListServices{},
	}); err != nil {
		return nil, err
	}

	resp, err := client.Recv()
	if err != nil {
		return nil, err
	}

	reflectionRpc := rpb.File_reflection_grpc_reflection_v1alpha_reflection_proto.Services().Get(0).FullName()
	services := make([]string, 0, len(resp.GetListServicesResponse().Service))
	for _, service := range resp.GetListServicesResponse().Service {
		if service.GetName() != string(reflectionRpc) {
			services = append(services, service.Name)
		}
	}

	return services, nil
}

func (o operatorHandler) getInfo(
	ctx context.Context, _ *daemonv1.GetInfoRequest,
) (*daemonv1.GetInfoResponse, error) {
	info, err := o.operatorSvc.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	accountInfo := make([]*daemonv1.AccountInfo, 0, len(info.Accounts))
	for _, a := range info.Accounts {
		accountInfo = append(accountInfo, &daemonv1.AccountInfo{
			AccountIndex:        a.Index,
			DerivationPath:      a.DerivationPath,
			Xpub:                a.Xpub,
			LastExternalDerived: a.LastExternalDerived,
			LastInternalDerived: a.LastInternalDerived,
		})
	}
	return &daemonv1.GetInfoResponse{
		RootPath:          info.RootPath,
		MasterBlindingKey: info.MasterBlindingKey,
		AccountInfo:       accountInfo,
		Network:           info.Network,
		BuildData: &daemonv1.BuildInfo{
			Version: info.BuildInfo.Version,
			Commit:  info.BuildInfo.Commit,
			Date:    info.BuildInfo.Date,
		},
		FixedBaseAsset:  info.BaseAsset,
		FixedQuoteAsset: info.QuoteAsset,
	}, nil
}

func (o operatorHandler) getFeeAddress(
	ctx context.Context, req *daemonv1.GetFeeAddressRequest,
) (*daemonv1.GetFeeAddressResponse, error) {
	info, err := o.operatorSvc.GetFeeAddress(
		ctx, int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*daemonv1.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &daemonv1.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &daemonv1.GetFeeAddressResponse{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) listFeeAddresses(
	ctx context.Context, _ *daemonv1.ListFeeAddressesRequest,
) (*daemonv1.ListFeeAddressesResponse, error) {
	info, err := o.operatorSvc.ListFeeExternalAddresses(ctx)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*daemonv1.AddressWithBlindingKey, 0, len(info))
	for _, i := range info {
		addressesAndKeys = append(addressesAndKeys, &daemonv1.AddressWithBlindingKey{
			Address:  i.Address,
			Blinding: i.BlindingKey,
		})
	}

	return &daemonv1.ListFeeAddressesResponse{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) getFeeBalance(
	ctx context.Context, req *daemonv1.GetFeeBalanceRequest,
) (*daemonv1.GetFeeBalanceResponse, error) {
	unlockedBalance, totalBalance, err := o.operatorSvc.GetFeeBalance(ctx)
	if err != nil {
		return nil, err
	}

	return &daemonv1.GetFeeBalanceResponse{
		AvailableBalance: uint64(unlockedBalance),
		TotalBalance:     uint64(totalBalance),
	}, nil
}

func (o operatorHandler) claimFeeDeposits(
	ctx context.Context, req *daemonv1.ClaimFeeDepositsRequest,
) (*daemonv1.ClaimFeeDepositsResponse, error) {
	outpoints := parseOutpoints(req.GetOutpoints())

	if err := o.operatorSvc.ClaimFeeDeposits(ctx, outpoints); err != nil {
		return nil, err
	}

	return &daemonv1.ClaimFeeDepositsResponse{}, nil
}

func (o operatorHandler) withdrawFee(
	ctx context.Context, req *daemonv1.WithdrawFeeRequest,
) (*daemonv1.WithdrawFeeResponse, error) {
	args := application.WithdrawFeeReq{
		Amount:          req.GetAmount(),
		Address:         req.GetAddress(),
		Asset:           req.GetAsset(),
		MillisatPerByte: req.GetMillisatsPerByte(),
		Password:        req.GetPassword(),
		Push:            true,
	}
	if err := args.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, txid, err := o.operatorSvc.WithdrawFeeFunds(ctx, args)
	if err != nil {
		return nil, err
	}

	return &daemonv1.WithdrawFeeResponse{Txid: hex.EncodeToString(txid)}, nil
}

func (o operatorHandler) newMarket(
	ctx context.Context, req *daemonv1.NewMarketRequest,
) (*daemonv1.NewMarketResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	basePrecision, err := parsePrecision(req.GetBaseAssetPrecision())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	quotePrecision, err := parsePrecision(req.GetQuoteAssetPrecision())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.NewMarket(ctx, market, basePrecision, quotePrecision); err != nil {
		return nil, err
	}

	return &daemonv1.NewMarketResponse{}, nil
}

func (o operatorHandler) getMarketInfo(
	ctx context.Context, req *daemonv1.GetMarketInfoRequest,
) (*daemonv1.GetMarketInfoResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := o.operatorSvc.GetMarketInfo(ctx, market)
	if err != nil {
		return nil, err
	}
	basePrice, _ := info.Price.BasePrice.BigFloat().Float64()
	quotePrice, _ := info.Price.QuotePrice.BigFloat().Float64()

	return &daemonv1.GetMarketInfoResponse{
		Info: &daemonv1.MarketInfo{
			Market: &tdexv1.Market{
				BaseAsset:  info.Market.BaseAsset,
				QuoteAsset: info.Market.QuoteAsset,
			},
			Fee: &tdexv1.Fee{
				BasisPoint: info.Fee.BasisPoint,
				Fixed: &tdexv1.Fixed{
					BaseFee:  info.Fee.FixedBaseFee,
					QuoteFee: info.Fee.FixedQuoteFee,
				},
			},
			Tradable:     info.Tradable,
			StrategyType: daemonv1.StrategyType(info.StrategyType),
			AccountIndex: info.AccountIndex,
			Price: &tdexv1.Price{
				BasePrice:  basePrice,
				QuotePrice: quotePrice,
			},
			Balance: &tdexv1.Balance{
				BaseAmount:  info.Balance.BaseAmount,
				QuoteAmount: info.Balance.QuoteAmount,
			},
			BaseAssetPrecision:  uint32(info.BaseAssetPrecision),
			QuoteAssetPrecision: uint32(info.QuoteAssetPrecision),
		},
	}, nil
}

func (o operatorHandler) getMarketAddress(
	ctx context.Context, req *daemonv1.GetMarketAddressRequest,
) (*daemonv1.GetMarketAddressResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := o.operatorSvc.GetMarketAddress(
		ctx, market, int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*daemonv1.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &daemonv1.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &daemonv1.GetMarketAddressResponse{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) listMarketAddresses(
	ctx context.Context, req *daemonv1.ListMarketAddressesRequest,
) (*daemonv1.ListMarketAddressesResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := o.operatorSvc.ListMarketExternalAddresses(ctx, market)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*daemonv1.AddressWithBlindingKey, 0, len(info))
	for _, i := range info {
		addressesAndKeys = append(addressesAndKeys, &daemonv1.AddressWithBlindingKey{
			Address:  i.Address,
			Blinding: i.BlindingKey,
		})
	}

	return &daemonv1.ListMarketAddressesResponse{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) getMarketBalance(
	ctx context.Context, req *daemonv1.GetMarketBalanceRequest,
) (*daemonv1.GetMarketBalanceResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	unlockedBalance, totalBalance, err := o.operatorSvc.GetMarketBalance(
		ctx, market,
	)
	if err != nil {
		return nil, err
	}

	return &daemonv1.GetMarketBalanceResponse{
		AvailableBalance: &tdexv1.Balance{
			BaseAmount:  unlockedBalance.BaseAmount,
			QuoteAmount: unlockedBalance.QuoteAmount,
		},
		TotalBalance: &tdexv1.Balance{
			BaseAmount:  totalBalance.BaseAmount,
			QuoteAmount: totalBalance.QuoteAmount,
		},
	}, nil
}

func (o operatorHandler) claimMarketDeposits(
	ctx context.Context,
	req *daemonv1.ClaimMarketDepositsRequest,
) (*daemonv1.ClaimMarketDepositsResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	outpoints := parseOutpoints(req.GetOutpoints())

	if err := o.operatorSvc.ClaimMarketDeposits(
		ctx, market, outpoints,
	); err != nil {
		return nil, err
	}

	return &daemonv1.ClaimMarketDepositsResponse{}, nil
}

func (o operatorHandler) openMarket(
	ctx context.Context, req *daemonv1.OpenMarketRequest,
) (*daemonv1.OpenMarketResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.OpenMarket(ctx, market); err != nil {
		return nil, err
	}

	return &daemonv1.OpenMarketResponse{}, nil
}

func (o operatorHandler) closeMarket(
	ctx context.Context, req *daemonv1.CloseMarketRequest,
) (*daemonv1.CloseMarketResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.CloseMarket(ctx, market); err != nil {
		return nil, err
	}

	return &daemonv1.CloseMarketResponse{}, nil
}

func (o operatorHandler) dropMarket(
	ctx context.Context, req *daemonv1.DropMarketRequest,
) (*daemonv1.DropMarketResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.DropMarket(ctx, market); err != nil {
		return nil, err
	}

	return &daemonv1.DropMarketResponse{}, nil
}

func (o operatorHandler) getMarketCollectedSwapFees(
	ctx context.Context, req *daemonv1.GetMarketCollectedSwapFeesRequest,
) (*daemonv1.GetMarketCollectedSwapFeesResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	page := parsePage(req.GetPage())

	report, err := o.operatorSvc.GetMarketCollectedFee(ctx, market, page)
	if err != nil {
		return nil, err
	}

	collectedFees := make([]*daemonv1.FeeInfo, 0)
	for _, fee := range report.CollectedFees {
		marketPrice, _ := fee.MarketPrice.BigFloat().Float32()
		collectedFees = append(collectedFees, &daemonv1.FeeInfo{
			TradeId:             fee.TradeID,
			BasisPoint:          fee.BasisPoint,
			Asset:               fee.Asset,
			PercentageFeeAmount: fee.PercentageFeeAmount,
			FixedFeeAmount:      fee.FixedFeeAmount,
			MarketPrice:         marketPrice,
		})
	}

	return &daemonv1.GetMarketCollectedSwapFeesResponse{
		CollectedFees:              collectedFees,
		TotalCollectedFeesPerAsset: report.TotalCollectedFeesPerAsset,
	}, nil
}

func (o operatorHandler) withdrawMarket(
	ctx context.Context, req *daemonv1.WithdrawMarketRequest,
) (*daemonv1.WithdrawMarketResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	balanceToWithdraw := parseBalance(req.GetBalanceToWithdraw())

	args := application.WithdrawMarketReq{
		Market:            market,
		BalanceToWithdraw: balanceToWithdraw,
		MillisatPerByte:   req.GetMillisatsPerByte(),
		Address:           req.GetAddress(),
		Password:          req.GetPassword(),
		Push:              true,
	}
	if err := args.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, txid, err := o.operatorSvc.WithdrawMarketFunds(ctx, args)
	if err != nil {
		return nil, err
	}

	return &daemonv1.WithdrawMarketResponse{Txid: hex.EncodeToString(txid)}, nil
}

func (o operatorHandler) updateMarketPercentageFee(
	ctx context.Context, req *daemonv1.UpdateMarketPercentageFeeRequest,
) (*daemonv1.UpdateMarketPercentageFeeResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mwf := application.MarketWithFee{
		Market: market,
		Fee: application.Fee{
			BasisPoint: req.GetBasisPoint(),
		},
	}
	result, err := o.operatorSvc.UpdateMarketPercentageFee(ctx, mwf)
	if err != nil {
		return nil, err
	}

	return &daemonv1.UpdateMarketPercentageFeeResponse{
		MarketWithFee: &tdexv1.MarketWithFee{
			Market: &tdexv1.Market{
				BaseAsset:  result.BaseAsset,
				QuoteAsset: result.QuoteAsset,
			},
			Fee: &tdexv1.Fee{
				BasisPoint: result.Fee.BasisPoint,
				Fixed: &tdexv1.Fixed{
					BaseFee:  result.FixedBaseFee,
					QuoteFee: result.FixedQuoteFee,
				},
			},
		},
	}, nil
}

func (o operatorHandler) updateMarketFixedFee(
	ctx context.Context, req *daemonv1.UpdateMarketFixedFeeRequest,
) (*daemonv1.UpdateMarketFixedFeeResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	fee := parseFixedFee(req.GetFixed())

	mwf := application.MarketWithFee{market, fee}
	result, err := o.operatorSvc.UpdateMarketFixedFee(ctx, mwf)
	if err != nil {
		return nil, err
	}

	return &daemonv1.UpdateMarketFixedFeeResponse{
		MarketWithFee: &tdexv1.MarketWithFee{
			Market: &tdexv1.Market{
				BaseAsset:  result.BaseAsset,
				QuoteAsset: result.QuoteAsset,
			},
			Fee: &tdexv1.Fee{
				BasisPoint: result.BasisPoint,
				Fixed: &tdexv1.Fixed{
					BaseFee:  result.FixedBaseFee,
					QuoteFee: result.FixedQuoteFee,
				},
			},
		},
	}, nil
}

func (o operatorHandler) updateMarketAssetsPrecision(
	ctx context.Context, req *daemonv1.UpdateMarketAssetsPrecisionRequest,
) (*daemonv1.UpdateMarketAssetsPrecisionResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.UpdateMarketAssetsPrecision(
		ctx, market, int(req.GetBaseAssetPrecision()), int(req.GetQuoteAssetPrecision()),
	); err != nil {
		return nil, err
	}

	return &daemonv1.UpdateMarketAssetsPrecisionResponse{}, nil
}

func (o operatorHandler) updateMarketPrice(
	ctx context.Context, req *daemonv1.UpdateMarketPriceRequest,
) (*daemonv1.UpdateMarketPriceResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	price, err := parsePrice(req.GetPrice())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mwp := application.MarketWithPrice{market, price}
	if err := o.operatorSvc.UpdateMarketPrice(ctx, mwp); err != nil {
		return nil, err
	}

	return &daemonv1.UpdateMarketPriceResponse{}, nil
}

func (o operatorHandler) updateMarketStrategy(
	ctx context.Context, req *daemonv1.UpdateMarketStrategyRequest,
) (*daemonv1.UpdateMarketStrategyResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	strategyType, err := parseStrategy(req.GetStrategyType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ms := application.MarketStrategy{market, strategyType}
	if err := o.operatorSvc.UpdateMarketStrategy(ctx, ms); err != nil {
		return nil, err
	}

	return &daemonv1.UpdateMarketStrategyResponse{}, nil
}

func (o operatorHandler) getFeeFragmenterAddress(
	ctx context.Context, req *daemonv1.GetFeeFragmenterAddressRequest,
) (*daemonv1.GetFeeFragmenterAddressResponse, error) {
	info, err := o.operatorSvc.GetFeeFragmenterAddress(
		ctx, int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*daemonv1.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &daemonv1.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &daemonv1.GetFeeFragmenterAddressResponse{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) listFeeFragmenterAddresses(
	ctx context.Context, req *daemonv1.ListFeeFragmenterAddressesRequest,
) (*daemonv1.ListFeeFragmenterAddressesResponse, error) {
	info, err := o.operatorSvc.ListFeeFragmenterExternalAddresses(ctx)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*daemonv1.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &daemonv1.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &daemonv1.ListFeeFragmenterAddressesResponse{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) getFeeFragmenterBalance(
	ctx context.Context, req *daemonv1.GetFeeFragmenterBalanceRequest,
) (*daemonv1.GetFeeFragmenterBalanceResponse, error) {
	info, err := o.operatorSvc.GetFeeFragmenterBalance(ctx)
	if err != nil {
		return nil, err
	}

	balance := make(map[string]*daemonv1.BalanceInfo)
	for a, b := range info {
		balance[a] = &daemonv1.BalanceInfo{
			ConfirmedBalance:   b.ConfirmedBalance,
			UnconfirmedBalance: b.UnconfirmedBalance,
			TotalBalance:       b.TotalBalance,
		}
	}

	return &daemonv1.GetFeeFragmenterBalanceResponse{
		Balance: balance,
	}, nil
}

func (o operatorHandler) feeFragmenterSplitFunds(
	req *daemonv1.FeeFragmenterSplitFundsRequest,
	stream daemonv1.OperatorService_FeeFragmenterSplitFundsServer,
) error {
	chReplies := make(chan application.FragmenterSplitFundsReply)
	go o.operatorSvc.FeeFragmenterSplitFunds(
		stream.Context(), req.GetMaxFragments(), req.GetMillisatsPerByte(),
		chReplies,
	)

	for reply := range chReplies {
		if reply.Err != nil {
			return reply.Err
		}

		if err := stream.Send(&daemonv1.FeeFragmenterSplitFundsResponse{
			Message: reply.Msg,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (o operatorHandler) withdrawFeeFragmenter(
	ctx context.Context, req *daemonv1.WithdrawFeeFragmenterRequest,
) (*daemonv1.WithdrawFeeFragmenterResponse, error) {
	txid, err := o.operatorSvc.WithdrawFeeFragmenterFunds(
		ctx, req.GetAddress(), req.GetMillisatsPerByte(), req.GetPassword(),
	)
	if err != nil {
		return nil, err
	}

	return &daemonv1.WithdrawFeeFragmenterResponse{
		Txid: txid,
	}, nil
}

func (o operatorHandler) getMarketFragmenterAddress(
	ctx context.Context, req *daemonv1.GetMarketFragmenterAddressRequest,
) (*daemonv1.GetMarketFragmenterAddressResponse, error) {
	info, err := o.operatorSvc.GetMarketFragmenterAddress(
		ctx, int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*daemonv1.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &daemonv1.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &daemonv1.GetMarketFragmenterAddressResponse{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) listMarketFragmenterAddresses(
	ctx context.Context, req *daemonv1.ListMarketFragmenterAddressesRequest,
) (*daemonv1.ListMarketFragmenterAddressesResponse, error) {
	info, err := o.operatorSvc.ListMarketFragmenterExternalAddresses(ctx)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*daemonv1.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &daemonv1.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &daemonv1.ListMarketFragmenterAddressesResponse{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) getMarketFragmenterBalance(
	ctx context.Context, req *daemonv1.GetMarketFragmenterBalanceRequest,
) (*daemonv1.GetMarketFragmenterBalanceResponse, error) {
	info, err := o.operatorSvc.GetMarketFragmenterBalance(ctx)
	if err != nil {
		return nil, err
	}

	balance := make(map[string]*daemonv1.BalanceInfo)
	for a, b := range info {
		balance[a] = &daemonv1.BalanceInfo{
			ConfirmedBalance:   b.ConfirmedBalance,
			UnconfirmedBalance: b.UnconfirmedBalance,
			TotalBalance:       b.TotalBalance,
		}
	}

	return &daemonv1.GetMarketFragmenterBalanceResponse{
		Balance: balance,
	}, nil
}

func (o operatorHandler) marketFragmenterSplitFunds(
	req *daemonv1.MarketFragmenterSplitFundsRequest,
	stream daemonv1.OperatorService_MarketFragmenterSplitFundsServer,
) error {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	chReplies := make(chan application.FragmenterSplitFundsReply)
	go o.operatorSvc.MarketFragmenterSplitFunds(
		stream.Context(), market, req.GetMillisatsPerByte(), chReplies,
	)

	for reply := range chReplies {
		if reply.Err != nil {
			return reply.Err
		}

		if err := stream.Send(&daemonv1.MarketFragmenterSplitFundsResponse{
			Message: reply.Msg,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (o operatorHandler) withdrawMarketFragmenter(
	ctx context.Context, req *daemonv1.WithdrawMarketFragmenterRequest,
) (*daemonv1.WithdrawMarketFragmenterResponse, error) {
	txid, err := o.operatorSvc.WithdrawMarketFragmenterFunds(
		ctx, req.GetAddress(), req.GetMillisatsPerByte(), req.GetPassword(),
	)
	if err != nil {
		return nil, err
	}

	return &daemonv1.WithdrawMarketFragmenterResponse{
		Txid: txid,
	}, nil
}

func (o operatorHandler) listTrades(
	ctx context.Context, req *daemonv1.ListTradesRequest,
) (*daemonv1.ListTradesResponse, error) {
	page := parsePage(req.GetPage())

	var tradeInfo []application.TradeInfo
	var err error
	if mkt := req.GetMarket(); mkt == nil {
		tradeInfo, err = o.operatorSvc.ListTrades(ctx, page)
		if err != nil {
			return nil, err
		}
	} else {
		market, err := parseMarket(req.GetMarket())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		tradeInfo, err = o.operatorSvc.ListTradesForMarket(ctx, market, page)
		if err != nil {
			return nil, err
		}
	}

	pbTradeInfo := make([]*daemonv1.TradeInfo, 0, len(tradeInfo))
	for _, info := range tradeInfo {
		basePrice, _ := info.Price.BasePrice.Float64()
		quotePrice, _ := info.Price.QuotePrice.Float64()

		pbInfo := &daemonv1.TradeInfo{
			TradeId: info.ID,
			Status: &daemonv1.TradeStatusInfo{
				Status: daemonv1.TradeStatus(info.Status.Code),
				Failed: info.Status.Failed,
			},
			MarketWithFee: &tdexv1.MarketWithFee{
				Market: &tdexv1.Market{
					BaseAsset:  info.MarketWithFee.BaseAsset,
					QuoteAsset: info.MarketWithFee.QuoteAsset,
				},
				Fee: &tdexv1.Fee{
					BasisPoint: info.MarketWithFee.BasisPoint,
					Fixed: &tdexv1.Fixed{
						BaseFee:  info.MarketWithFee.FixedBaseFee,
						QuoteFee: info.MarketWithFee.FixedQuoteFee,
					},
				},
			},
			Price: &daemonv1.TradePrice{
				BasePrice:  basePrice,
				QuotePrice: quotePrice,
			},
			TxUrl:            info.TxURL,
			RequestTimeUnix:  info.RequestTimeUnix,
			AcceptTimeUnix:   info.AcceptTimeUnix,
			CompleteTimeUnix: info.CompleteTimeUnix,
			SettleTimeUnix:   info.SettleTimeUnix,
			ExpiryTimeUnix:   info.ExpiryTimeUnix,
		}

		swapInfoEmpty := info.SwapInfo == application.SwapInfo{}
		if !swapInfoEmpty {
			pbInfo.SwapInfo = &daemonv1.SwapInfo{
				AssetP:  info.SwapInfo.AssetP,
				AmountP: info.SwapInfo.AmountP,
				AssetR:  info.SwapInfo.AssetR,
				AmountR: info.SwapInfo.AmountR,
			}
		}

		failInfoEmpty := info.SwapFailInfo == application.SwapFailInfo{}
		if !failInfoEmpty {
			pbInfo.FailInfo = &daemonv1.SwapFailInfo{
				FailureCode:    uint32(info.SwapFailInfo.Code),
				FailureMessage: info.SwapFailInfo.Message,
			}
		}

		if tt := info.RequestTimeUnix; tt > 0 {
			pbInfo.RequestTimeUtc = time.Unix(int64(tt), 0).UTC().String()
		}
		if tt := info.AcceptTimeUnix; tt > 0 {
			pbInfo.AcceptTimeUtc = time.Unix(int64(tt), 0).UTC().String()
		}
		if tt := info.CompleteTimeUnix; tt > 0 {
			pbInfo.CompleteTimeUtc = time.Unix(int64(tt), 0).UTC().String()
		}
		if tt := info.SettleTimeUnix; tt > 0 {
			pbInfo.SettleTimeUtc = time.Unix(int64(tt), 0).UTC().String()
		}
		if tt := info.ExpiryTimeUnix; tt > 0 {
			pbInfo.ExpiryTimeUtc = time.Unix(int64(tt), 0).UTC().String()
		}

		pbTradeInfo = append(pbTradeInfo, pbInfo)
	}

	return &daemonv1.ListTradesResponse{Trades: pbTradeInfo}, nil
}

func (o operatorHandler) listMarkets(
	ctx context.Context, req *daemonv1.ListMarketsRequest,
) (*daemonv1.ListMarketsResponse, error) {
	marketInfos, err := o.operatorSvc.ListMarkets(ctx)
	if err != nil {
		return nil, err
	}

	pbMarketInfos := make([]*daemonv1.MarketInfo, 0, len(marketInfos))

	for _, marketInfo := range marketInfos {
		basePrice, _ := marketInfo.Price.BasePrice.BigFloat().Float64()
		quotePrice, _ := marketInfo.Price.QuotePrice.BigFloat().Float64()

		pbMarketInfos = append(pbMarketInfos, &daemonv1.MarketInfo{
			Market: &tdexv1.Market{
				BaseAsset:  marketInfo.Market.BaseAsset,
				QuoteAsset: marketInfo.Market.QuoteAsset,
			},
			Fee: &tdexv1.Fee{
				BasisPoint: marketInfo.Fee.BasisPoint,
				Fixed: &tdexv1.Fixed{
					BaseFee:  marketInfo.Fee.FixedBaseFee,
					QuoteFee: marketInfo.Fee.FixedQuoteFee,
				},
			},
			Tradable:     marketInfo.Tradable,
			StrategyType: daemonv1.StrategyType(marketInfo.StrategyType),
			AccountIndex: marketInfo.AccountIndex,
			Price: &tdexv1.Price{
				BasePrice:  basePrice,
				QuotePrice: quotePrice,
			},
			Balance: &tdexv1.Balance{
				BaseAmount:  marketInfo.Balance.BaseAmount,
				QuoteAmount: marketInfo.Balance.QuoteAmount,
			},
			BaseAssetPrecision:  uint32(marketInfo.BaseAssetPrecision),
			QuoteAssetPrecision: uint32(marketInfo.QuoteAssetPrecision),
		})
	}

	return &daemonv1.ListMarketsResponse{Markets: pbMarketInfos}, nil
}

func (o operatorHandler) listUtxos(
	ctx context.Context, req *daemonv1.ListUtxosRequest,
) (*daemonv1.ListUtxosResponse, error) {
	var page *application.Page
	if pg := req.GetPage(); pg != nil {
		page = &application.Page{
			Number: int(pg.PageNumber),
			Size:   int(pg.PageSize),
		}
	}
	accountIndex := int(req.GetAccountIndex())

	utxoInfo, err := o.operatorSvc.ListUtxos(ctx, accountIndex, page)
	if err != nil {
		return nil, err
	}

	unspents := toUtxoInfoList(utxoInfo.Unspents)
	spents := toUtxoInfoList(utxoInfo.Spents)
	locks := toUtxoInfoList(utxoInfo.Locks)

	return &daemonv1.ListUtxosResponse{
		Unspents: unspents,
		Spents:   spents,
		Locks:    locks,
	}, nil
}

func (o operatorHandler) addWebhook(
	ctx context.Context, req *daemonv1.AddWebhookRequest,
) (*daemonv1.AddWebhookResponse, error) {
	hook := application.Webhook{
		ActionType: int(req.GetAction()),
		Endpoint:   req.GetEndpoint(),
		Secret:     req.GetSecret(),
	}
	hookID, err := o.operatorSvc.AddWebhook(ctx, hook)
	if err != nil {
		return nil, err
	}
	return &daemonv1.AddWebhookResponse{Id: hookID}, nil
}

func (o operatorHandler) removeWebhook(
	ctx context.Context, req *daemonv1.RemoveWebhookRequest,
) (*daemonv1.RemoveWebhookResponse, error) {
	if err := o.operatorSvc.RemoveWebhook(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &daemonv1.RemoveWebhookResponse{}, nil
}

func (o operatorHandler) listWebhooks(
	ctx context.Context, req *daemonv1.ListWebhooksRequest,
) (*daemonv1.ListWebhooksResponse, error) {
	hooks, err := o.operatorSvc.ListWebhooks(ctx, int(req.GetAction()))
	if err != nil {
		return nil, err
	}
	hooksInfo := make([]*daemonv1.WebhookInfo, 0, len(hooks))
	for _, h := range hooks {
		hooksInfo = append(hooksInfo, &daemonv1.WebhookInfo{
			Id:         h.Id,
			Endpoint:   h.Endpoint,
			IsSecured:  h.IsSecured,
			ActionType: daemonv1.ActionType(h.ActionType),
		})
	}
	return &daemonv1.ListWebhooksResponse{
		WebhookInfo: hooksInfo,
	}, nil
}

func (o operatorHandler) listDeposits(
	ctx context.Context, req *daemonv1.ListDepositsRequest,
) (*daemonv1.ListDepositsResponse, error) {
	page := parsePage(req.GetPage())
	deposits, err := o.operatorSvc.ListDeposits(
		ctx, int(req.GetAccountIndex()), page,
	)
	if err != nil {
		return nil, err
	}

	depositsProto := make([]*daemonv1.Deposit, 0, len(deposits))
	for _, v := range deposits {
		dd := &daemonv1.Deposit{
			Utxo: &daemonv1.UtxoInfo{
				Outpoint: &daemonv1.Outpoint{
					Hash:  v.TxID,
					Index: int32(v.VOut),
				},
				Value: v.Value,
				Asset: v.Asset,
			},
		}
		if v.Timestamp > 0 {
			dd.TimestampUnix = v.Timestamp
			dd.TimestampUtc = time.Unix(int64(v.Timestamp), 0).UTC().String()
		}
		depositsProto = append(depositsProto, dd)
	}

	return &daemonv1.ListDepositsResponse{
		AccountIndex: req.GetAccountIndex(),
		Deposits:     depositsProto,
	}, err
}

func (o operatorHandler) listWithdrawals(
	ctx context.Context, req *daemonv1.ListWithdrawalsRequest,
) (*daemonv1.ListWithdrawalsResponse, error) {
	page := parsePage(req.GetPage())

	withdrawals, err := o.operatorSvc.ListWithdrawals(
		ctx, int(req.GetAccountIndex()), page,
	)

	withdrawalsProto := make([]*daemonv1.Withdrawal, 0, len(withdrawals))
	for _, v := range withdrawals {
		ww := &daemonv1.Withdrawal{
			TxId: v.TxID,
			Balance: &tdexv1.Balance{
				BaseAmount:  v.BaseAmount,
				QuoteAmount: v.QuoteAmount,
			},
			Address: v.Address,
		}
		if v.Timestamp > 0 {
			ww.TimestampUnix = v.Timestamp
			ww.TimestampUtc = time.Unix(int64(v.Timestamp), 0).UTC().String()
		}
		withdrawalsProto = append(withdrawalsProto, ww)
	}

	return &daemonv1.ListWithdrawalsResponse{
		AccountIndex: req.GetAccountIndex(),
		Withdrawals:  withdrawalsProto,
	}, err
}

func (o operatorHandler) getMarketReport(
	ctx context.Context, req *daemonv1.GetMarketReportRequest,
) (*daemonv1.GetMarketReportResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	timeRange, err := parseTimeRange(req.GetTimeRange())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	groupByTimeFrame := parseTimeFrame(req.GetTimeFrame())

	report, err := o.operatorSvc.GetMarketReport(ctx, market, *timeRange, groupByTimeFrame)
	if err != nil {
		return nil, err
	}

	groupedVolume := make([]*daemonv1.MarketVolume, 0, len(report.GroupedVolume))
	for _, v := range report.GroupedVolume {
		groupedVolume = append(groupedVolume, &daemonv1.MarketVolume{
			BaseVolume:  v.BaseVolume,
			QuoteVolume: v.QuoteVolume,
			StartDate:   v.StartTime.Format(time.RFC3339),
			EndDate:     v.EndTime.Format(time.RFC3339),
		})
	}

	return &daemonv1.GetMarketReportResponse{
		Report: &daemonv1.MarketReport{
			TotalCollectedFees: &daemonv1.MarketCollectedFees{
				BaseAmount:  report.CollectedFees.BaseAmount,
				QuoteAmount: report.CollectedFees.QuoteAmount,
				StartDate:   report.CollectedFees.StartTime.Format(time.RFC3339),
				EndDate:     report.CollectedFees.EndTime.Format(time.RFC3339),
			},
			TotalVolume: &daemonv1.MarketVolume{
				BaseVolume:  report.Volume.BaseVolume,
				QuoteVolume: report.Volume.QuoteVolume,
				StartDate:   report.Volume.StartTime.Format(time.RFC3339),
				EndDate:     report.Volume.EndTime.Format(time.RFC3339),
			},
			GroupedVolume: groupedVolume,
		},
	}, nil
}

func parseMarket(mkt *tdexv1.Market) (market application.Market, err error) {
	var baseAsset, quoteAsset string
	if mkt != nil {
		baseAsset, quoteAsset = mkt.GetBaseAsset(), mkt.GetQuoteAsset()
	}
	m := application.Market{baseAsset, quoteAsset}
	if err = m.Validate(); err != nil {
		return
	}

	market = m
	return
}

func parseOutpoints(outs []*daemonv1.Outpoint) []application.TxOutpoint {
	outpoints := make([]application.TxOutpoint, 0, len(outs))
	for _, v := range outs {
		outpoints = append(outpoints, application.TxOutpoint{
			Hash:  v.Hash,
			Index: int(v.Index),
		})
	}
	return outpoints
}

func parsePage(p *daemonv1.Page) *application.Page {
	if p == nil {
		return nil
	}
	return &application.Page{
		Number: int(p.PageNumber),
		Size:   int(p.PageSize),
	}
}

func parseBalance(bal *tdexv1.Balance) application.Balance {
	var baseAmount, quoteAmount uint64
	if bal != nil {
		baseAmount = bal.GetBaseAmount()
		quoteAmount = bal.GetQuoteAmount()
	}
	return application.Balance{baseAmount, quoteAmount}
}

func parseFixedFee(fee *tdexv1.Fixed) application.Fee {
	var baseFee, quoteFee int64
	if fee != nil {
		baseFee = fee.GetBaseFee()
		quoteFee = fee.GetQuoteFee()
	}
	return application.Fee{
		FixedBaseFee:  baseFee,
		FixedQuoteFee: quoteFee,
	}
}

func parsePrecision(precision uint32) (uint, error) {
	if precision > 8 {
		return 0, fmt.Errorf("asset precision must be in range [0, 8]")
	}
	return uint(precision), nil
}

func parsePrice(p *tdexv1.Price) (price application.Price, err error) {
	var basePrice, quotePrice = decimal.NewFromInt(0), decimal.NewFromInt(0)
	if p != nil {
		basePrice = decimal.NewFromFloat(p.GetBasePrice())
		quotePrice = decimal.NewFromFloat(p.GetQuotePrice())
	}
	pp := application.Price{
		BasePrice:  basePrice,
		QuotePrice: quotePrice,
	}
	if err = pp.Validate(); err != nil {
		return
	}
	price = pp
	return
}

func parseStrategy(sType daemonv1.StrategyType) (domain.StrategyType, error) {
	strategyType := domain.StrategyType(sType)
	if strategyType < domain.StrategyTypePluggable ||
		strategyType > domain.StrategyTypeUnbalanced {
		return -1, errors.New("strategy type is unknown")
	}
	return strategyType, nil
}

func toUtxoInfoList(list []application.UtxoInfo) []*daemonv1.UtxoInfo {
	res := make([]*daemonv1.UtxoInfo, 0, len(list))
	for _, u := range list {
		res = append(res, &daemonv1.UtxoInfo{
			Outpoint: &daemonv1.Outpoint{
				Hash:  u.Outpoint.Hash,
				Index: int32(u.Outpoint.Index),
			},
			Value: u.Value,
			Asset: u.Asset,
		})
	}
	return res
}

func parseTimeRange(timeRange *daemonv1.TimeRange) (*application.TimeRange, error) {
	var predefinedPeriod *application.PredefinedPeriod
	if timeRange.GetPredefinedPeriod() > daemonv1.PredefinedPeriod_PREDEFINED_PERIOD_UNSPECIFIED {
		pp := parsePredefinedPeriod(timeRange.GetPredefinedPeriod())
		predefinedPeriod = &pp
	}
	var customPeriod *application.CustomPeriod
	if timeRange.GetCustomPeriod() != nil {
		customPeriod = &application.CustomPeriod{
			StartDate: timeRange.GetCustomPeriod().GetStartDate(),
			EndDate:   timeRange.GetCustomPeriod().GetEndDate(),
		}
	}
	tr := &application.TimeRange{
		PredefinedPeriod: predefinedPeriod,
		CustomPeriod:     customPeriod,
	}
	if err := tr.Validate(); err != nil {
		return nil, err
	}
	return tr, nil
}

func parseTimeFrame(timeFrame daemonv1.TimeFrame) int {
	switch timeFrame {
	case daemonv1.TimeFrame_TIME_FRAME_HOUR:
		return 1
	case daemonv1.TimeFrame_TIME_FRAME_FOUR_HOURS:
		return 4
	case daemonv1.TimeFrame_TIME_FRAME_DAY:
		return 24
	case daemonv1.TimeFrame_TIME_FRAME_WEEK:
		return 24 * 7
	case daemonv1.TimeFrame_TIME_FRAME_MONTH:
		year, month, _ := time.Now().Date()
		numOfDaysForCurrentMont := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
		return numOfDaysForCurrentMont
	}

	return 1
}

func parsePredefinedPeriod(predefinedPeriod daemonv1.PredefinedPeriod) application.PredefinedPeriod {
	switch predefinedPeriod {
	case daemonv1.PredefinedPeriod_PREDEFINED_PERIOD_LAST_HOUR:
		return application.LastHour
	case daemonv1.PredefinedPeriod_PREDEFINED_PERIOD_LAST_DAY:
		return application.LastDay
	case daemonv1.PredefinedPeriod_PREDEFINED_PERIOD_LAST_WEEK:
		return application.LastWeek
	case daemonv1.PredefinedPeriod_PREDEFINED_PERIOD_LAST_MONTH:
		return application.LastMonth
	case daemonv1.PredefinedPeriod_PREDEFINED_PERIOD_LAST_THREE_MONTHS:
		return application.LastThreeMonths
	case daemonv1.PredefinedPeriod_PREDEFINED_PERIOD_YEAR_TO_DATE:
		return application.YearToDate
	case daemonv1.PredefinedPeriod_PREDEFINED_PERIOD_LAST_YEAR:
		return application.LastYear
	case daemonv1.PredefinedPeriod_PREDEFINED_PERIOD_ALL:
		return application.All
	}

	return application.NIL
}
