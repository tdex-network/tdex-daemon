package application

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

type BlockchainListener interface {
	ObserveBlockchain()
	StopObserveBlockchain()
}

type blockchainListener struct {
	unspentRepository domain.UnspentRepository
	marketRepository  domain.MarketRepository
	vaultRepository   domain.VaultRepository
	crawlerSvc        crawler.Service
	explorerSvc       explorer.Service
	dbManager         ports.DbManager
}

func NewBlockchainListener(
	unspentRepository domain.UnspentRepository,
	marketRepository domain.MarketRepository,
	vaultRepository domain.VaultRepository,
	crawlerSvc crawler.Service,
	explorerSvc explorer.Service,
	dbManager ports.DbManager,
) BlockchainListener {
	return newBlockchainListener(
		unspentRepository,
		marketRepository,
		vaultRepository,
		crawlerSvc,
		explorerSvc,
		dbManager,
	)
}

func newBlockchainListener(
	unspentRepository domain.UnspentRepository,
	marketRepository domain.MarketRepository,
	vaultRepository domain.VaultRepository,
	crawlerSvc crawler.Service,
	explorerSvc explorer.Service,
	dbManager ports.DbManager,
) *blockchainListener {
	return &blockchainListener{
		unspentRepository: unspentRepository,
		marketRepository:  marketRepository,
		vaultRepository:   vaultRepository,
		crawlerSvc:        crawlerSvc,
		explorerSvc:       explorerSvc,
		dbManager:         dbManager,
	}
}

func (b *blockchainListener) ObserveBlockchain() {
	go b.crawlerSvc.Start()
	go b.handleBlockChainEvents()
}

func (b *blockchainListener) StopObserveBlockchain() {
	b.crawlerSvc.Stop()
}

func (b *blockchainListener) handleBlockChainEvents() {
	for event := range b.crawlerSvc.GetEventChannel() {
		eventHandler := b.getHandlerForEvent(event.Type())
		if eventHandler == nil {
			log.Warnf("unkonwn event of type %s\n", event.Type())
			continue
		}

		if _, err := b.dbManager.RunTransaction(
			context.Background(),
			func(ctx context.Context) (interface{}, error) {
				return nil, eventHandler(ctx, event)
			},
		); err != nil {
			log.Warnf("trying to handle event %s: %s\n", event.Type(), err.Error())
			break
		}
	}
}

func (b *blockchainListener) getHandlerForEvent(eventType crawler.EventType) func(context.Context, crawler.Event) error {
	switch eventType {
	case crawler.FeeAccountDeposit:
		return b.handleFeeDepositEvent
	case crawler.MarketAccountDeposit:
		return b.handleMarketDepositEvent
	default:
		return nil
	}
}

func (b *blockchainListener) handleFeeDepositEvent(ctx context.Context, event crawler.Event) error {
	e := event.(crawler.AddressEvent)
	unspents := unspentsFromEvent(e)

	if err := b.updateUnspentsForAddress(ctx, unspents, e.Address); err != nil {
		return err
	}

	addresses, _, err := b.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return err
	}

	feeAccountBalance, err := b.unspentRepository.GetBalance(
		ctx,
		addresses,
		config.GetString(config.BaseAssetKey),
	)
	if err != nil {
		return err
	}

	if feeAccountBalance < uint64(config.GetInt(config.FeeAccountBalanceThresholdKey)) {
		log.Warn(
			"fee account balance too low. Trades for markets won't be " +
				"served properly. Fund the fee account as soon as possible",
		)
	}
	return nil
}

func (b *blockchainListener) handleMarketDepositEvent(ctx context.Context, event crawler.Event) error {
	e := event.(crawler.AddressEvent)
	unspents := unspentsFromEvent(e)

	if err := b.updateUnspentsForAddress(ctx, unspents, e.Address); err != nil {
		return err
	}

	market, err := b.marketRepository.GetMarketByAccount(ctx, e.AccountIndex)
	if err != nil {
		return err
	}

	// if market is not found it means it's never been opened, therefore
	// let's notify whether the market can be safely opened, base or quote
	// asset are missing, or if the market account owns too many assets.
	if market == nil || !market.IsFunded() {
		addresses, _, err := b.vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, e.AccountIndex)
		if err != nil {
			return err
		}
		unspents, err := b.unspentRepository.GetUnspentsForAddresses(ctx, addresses)
		if err != nil {
			return err
		}
		unspentsAssetType := map[string]bool{}
		for _, u := range unspents {
			unspentsAssetType[u.AssetHash] = true
		}

		switch len(unspentsAssetType) {
		case 0:
			log.Warnf("no funds detected for market %d", e.AccountIndex)
		case 1:
			asset := "base"
			for k := range unspentsAssetType {
				if k == config.GetString(config.BaseAssetKey) {
					asset = "quote"
				}
			}
			log.Warnf("%s asset is missing for market %d", asset, e.AccountIndex)
		case 2:
			var asset string
			for k := range unspentsAssetType {
				if k != config.GetString(config.BaseAssetKey) {
					asset = k
				}
			}
			log.Infof("funding market with quote asset %s", asset)

			// Prepare unspents to become outpoint for the market to run validations
			outpoints := make([]domain.OutpointWithAsset, 0, len(unspents))
			for _, u := range unspents {
				outpoints = append(outpoints, domain.OutpointWithAsset{
					Txid:  u.TxID,
					Vout:  int(u.VOut),
					Asset: u.AssetHash,
				})
			}

			// Update the market trying to funding attaching the newly found quote asset.
			if err := b.marketRepository.UpdateMarket(
				ctx,
				e.AccountIndex,
				func(m *domain.Market) (*domain.Market, error) {
					if err := m.FundMarket(outpoints); err != nil {
						return nil, err
					}

					return m, nil
				},
			); err != nil {
				log.Warn(err)
			}

		default:
			log.Warnf(
				"market with account %d funded with more than 2 different assets."+
					"It will be impossible to determine the correct quote asset "+
					"and market won't be opened. Funds must be moved away from "+
					"this account so that it owns only unspents of 2 type of assets",
				e.AccountIndex,
			)
		}
	}
	return nil
}

func (b *blockchainListener) updateUnspentsForAddress(
	ctx context.Context,
	unspents []domain.Unspent,
	address string,
) error {
	existingUnspents, err := b.unspentRepository.GetAllUnspentsForAddresses(
		ctx,
		[]string{address},
	)
	if err != nil {
		return err
	}

	// check for unspents to add to the storage
	unspentsToAdd := make([]domain.Unspent, 0)
	for _, u := range unspents {
		if index := findUnspent(existingUnspents, u); index < 0 {
			unspentsToAdd = append(unspentsToAdd, u)
		}
	}

	//update spent
	unspentsToMarkAsSpent := make([]domain.UnspentKey, 0)
	unspentsToMarkAsConfirmed := make([]domain.UnspentKey, 0)
	for _, existingUnspent := range existingUnspents {
		if index := findUnspent(unspents, existingUnspent); index < 0 {
			unspentsToMarkAsSpent = append(unspentsToMarkAsSpent, existingUnspent.Key())
		} else {
			if existingUnspent.IsConfirmed() != unspents[index].IsConfirmed() {
				unspentsToMarkAsConfirmed = append(unspentsToMarkAsConfirmed, existingUnspent.Key())
			}
		}
	}

	if len(unspentsToAdd) > 0 {
		if err := b.unspentRepository.AddUnspents(ctx, unspentsToAdd); err != nil {
			return err
		}
	}
	if len(unspentsToMarkAsSpent) > 0 {
		if err := b.unspentRepository.SpendUnspents(ctx, unspentsToMarkAsSpent); err != nil {
			return err
		}
	}
	if len(unspentsToMarkAsConfirmed) > 0 {
		if err := b.unspentRepository.ConfirmUnspents(ctx, unspentsToMarkAsConfirmed); err != nil {
			return err
		}
	}
	return nil
}

func unspentsFromEvent(event crawler.AddressEvent) []domain.Unspent {
	unspents := make([]domain.Unspent, 0, len(event.Utxos))
	for _, utxo := range event.Utxos {
		u := domain.Unspent{
			TxID:            utxo.Hash(),
			VOut:            utxo.Index(),
			Value:           utxo.Value(),
			AssetHash:       utxo.Asset(),
			ValueCommitment: utxo.ValueCommitment(),
			AssetCommitment: utxo.AssetCommitment(),
			ScriptPubKey:    utxo.Script(),
			Nonce:           utxo.Nonce(),
			RangeProof:      utxo.RangeProof(),
			SurjectionProof: utxo.SurjectionProof(),
			Confirmed:       utxo.IsConfirmed(),
			Address:         event.Address,
		}
		unspents = append(unspents, u)
	}
	return unspents
}

func findUnspent(list []domain.Unspent, unspent domain.Unspent) int {
	for i, u := range list {
		if u.IsKeyEqual(unspent.Key()) {
			return i
		}
	}
	return -1
}
