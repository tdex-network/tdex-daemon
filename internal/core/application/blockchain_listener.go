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
	UpdateUnspentsForAddress(
		ctx context.Context,
		unspents []domain.Unspent,
		address string,
	) error
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

func (b *blockchainListener) handleBlockChainEvents() {

events:
	for event := range b.crawlerSvc.GetEventChannel() {
		tx := b.startTx()
		ctx := context.WithValue(context.Background(), "utx", tx)
		switch event.Type() {
		case crawler.FeeAccountDeposit:
			e := event.(crawler.AddressEvent)
			unspents := make([]domain.Unspent, 0)
		utxoLoop:
			for _, utxo := range e.Utxos {
				isTrxConfirmed, err := b.explorerSvc.IsTransactionConfirmed(
					utxo.Hash(),
				)
				if err != nil {
					tx.Discard()
					log.Warn(err)
					continue utxoLoop
				}

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
					Address:         e.Address,
					Spent:           false,
					Locked:          false,
					LockedBy:        nil,
					Confirmed:       isTrxConfirmed,
				}
				unspents = append(unspents, u)
			}
			if err := b.UpdateUnspentsForAddress(ctx, unspents, e.Address); err != nil {
				tx.Discard()
				log.Warn(err)
				continue events
			}

			addresses, _, err := b.vaultRepository.
				GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, domain.FeeAccount)
			if err != nil {
				tx.Discard()
				log.Warn(err)
				continue events
			}

			feeAccountBalance, err := b.unspentRepository.GetBalance(
				ctx,
				addresses,
				config.GetString(config.BaseAssetKey),
			)
			if err != nil {
				tx.Discard()
				log.Warn(err)
				continue events
			}

			if feeAccountBalance < uint64(config.GetInt(config.FeeAccountBalanceThresholdKey)) {
				log.Warn(
					"fee account balance too low. Trades for markets won't be " +
						"served properly. Fund the fee account as soon as possible",
				)
				continue events
			}

		case crawler.MarketAccountDeposit:
			e := event.(crawler.AddressEvent)
			unspents := make([]domain.Unspent, 0)
			if len(e.Utxos) > 0 {
			utxoLoop1:
				for _, utxo := range e.Utxos {
					isTrxConfirmed, err := b.explorerSvc.IsTransactionConfirmed(
						utxo.Hash(),
					)
					if err != nil {
						tx.Discard()
						log.Warn(err)
						continue utxoLoop1
					}
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
						Address:         e.Address,
						Confirmed:       isTrxConfirmed,
					}
					unspents = append(unspents, u)
				}
				err := b.UpdateUnspentsForAddress(
					ctx,
					unspents,
					e.Address,
				)
				if err != nil {
					tx.Discard()
					log.Warn(err)
					continue events
				}

				market, err := b.marketRepository.GetMarketByAccount(ctx, e.AccountIndex)
				if err != nil {
					log.Warn(err)
				}

				// if market is not found it means it's never been opened, therefore
				// let's notify whether the market can be safely opened, base or quote
				// asset are missing, or if the market account owns too many assets.
				if market == nil || !market.IsFunded() {
					addresses, _, err := b.vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, e.AccountIndex)
					if err != nil {
						log.Warn(err)
					}
					unspents, err := b.unspentRepository.GetUnspentsForAddresses(ctx, addresses)
					if err != nil {
						log.Warn(err)
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
						log.Infof("market with quote asset '%s' can be opened", asset)
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
			}

		case crawler.TransactionConfirmed:
			//TODO
		}
		b.commitTx(tx)
	}
}

func (b *blockchainListener) UpdateUnspentsForAddress(
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

	//add new unspent
	unspentsToAdd := make([]domain.Unspent, 0)
	for _, newUnspent := range unspents {
		exist := false
		for _, existingUnspent := range existingUnspents {
			if newUnspent.IsKeyEqual(existingUnspent.Key()) {
				exist = true
				break
			}
		}
		if !exist {
			unspentsToAdd = append(unspentsToAdd, newUnspent)
		}
	}

	if len(unspentsToAdd) > 0 {
		if err := b.unspentRepository.AddUnspents(ctx, unspentsToAdd); err != nil {
			return err
		}
	}

	//update spent
	unspentsToMarkAsSpent := make([]domain.UnspentKey, 0)
	for _, existingUnspent := range existingUnspents {
		exist := false
		for _, newUnspent := range unspents {
			if existingUnspent.IsKeyEqual(newUnspent.Key()) {
				exist = true
				break
			}
		}
		if !existingUnspent.IsSpent() && !exist {
			unspentsToMarkAsSpent = append(unspentsToMarkAsSpent, existingUnspent.Key())
		}
	}

	if len(unspentsToMarkAsSpent) > 0 {
		if err := b.unspentRepository.SpendUnspents(ctx, unspentsToMarkAsSpent); err != nil {
			return err
		}
	}
	return nil
}

func (b *blockchainListener) startTx() ports.Transaction {
	return b.dbManager.NewUnspentsTransaction()
}

func (b *blockchainListener) commitTx(tx ports.Transaction) {
	err := tx.Commit()
	if err != nil {
		log.Error(err)
	}
}
