package operatorservice

import (
	"context"
	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

func (s *Service) DepositMarket(
	ctx context.Context,
	depositMarketReq *pb.DepositMarketRequest,
) (reply *pb.DepositMarketReply, errResp error) {

	var accountIndex int
	if depositMarketReq.GetMarket().GetQuoteAsset() != "" {
		log.Debug("existing market")
		_, a, err := s.marketRepository.GetMarketByAsset(
			ctx,
			depositMarketReq.GetMarket().GetQuoteAsset(),
		)
		if err != nil {
			log.Error(err)
			errResp = err
			return
		}
		accountIndex = a
	}

	if accountIndex == 0 {
		log.Debug("new market")
		_, latestAccountIndex, err := s.marketRepository.GetLatestMarket(context.Background())
		if err != nil {
			log.Error(err)
			errResp = err
			return
		}

		accountIndex = latestAccountIndex + 1
		_, err = s.marketRepository.GetOrCreateMarket(ctx, accountIndex)
		if err != nil {
			log.Error(err)
			errResp = err
			return
		}
	}

	if err := s.vaultRepository.UpdateVault(ctx, nil, "",
		func(v *vault.Vault) (*vault.Vault, error) {
			addr, _, blindingKey, err := v.DeriveNextExternalAddressForAccount(accountIndex)
			if err != nil {
				return nil, err
			}

			reply = &pb.DepositMarketReply{
				Address: addr,
			}

			s.crawlerSvc.AddObservable(&crawler.AddressObservable{
				AccountIndex: accountIndex,
				Address:      addr,
				BlindingKey:  blindingKey,
			})

			return v, nil
		}); err != nil {
		log.Error(err)
		errResp = status.Error(codes.Internal, err.Error())
		return
	}

	return
}
