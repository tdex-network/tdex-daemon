package operatorservice

import (
	"context"
	"fmt"
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
) (reply *pb.DepositMarketReply, err error) {

	_, latestAccountIndex, err := s.marketRepository.GetLatestMarket(context.Background())
	if err != nil {
		log.Debug("latest market")
		panic(fmt.Errorf("latest market: %w", err))
	}

	accountIndex := latestAccountIndex + 1
	_, err = s.marketRepository.GetOrCreateMarket(ctx, accountIndex)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if err = s.vaultRepository.UpdateVault(ctx, nil, "", func(v *vault.Vault) (*vault.Vault, error) {
		addr, _, blindingKey, err := v.DeriveNextExternalAddressForAccount(accountIndex)
		if err != nil {
			return nil, err
		}

		reply = &pb.DepositMarketReply{
			Address: addr,
		}

		s.crawlerSvc.AddObservable(&crawler.AddressObservable{
			AccountType: accountIndex,
			Address:     addr,
			BlindingKey: blindingKey,
		})

		return v, nil
	}); err != nil {
		err = status.Error(codes.Internal, err.Error())
		return
	}

	return nil, nil
}
