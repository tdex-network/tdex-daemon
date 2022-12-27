package application

import (
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
)

const (
	DBBadger = "badger"
)

var (
	SupportedDBType = map[string]struct{}{
		DBBadger: {},
	}
)

type Config struct {
	DBType   string
	DBConfig interface{}

	OceanWallet         ports.WalletService
	SecurePubSub        ports.SecurePubSub
	MarketPercentageFee uint32
	FeeBalanceThreshold uint64
	TradePriceSlippage  decimal.Decimal

	repo     ports.RepoManager
	pubsub   PubSubService
	wallet   WalletService
	operator OperatorService
	trade    TradeService
}

func (c *Config) Validate() error {
	if _, err := c.walletService(); err != nil {
		return err
	}
	if _, err := c.repoManager(); err != nil {
		return err
	}
	return nil
}

func (c *Config) RepoManager() ports.RepoManager {
	svc, _ := c.repoManager()
	return svc
}

func (c *Config) PubSubService() PubSubService {
	svc, _ := c.pubsubService()
	return svc
}

func (c *Config) WalletService() WalletService {
	svc, _ := c.walletService()
	return svc
}

func (c *Config) OperatorService() OperatorService {
	svc, _ := c.operatorService()
	return svc
}

func (c *Config) TradeService() TradeService {
	svc, _ := c.tradeService()
	return svc
}

func (c *Config) repoManager() (ports.RepoManager, error) {
	if c.repo == nil {
		if c.DBType == DBBadger {
			datadir := c.DBConfig.(string)
			repoManager, err := dbbadger.NewRepoManager(datadir, log.New())
			if err != nil {
				return nil, err
			}
			c.repo = repoManager
		}
	}
	return c.repo, nil
}

func (c *Config) pubsubService() (PubSubService, error) {
	if c.pubsub == nil {
		c.pubsub = NewPubSubService(c.SecurePubSub)
	}
	return c.pubsub, nil
}

func (c *Config) walletService() (WalletService, error) {
	if c.wallet == nil {
		wallet, err := NewWalletService(c.OceanWallet)
		if err != nil {
			return nil, err
		}
		c.wallet = wallet
	}
	return c.wallet, nil
}

func (c *Config) operatorService() (OperatorService, error) {
	if c.operator == nil {
		wallet, _ := c.walletService()
		pubsub, _ := c.pubsubService()
		repo, _ := c.repoManager()
		operator, err := NewOperatorService(
			wallet, pubsub, repo, c.MarketPercentageFee, c.FeeBalanceThreshold,
		)
		if err != nil {
			return nil, err
		}
		c.operator = operator
	}
	return c.operator, nil
}

func (c *Config) tradeService() (TradeService, error) {
	if c.trade == nil {
		wallet, _ := c.walletService()
		pubsub, _ := c.pubsubService()
		repo, _ := c.repoManager()
		trade, err := NewTradeService(
			wallet, pubsub, repo, c.TradePriceSlippage,
		)
		if err != nil {
			return nil, err
		}
		c.trade = trade
	}
	return c.trade, nil
}