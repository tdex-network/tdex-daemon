package oceanwallet

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type service struct {
	addr string
	conn *grpc.ClientConn

	walletManager  *walletManager
	accountManager *accountManager
	txManager      *txManager
	notifyManager  *notifyManager
}

func NewService(addr string) (ports.OceanWallet, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	svc := &service{
		addr:           addr,
		conn:           conn,
		walletManager:  newWalletManager(conn),
		accountManager: newAccountManager(conn),
		txManager:      newTxManager(conn),
	}
	if _, err := svc.Wallet().Status(context.Background()); err != nil {
		return nil, err
	}
	svc.notifyManager, _ = newNotifyManager(conn)

	return svc, nil
}

func (s *service) Wallet() ports.WalletManager {
	return s.walletManager
}

func (s *service) Account() ports.AccountManager {
	return s.accountManager
}

func (s *service) Transaction() ports.TransactionManager {
	return s.txManager
}

func (s *service) Notification() ports.NotificationManager {
	return s.notifyManager
}

func (s *service) Close() {
	s.conn.Close()
}
