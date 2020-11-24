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

const readOnlyTx = true

//BlockchainListener defines the needed method sto start and stop a blockchain listener
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
	// Loggers
	feeDepositLogged    bool
	feeBalanceLowLogged bool
}

// NewBlockchainListener returns a BlockchainListener with all the needed services
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
		e := event.(crawler.AddressEvent)
		unspents := unspentsFromEvent(e)
		ctx := context.Background()

		if _, err := b.dbManager.RunUnspentsTransaction(
			ctx,
			!readOnlyTx,
			func(ctx context.Context) (interface{}, error) {
				return nil, b.updateUnspentsForAddress(ctx, unspents, e.Address)
			},
		); err != nil {
			log.Warnf("trying to update unspents for address %s: %s\n", e.Address, err.Error())
			break
		}

		switch event.Type() {
		case crawler.FeeAccountDeposit:
			if _, err := b.dbManager.RunTransaction(
				ctx,
				readOnlyTx,
				func(ctx context.Context) (interface{}, error) {
					return nil, b.checkFeeAccountBalance(ctx, event)
				},
			); err != nil {
				log.Warnf("trying to check balance for fee account: %s\n", err.Error())
				break
			}

		case crawler.MarketAccountDeposit:
			if _, err := b.dbManager.RunTransaction(
				ctx,
				!readOnlyTx,
				func(ctx context.Context) (interface{}, error) {
					return nil, b.checkMarketAccountFundings(ctx, e.AccountIndex)
				},
			); err != nil {
				log.Warnf("trying to check fundings for market account %d: %s\n", e.AccountIndex, err.Error())
				break
			}
		}
	}
}

func (b *blockchainListener) checkFeeAccountBalance(ctx context.Context, event crawler.Event) error {
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
		if !b.feeBalanceLowLogged {
			log.Warn(
				"fee account balance for account index too low. Trades for markets won't be " +
					"served properly. Fund the fee account as soon as possible",
			)
			b.feeBalanceLowLogged = true
		}
		b.feeDepositLogged = false
	} else {
		if !b.feeDepositLogged {
			log.Info("fee account deposited, trades can be served")
			b.feeDepositLogged = true
		}
		b.feeBalanceLowLogged = false
	}
	return nil
}

func (b *blockchainListener) checkMarketAccountFundings(ctx context.Context, accountIndex int) error {
	market, err := b.marketRepository.GetMarketByAccount(ctx, accountIndex)
	if err != nil {
		return err
	}

	// if market is not found it means it's never been opened, therefore
	// let's notify whether the market can be safely opened, base or quote
	// asset are missing, or if the market account owns too many assets.
	if market == nil || !market.IsFunded() {
		addresses, _, err := b.vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, accountIndex)
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
			log.Warnf("no funds detected for market %d", accountIndex)
		case 1:
			asset := "base"
			for k := range unspentsAssetType {
				if k == config.GetString(config.BaseAssetKey) {
					asset = "quote"
				}
			}
			log.Warnf("%s asset is missing for market %d", asset, accountIndex)
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
				accountIndex,
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
				accountIndex,
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
