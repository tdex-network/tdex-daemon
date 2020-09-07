package operatorservice

import (
	"context"
	"errors"
	"fmt"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/constant"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

func (s *Service) DepositMarket(
	ctx context.Context,
	depositMarketReq *pb.DepositMarketRequest,
) (*pb.DepositMarketReply, error) {

	//TODO
	//generate fee account address
	feeAccountAddress := "dummy"
	//fetch fee account balance
	feeAccountBalance := s.unspentRepository.GetBalance(
		feeAccountAddress,
		config.GetString(config.BaseAssetKey),
	)

	//if fee account balance > FEE_ACCOUNT_BALANCE_LIMIT
	if feeAccountBalance < uint64(config.GetInt(config.FeeAccountBalanceThresholdKey)) {
		fmt.Println("fee account balance too low, cant deposit market")
		return nil, errors.New("fee account balance too low, " +
			"cant deposit market")
	}
	//create market
	_, latestAccountIndex, err := s.marketRepository.GetLatestMarket(context.Background())
	if err != nil {
		println("latest market")
		panic(fmt.Errorf("latest market: %w", err))
	}

	nextAccountIndex := latestAccountIndex + 1
	_, err = s.marketRepository.GetOrCreateMarket(ctx, nextAccountIndex)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	//Create new address for market
	marketAddress := "dummy"
	//Add newly created address to crawler
	s.crawlerSvc.AddObservable(crawler.Observable{
		AccountType: constant.MarketAccountStart, //TODO update
		AssetHash:   depositMarketReq.GetMarket().GetQuoteAsset(),
		Address:     marketAddress,
	})

	return &pb.DepositMarketReply{
		Address: marketAddress,
	}, nil
}
