package application

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"testing"
)

func newTestOperator() (OperatorService, context.Context) {
	dbManager, err := dbbadger.NewDbManager("testoperator", nil)
	if err != nil {
		panic(err)
	}

	explorerSvc := explorer.NewService("localhost:3001")
	operatorService := NewOperatorService(
		dbbadger.NewMarketRepositoryImpl(dbManager),
		dbbadger.NewVaultRepositoryImpl(dbManager),
		dbbadger.NewTradeRepositoryImpl(dbManager),
		dbbadger.NewUnspentRepositoryImpl(dbManager),
		explorerSvc,
		crawler.NewService(explorerSvc, []crawler.Observable{}, func(err error) {}),
	)

	tx := dbManager.NewTransaction()
	ctx := context.WithValue(context.Background(), "tx", tx)

	return operatorService, ctx
}

func TestListMarket(t *testing.T) {
	t.Run("ListMaker test", func(t *testing.T) {
		operatorService, ctx := newTestOperator()

		listMarketReply, err := operatorService.ListMarket(ctx, ListMarketRequest{})
		fmt.Println(listMarketReply)
		assert.Equal(t, err, nil)
	})
}
