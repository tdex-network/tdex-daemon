package application

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/circuitbreaker"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/transaction"
)

var (
	DefaultFeeFragmenterFragments = uint32(50)
	DefaultFeeFragmenterAmount    = uint64(5000)
	FragmentationMap              = map[int]int{
		1: 30,
		2: 15,
		3: 10,
		5: 2,
	}
)

// OperatorService defines the methods of the application layer for the operator service.
type OperatorService interface {
	GetInfo(ctx context.Context) (*HDWalletInfo, error)
	// Fee account
	GetFeeAddress(
		ctx context.Context, numOfAddresses int,
	) ([]AddressAndBlindingKey, error)
	ListFeeExternalAddresses(
		ctx context.Context,
	) ([]AddressAndBlindingKey, error)
	GetFeeBalance(ctx context.Context) (int64, int64, error)
	ClaimFeeDeposits(ctx context.Context, outpoints []TxOutpoint) error
	WithdrawFeeFunds(
		ctx context.Context, req WithdrawFeeReq,
	) ([]byte, []byte, error)
	// Market account
	NewMarket(ctx context.Context, market Market) error
	GetMarketAddress(
		ctx context.Context, market Market, numOfAddresses int,
	) ([]AddressAndBlindingKey, error)
	ListMarketExternalAddresses(
		ctx context.Context, req Market,
	) ([]AddressAndBlindingKey, error)
	GetMarketBalance(ctx context.Context, market Market) (*Balance, *Balance, error)
	ClaimMarketDeposits(
		ctx context.Context, market Market, outpoints []TxOutpoint,
	) error
	OpenMarket(ctx context.Context, market Market) error
	CloseMarket(ctx context.Context, market Market) error
	DropMarket(ctx context.Context, market Market) error
	GetMarketCollectedFee(
		ctx context.Context, market Market, page *Page,
	) (*ReportMarketFee, error)
	WithdrawMarketFunds(
		ctx context.Context, req WithdrawMarketReq,
	) ([]byte, []byte, error)
	UpdateMarketPercentageFee(
		ctx context.Context, req MarketWithFee,
	) (*MarketWithFee, error)
	UpdateMarketFixedFee(
		ctx context.Context, req MarketWithFee,
	) (*MarketWithFee, error)
	UpdateMarketPrice(ctx context.Context, req MarketWithPrice) error
	UpdateMarketStrategy(ctx context.Context, req MarketStrategy) error
	// Fee Fragmenter account
	GetFeeFragmenterAddress(
		ctx context.Context, numOfAddresses int,
	) ([]AddressAndBlindingKey, error)
	ListFeeFragmenterExternalAddresses(
		ctx context.Context,
	) ([]AddressAndBlindingKey, error)
	GetFeeFragmenterBalance(ctx context.Context) (map[string]BalanceInfo, error)
	FeeFragmenterSplitFunds(
		ctx context.Context, maxFragments uint32, millisatsPerByte uint64,
		chRes chan FragmenterSplitFundsReply,
	)
	WithdrawFeeFragmenterFunds(
		ctx context.Context, address string, millisatsPerByte uint64,
	) (string, error)
	// Market fragmenter account
	GetMarketFragmenterAddress(
		ctx context.Context, numOfAddresses int,
	) ([]AddressAndBlindingKey, error)
	ListMarketFragmenterExternalAddresses(
		ctx context.Context,
	) ([]AddressAndBlindingKey, error)
	GetMarketFragmenterBalance(
		ctx context.Context,
	) (map[string]BalanceInfo, error)
	MarketFragmenterSplitFunds(
		ctx context.Context, market Market, millisatsPerByte uint64,
		chRes chan FragmenterSplitFundsReply,
	)
	WithdrawMarketFragmenterFunds(
		ctx context.Context, address string, millisatsPerByte uint64,
	) (string, error)
	// List methods
	ListMarkets(ctx context.Context) ([]MarketInfo, error)
	ListTrades(ctx context.Context, page *Page) ([]TradeInfo, error)
	ListTradesForMarket(
		ctx context.Context, market Market, page *Page,
	) ([]TradeInfo, error)
	ListUtxos(
		ctx context.Context, accountIndex int, page *Page,
	) (*UtxoInfoList, error)
	ListDeposits(
		ctx context.Context, accountIndex int, page *Page,
	) (Deposits, error)
	ListWithdrawals(
		ctx context.Context, accountIndex int, page *Page,
	) (Withdrawals, error)
	// Reload utxo set
	ReloadUtxos(ctx context.Context) error
	// Webhook
	AddWebhook(ctx context.Context, hook Webhook) (string, error)
	RemoveWebhook(ctx context.Context, id string) error
	ListWebhooks(ctx context.Context, actionType int) ([]WebhookInfo, error)
}

type operatorService struct {
	repoManager                ports.RepoManager
	explorerSvc                explorer.Service
	blockchainListener         BlockchainListener
	marketBaseAsset            string
	marketQuoteAsset           string
	marketFee                  int64
	network                    *network.Network
	feeAccountBalanceThreshold uint64

	fragmenterLock *sync.RWMutex
}

// NewOperatorService is a constructor function for OperatorService.
func NewOperatorService(
	repoManager ports.RepoManager,
	explorerSvc explorer.Service,
	bcListener BlockchainListener,
	marketBaseAsset, marketQuoteAsset string,
	marketFee int64,
	net *network.Network,
	feeAccountBalanceThreshold uint64,
) OperatorService {
	return &operatorService{
		repoManager:                repoManager,
		explorerSvc:                explorerSvc,
		blockchainListener:         bcListener,
		marketBaseAsset:            marketBaseAsset,
		marketQuoteAsset:           marketQuoteAsset,
		marketFee:                  marketFee,
		network:                    net,
		feeAccountBalanceThreshold: feeAccountBalanceThreshold,
		fragmenterLock:             &sync.RWMutex{},
	}
}

func (o *operatorService) GetInfo(ctx context.Context) (*HDWalletInfo, error) {
	vault, err := o.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		return nil, err
	}

	mnemonic, err := vault.GetMnemonicSafe()
	if err != nil {
		return nil, err
	}

	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	})
	if err != nil {
		return nil, err
	}

	rootPath := wallet.DefaultBaseDerivationPath
	masterBlindingKey, err := w.MasterBlindingKey()
	if err != nil {
		return nil, err
	}

	accountInfo := make([]AccountInfo, 0, len(vault.Accounts))
	for _, a := range vault.Accounts {
		accountIndex := uint32(a.AccountIndex)
		lastExternalDerived := uint32(a.LastExternalIndex)
		lastInternalDerived := uint32(a.LastInternalIndex)
		derivationPath := fmt.Sprintf("%s/%d'", rootPath.String(), a.AccountIndex)
		xpub, err := w.ExtendedPublicKey(wallet.ExtendedKeyOpts{
			Account: accountIndex,
		})
		if err != nil {
			return nil, err
		}
		accountInfo = append(accountInfo, AccountInfo{
			Index:               accountIndex,
			DerivationPath:      derivationPath,
			Xpub:                xpub,
			LastExternalDerived: lastExternalDerived,
			LastInternalDerived: lastInternalDerived,
		})
	}

	sort.SliceStable(accountInfo, func(i, j int) bool {
		return accountInfo[i].Index < accountInfo[j].Index
	})

	return &HDWalletInfo{
		RootPath:          rootPath.String(),
		MasterBlindingKey: masterBlindingKey,
		Accounts:          accountInfo,
	}, nil
}

func (o *operatorService) GetFeeAddress(
	ctx context.Context, numOfAddresses int,
) ([]AddressAndBlindingKey, error) {
	return o.generateAddressesForAccount(ctx, domain.FeeAccount, numOfAddresses)
}

func (o *operatorService) ListFeeExternalAddresses(
	ctx context.Context,
) ([]AddressAndBlindingKey, error) {
	return o.listExternalAddressesForAccount(ctx, domain.FeeAccount)
}

func (o *operatorService) GetFeeBalance(ctx context.Context) (int64, int64, error) {
	unlockedBalance, err := getUnlockedBalanceForFee(
		o.repoManager, ctx, o.network.AssetID,
	)
	if err != nil {
		return -1, -1, err
	}
	totalBalance, err := getBalanceForFee(o.repoManager, ctx, o.network.AssetID, false)
	if err != nil {
		return -1, -1, err
	}
	return int64(unlockedBalance), int64(totalBalance), nil
}

// ClaimFeeDeposit adds unspents to the Fee Account
func (o *operatorService) ClaimFeeDeposits(
	ctx context.Context, outpoints []TxOutpoint,
) error {
	accountInfo, err := o.repoManager.VaultRepository().
		GetAllDerivedExternalAddressesInfoForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return err
	}

	return o.claimDeposit(ctx, accountInfo, outpoints, nil)
}

func (o *operatorService) WithdrawFeeFunds(
	ctx context.Context, req WithdrawFeeReq,
) ([]byte, []byte, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	lbtcAsset := o.network.AssetID
	asset := lbtcAsset
	if req.Asset != "" {
		asset = req.Asset
	}
	balance, err := getUnlockedBalanceForFee(o.repoManager, ctx, asset)
	if err != nil {
		return nil, nil, err
	}
	if req.Amount > balance {
		return nil, nil, ErrWithdrawAmountTooBig
	}

	vault, err := o.repoManager.VaultRepository().GetOrCreateVault(
		ctx, nil, "", nil,
	)
	if err != nil {
		return nil, nil, err
	}

	mnemonic, err := vault.GetMnemonicSafe()
	if err != nil {
		return nil, nil, err
	}
	feeAccount, err := vault.AccountByIndex(domain.FeeAccount)
	if err != nil {
		return nil, nil, err
	}

	outs := []TxOut{
		{asset, int64(req.Amount), req.Address},
	}
	outputs, outputsBlindingKeys, err := parseRequestOutputs(outs)
	if err != nil {
		return nil, nil, err
	}

	unspents, err := o.getAllUnspentsForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return nil, nil, err
	}

	feeInfo, err := vault.DeriveNextInternalAddressForAccount(domain.FeeAccount)
	if err != nil {
		return nil, nil, err
	}
	feeChangePathByAsset := map[string]string{
		asset: feeAccount.DerivationPathByScript[feeInfo.Script],
	}

	if asset != lbtcAsset {
		feeInfo, err := vault.DeriveNextInternalAddressForAccount(domain.FeeAccount)
		if err != nil {
			return nil, nil, err
		}
		feeChangePathByAsset = map[string]string{
			lbtcAsset: feeAccount.DerivationPathByScript[feeInfo.Script],
		}
	}

	txHex, err := sendToManyWithoutFeeTopup(sendToManyWithoutFeeTopupOpts{
		mnemonic:            mnemonic,
		unspents:            unspents,
		outputs:             outputs,
		outputsBlindingKeys: outputsBlindingKeys,
		changePathsByAsset:  feeChangePathByAsset,
		inputPathsByScript:  feeAccount.DerivationPathByScript,
		milliSatPerByte:     int(req.MillisatPerByte),
		network:             o.network,
	})
	if err != nil {
		return nil, nil, err
	}

	var txid string
	if req.Push {
		cb := circuitbreaker.NewCircuitBreaker()
		iTxid, err := cb.Execute(func() (interface{}, error) {
			return o.explorerSvc.BroadcastTransaction(txHex)
		})
		if err != nil {
			return nil, nil, err
		}
		txid = iTxid.(string)
		log.Debugf("withdrawal tx broadcasted with id: %s", txid)
	}

	if err := o.repoManager.VaultRepository().UpdateVault(
		ctx,
		func(_ *domain.Vault) (*domain.Vault, error) {
			return vault, nil
		},
	); err != nil {
		return nil, nil, err
	}

	go extractUnspentsFromTxAndUpdateUtxoSet(
		o.repoManager.UnspentRepository(),
		o.repoManager.VaultRepository(),
		o.network,
		txHex,
		domain.FeeAccount,
	)

	// Start watching tx to confirm new unspents once the tx is in blockchain.
	go o.blockchainListener.StartObserveTx(txid, Market{})

	// Publish message for topic AccountWithdraw to pubsub service.
	go func() {
		if err := publishFeeWithdrawTopic(
			o.blockchainListener.PubSubService(),
			balance, req.Amount, req.Address, txid, asset,
		); err != nil {
			log.Warn(err)
		}
	}()

	go func() {
		count, err := o.repoManager.WithdrawalRepository().AddWithdrawals(
			ctx,
			[]domain.Withdrawal{
				{
					TxID:            txid,
					AccountIndex:    domain.FeeAccount,
					BaseAmount:      req.Amount,
					MillisatPerByte: int64(req.MillisatPerByte),
					Address:         req.Address,
				},
			},
		)
		if err != nil {
			log.WithError(err).Warn("an error occured while storing withdrawal info")
			return
		}
		log.Debugf("added %d withdrawals", count)
	}()

	rawTx, _ := hex.DecodeString(txHex)
	rawTxid, _ := hex.DecodeString(txid)
	return rawTx, rawTxid, nil
}

func (o *operatorService) NewMarket(ctx context.Context, mkt Market) error {
	if err := mkt.Validate(); err != nil {
		return err
	}
	if len(o.marketBaseAsset) > 0 && mkt.BaseAsset != o.marketBaseAsset {
		return ErrMarketInvalidBaseAsset
	}
	if len(o.marketQuoteAsset) > 0 && mkt.QuoteAsset != o.marketQuoteAsset {
		return ErrMarketInvalidQuoteAsset
	}

	_, existingAccountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, mkt.BaseAsset, mkt.QuoteAsset,
	)
	if err != nil {
		return err
	}
	if existingAccountIndex >= 0 {
		return ErrMarketAlreadyExist
	}

	vault, err := o.repoManager.VaultRepository().GetOrCreateVault(
		ctx, nil, "", nil,
	)
	if err != nil {
		return err
	}

	_, latestAccountIndex, err := o.repoManager.MarketRepository().
		GetLatestMarket(ctx)
	if err != nil {
		return err
	}

	accountIndex := latestAccountIndex + 1
	newMarket, err := domain.NewMarket(
		accountIndex, mkt.BaseAsset, mkt.QuoteAsset, o.marketFee,
	)
	if err != nil {
		return err
	}
	vault.InitAccount(accountIndex)

	_, err = o.repoManager.RunTransaction(
		ctx, false, func(ctx context.Context) (interface{}, error) {
			if _, err := o.repoManager.MarketRepository().GetOrCreateMarket(
				ctx, newMarket,
			); err != nil {
				return nil, err
			}
			if err := o.repoManager.VaultRepository().UpdateVault(
				ctx,
				func(_ *domain.Vault) (*domain.Vault, error) {
					return vault, nil
				},
			); err != nil {
				return nil, err
			}
			return nil, nil
		},
	)

	return err
}

func (o *operatorService) GetMarketAddress(
	ctx context.Context, mkt Market, numOfAddresses int,
) ([]AddressAndBlindingKey, error) {
	if err := mkt.Validate(); err != nil {
		return nil, err
	}

	_, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, mkt.BaseAsset, mkt.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	return o.generateAddressesForAccount(ctx, accountIndex, numOfAddresses)
}

func (o *operatorService) ListMarketExternalAddresses(
	ctx context.Context, mkt Market,
) ([]AddressAndBlindingKey, error) {
	if err := mkt.Validate(); err != nil {
		return nil, err
	}

	_, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, mkt.BaseAsset, mkt.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	return o.listExternalAddressesForAccount(ctx, accountIndex)
}

func (o *operatorService) GetMarketBalance(
	ctx context.Context, mkt Market,
) (*Balance, *Balance, error) {
	if err := mkt.Validate(); err != nil {
		return nil, nil, err
	}

	market, _, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, mkt.BaseAsset, mkt.QuoteAsset,
	)
	if err != nil {
		return nil, nil, err
	}
	if market == nil {
		return nil, nil, ErrMarketNotExist
	}

	unlockedBalance, err := getUnlockedBalanceForMarket(o.repoManager, ctx, market)
	if err != nil {
		return nil, nil, err
	}

	totalBalance, err := getBalanceForMarket(o.repoManager, ctx, market, false)
	if err != nil {
		return nil, nil, err
	}

	return unlockedBalance, totalBalance, nil
}

// ClaimMarketDeposit method add unspents to the market
func (o *operatorService) ClaimMarketDeposits(
	ctx context.Context, mkt Market, outpoints []TxOutpoint,
) error {
	if err := mkt.Validate(); err != nil {
		return err
	}

	market, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, mkt.BaseAsset, mkt.QuoteAsset,
	)
	if err != nil {
		return err
	}
	if accountIndex < 0 {
		return ErrMarketNotFunded
	}

	info, err := o.repoManager.VaultRepository().GetAllDerivedExternalAddressesInfoForAccount(
		ctx, accountIndex,
	)
	if err != nil {
		return err
	}

	return o.claimDeposit(ctx, info, outpoints, market)
}

func (o *operatorService) OpenMarket(ctx context.Context, mkt Market) error {
	if err := mkt.Validate(); err != nil {
		return err
	}

	// Check if some addresses of the fee account have been derived already
	if _, err := o.repoManager.VaultRepository().GetAllDerivedExternalAddressesInfoForAccount(
		ctx, domain.FeeAccount,
	); err != nil {
		if err == domain.ErrVaultAccountNotFound {
			return ErrFeeAccountNotFunded
		}
		return err
	}

	// Check if market exists
	_, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, mkt.BaseAsset, mkt.QuoteAsset,
	)
	if err != nil {
		return err
	}
	if accountIndex < 0 {
		return ErrMarketNotExist
	}

	// Open the market
	return o.repoManager.MarketRepository().UpdateMarket(
		ctx, accountIndex, func(m *domain.Market) (*domain.Market, error) {
			if err := m.MakeTradable(); err != nil {
				return nil, err
			}
			return m, nil
		},
	)
}

func (o *operatorService) CloseMarket(ctx context.Context, mkt Market) error {
	if err := mkt.Validate(); err != nil {
		return err
	}

	_, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, mkt.BaseAsset, mkt.QuoteAsset,
	)
	if err != nil {
		return err
	}
	if accountIndex < 0 {
		return ErrMarketNotExist
	}

	return o.repoManager.MarketRepository().UpdateMarket(
		ctx, accountIndex, func(m *domain.Market) (*domain.Market, error) {
			if err := m.MakeNotTradable(); err != nil {
				return nil, err
			}
			return m, nil
		})
}

func (o *operatorService) DropMarket(ctx context.Context, market Market) error {
	if err := market.Validate(); err != nil {
		return err
	}

	_, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.BaseAsset, market.QuoteAsset,
	)
	if err != nil {
		return err
	}
	if accountIndex < 0 {
		return ErrMarketNotExist
	}

	info, err := o.repoManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(
		ctx, accountIndex,
	)
	if err != nil {
		return err
	}

	unspents, err := o.repoManager.UnspentRepository().GetAllUnspentsForAddresses(ctx, info.Addresses())
	if err != nil {
		return err
	}

	// If the account owns either some unlocked unspents it cannot be dropped
	hasUnlockedUnspents := false
	for _, u := range unspents {
		if !u.Spent && !u.IsLocked() {
			hasUnlockedUnspents = true
			break
		}
	}
	if hasUnlockedUnspents {
		return ErrMarketNonZeroBalance
	}

	_, err = o.repoManager.RunTransaction(
		ctx, false, func(ctx context.Context) (interface{}, error) {
			if err := o.repoManager.MarketRepository().DeleteMarket(
				ctx, accountIndex,
			); err != nil {
				return nil, err
			}

			if err := o.repoManager.VaultRepository().UpdateVault(
				ctx, func(v *domain.Vault) (*domain.Vault, error) {
					v.InitAccount(accountIndex)
					return v, nil
				},
			); err != nil {
				return nil, err
			}
			return nil, nil
		},
	)

	return err
}

func (o *operatorService) GetMarketCollectedFee(
	ctx context.Context, mkt Market, page *Page,
) (*ReportMarketFee, error) {
	m, _, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, mkt.BaseAsset, mkt.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, ErrMarketNotExist
	}

	var trades []*domain.Trade
	if page == nil {
		trades, err = o.repoManager.TradeRepository().GetCompletedTradesByMarket(
			ctx, mkt.QuoteAsset,
		)
	} else {
		pg := page.ToDomain()
		trades, err = o.repoManager.TradeRepository().GetCompletedTradesByMarketAndPage(
			ctx, mkt.QuoteAsset, pg,
		)
	}
	if err != nil {
		return nil, err
	}

	// sort trades by timestamp like done in ListTrades
	sort.SliceStable(trades, func(i, j int) bool {
		return trades[i].SwapRequest.Timestamp < trades[j].SwapRequest.Timestamp
	})

	fees := make([]FeeInfo, 0, len(trades))
	total := make(map[string]int64)
	for _, trade := range trades {
		feeBasisPoint := trade.MarketFee
		swapRequest := trade.SwapRequestMessage()
		feeAsset := swapRequest.GetAssetP()
		amountP := swapRequest.GetAmountP()
		_, percentageFeeAmount := mathutil.LessFee(amountP, uint64(feeBasisPoint))

		marketPrice := trade.MarketPrice.BasePrice
		fixedFeeAmount := uint64(trade.MarketFixedQuoteFee)
		if feeAsset == m.BaseAsset {
			marketPrice = trade.MarketPrice.QuotePrice
			fixedFeeAmount = uint64(trade.MarketFixedBaseFee)
		}

		fees = append(fees, FeeInfo{
			TradeID:             trade.ID.String(),
			BasisPoint:          feeBasisPoint,
			Asset:               feeAsset,
			PercentageFeeAmount: percentageFeeAmount,
			FixedFeeAmount:      fixedFeeAmount,
			MarketPrice:         marketPrice,
		})

		total[feeAsset] += int64(percentageFeeAmount) + int64(fixedFeeAmount)
	}

	return &ReportMarketFee{
		CollectedFees:              fees,
		TotalCollectedFeesPerAsset: total,
	}, nil
}

func (o *operatorService) WithdrawMarketFunds(
	ctx context.Context, req WithdrawMarketReq,
) ([]byte, []byte, error) {
	if err := req.Market.Validate(); err != nil {
		return nil, nil, err
	}

	market, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, req.BaseAsset, req.QuoteAsset,
	)
	if err != nil {
		return nil, nil, err
	}
	if accountIndex < 0 {
		return nil, nil, ErrMarketNotExist
	}

	// Eventually, check fee and market account to notify for low balances.
	defer func() {
		go checkFeeAndMarketBalances(
			o.repoManager, o.blockchainListener.PubSubService(),
			ctx, market, o.network.AssetID, o.feeAccountBalanceThreshold,
		)
	}()

	vault, err := o.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		return nil, nil, err
	}

	balance, err := getUnlockedBalanceForMarket(
		o.repoManager, ctx, market,
	)
	if err != nil {
		return nil, nil, err
	}
	if balance.BaseAmount <= uint64(market.FixedFee.BaseFee) ||
		balance.QuoteAmount <= uint64(market.FixedFee.QuoteFee) {
		return nil, nil, ErrMarketBalanceTooLow
	}

	baseBalance, quoteBalance := balance.BaseAmount, balance.QuoteAmount

	if req.BalanceToWithdraw.BaseAmount > baseBalance {
		return nil, nil, ErrWithdrawBaseAmountTooBig
	}

	if req.BalanceToWithdraw.QuoteAmount > quoteBalance {
		return nil, nil, ErrWithdrawQuoteAmountTooBig
	}

	outs := make([]TxOut, 0)
	if req.BalanceToWithdraw.BaseAmount > 0 {
		outs = append(outs, TxOut{
			Asset:   req.BaseAsset,
			Value:   int64(req.BalanceToWithdraw.BaseAmount),
			Address: req.Address,
		})
	}
	if req.BalanceToWithdraw.QuoteAmount > 0 {
		outs = append(outs, TxOut{
			Asset:   req.QuoteAsset,
			Value:   int64(req.BalanceToWithdraw.QuoteAmount),
			Address: req.Address,
		})
	}

	outputs, outputsBlindingKeys, err := parseRequestOutputs(outs)
	if err != nil {
		return nil, nil, err
	}

	marketUnspents, err := o.getAllUnspentsForAccount(ctx, market.AccountIndex)
	if err != nil {
		return nil, nil, err
	}

	feeUnspents, err := o.getAllUnspentsForAccount(ctx, domain.FeeAccount)
	if err != nil {
		return nil, nil, err
	}

	mnemonic, err := vault.GetMnemonicSafe()
	if err != nil {
		return nil, nil, err
	}

	marketAccount, err := vault.AccountByIndex(market.AccountIndex)
	if err != nil {
		return nil, nil, err
	}
	feeAccount, err := vault.AccountByIndex(domain.FeeAccount)
	if err != nil {
		return nil, nil, err
	}

	changePathsByAsset := map[string]string{}
	feeChangePathByAsset := map[string]string{}
	for _, asset := range getAssetsOfOutputs(outputs) {
		info, err := vault.DeriveNextInternalAddressForAccount(accountIndex)
		if err != nil {
			return nil, nil, err
		}

		derivationPath := marketAccount.DerivationPathByScript[info.Script]
		changePathsByAsset[asset] = derivationPath
	}

	feeInfo, err := vault.DeriveNextInternalAddressForAccount(domain.FeeAccount)
	if err != nil {
		return nil, nil, err
	}
	feeChangePathByAsset[o.network.AssetID] =
		feeAccount.DerivationPathByScript[feeInfo.Script]

	txHex, err := sendToManyWithFeeTopup(sendToManyWithFeeTopupOpts{
		mnemonic:              mnemonic,
		unspents:              marketUnspents,
		feeUnspents:           feeUnspents,
		outputs:               outputs,
		outputsBlindingKeys:   outputsBlindingKeys,
		changePathsByAsset:    changePathsByAsset,
		feeChangePathByAsset:  feeChangePathByAsset,
		inputPathsByScript:    marketAccount.DerivationPathByScript,
		feeInputPathsByScript: feeAccount.DerivationPathByScript,
		milliSatPerByte:       int(req.MillisatPerByte),
		network:               o.network,
	})
	if err != nil {
		return nil, nil, err
	}

	var txid string
	if req.Push {
		cb := circuitbreaker.NewCircuitBreaker()
		iTxid, err := cb.Execute(func() (interface{}, error) {
			return o.explorerSvc.BroadcastTransaction(txHex)
		})
		if err != nil {
			return nil, nil, err
		}
		txid = iTxid.(string)
		log.Debugf("withdrawal tx broadcasted with id: %s", txid)
	}

	if err := o.repoManager.VaultRepository().UpdateVault(
		ctx,
		func(_ *domain.Vault) (*domain.Vault, error) {
			return vault, nil
		},
	); err != nil {
		return nil, nil, err
	}

	go extractUnspentsFromTxAndUpdateUtxoSet(
		o.repoManager.UnspentRepository(),
		o.repoManager.VaultRepository(),
		o.network,
		txHex,
		market.AccountIndex,
	)

	// Start watching tx to confirm new unspents once the tx is in blockchain.
	go o.blockchainListener.StartObserveTx(txid, req.Market)

	// Publish message for topic AccountWithdraw to pubsub service.
	go func() {
		if err := publishMarketWithdrawTopic(
			o.blockchainListener.PubSubService(),
			req.Market, *balance, req.BalanceToWithdraw, req.Address, txid,
		); err != nil {
			log.Warn(err)
		}
	}()

	go func() {
		count, err := o.repoManager.WithdrawalRepository().AddWithdrawals(
			ctx,
			[]domain.Withdrawal{
				{
					TxID:            txid,
					AccountIndex:    accountIndex,
					BaseAmount:      req.BalanceToWithdraw.BaseAmount,
					QuoteAmount:     req.BalanceToWithdraw.QuoteAmount,
					MillisatPerByte: req.MillisatPerByte,
					Address:         req.Address,
				},
			},
		)
		if err != nil {
			log.WithError(err).Warn("an error occured while storing withdrawal info")
			return
		}
		log.Debugf("added %d withdrawals", count)
	}()

	rawTx, _ := hex.DecodeString(txHex)
	rawTxid, _ := hex.DecodeString(txid)
	return rawTx, rawTxid, nil
}

// UpdateMarketPercentageFee changes the Liquidity Provider fee for the given market.
// MUST be expressed as basis point.
// Eg. To change the fee on each swap from 0.25% to 1% you need to pass down 100
// The Market MUST be closed before doing this change.
func (o *operatorService) UpdateMarketPercentageFee(
	ctx context.Context, req MarketWithFee,
) (*MarketWithFee, error) {
	if err := req.Market.Validate(); err != nil {
		return nil, err
	}

	mkt, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, req.BaseAsset, req.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	if err := mkt.ChangeFeeBasisPoint(req.BasisPoint); err != nil {
		return nil, err
	}

	if err := o.repoManager.MarketRepository().UpdateMarket(
		ctx, accountIndex, func(_ *domain.Market) (*domain.Market, error) {
			return mkt, nil
		},
	); err != nil {
		return nil, err
	}

	return &MarketWithFee{
		Market: Market{
			BaseAsset:  mkt.BaseAsset,
			QuoteAsset: mkt.QuoteAsset,
		},
		Fee: Fee{
			BasisPoint:    mkt.Fee,
			FixedBaseFee:  mkt.FixedFee.BaseFee,
			FixedQuoteFee: mkt.FixedFee.QuoteFee,
		},
	}, nil
}

// UpdateMarketFixedFee changes the Liquidity Provider fee for the given market.
// Values for both assets MUST be expressed as satoshis.
func (o *operatorService) UpdateMarketFixedFee(
	ctx context.Context, req MarketWithFee,
) (*MarketWithFee, error) {
	if err := req.Market.Validate(); err != nil {
		return nil, err
	}

	mkt, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, req.BaseAsset, req.QuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if accountIndex < 0 {
		return nil, ErrMarketNotExist
	}

	if err := mkt.ChangeFixedFee(req.FixedBaseFee, req.FixedQuoteFee); err != nil {
		return nil, err
	}

	if err := o.repoManager.MarketRepository().UpdateMarket(
		ctx,
		accountIndex,
		func(_ *domain.Market) (*domain.Market, error) {
			return mkt, nil
		},
	); err != nil {
		return nil, err
	}

	return &MarketWithFee{
		Market: Market{
			BaseAsset:  mkt.BaseAsset,
			QuoteAsset: mkt.QuoteAsset,
		},
		Fee: Fee{
			BasisPoint:    mkt.Fee,
			FixedBaseFee:  mkt.FixedFee.BaseFee,
			FixedQuoteFee: mkt.FixedFee.QuoteFee,
		},
	}, nil
}

// UpdateMarketPrice rpc updates the price for the given market
func (o *operatorService) UpdateMarketPrice(
	ctx context.Context, req MarketWithPrice,
) error {
	if err := req.Market.Validate(); err != nil {
		return err
	}
	if err := req.Price.Validate(); err != nil {
		return err
	}

	_, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, req.BaseAsset, req.QuoteAsset,
	)
	if err != nil {
		return err
	}
	if accountIndex < 0 {
		return ErrMarketNotExist
	}

	// Updates the base price and the quote price
	return o.repoManager.MarketRepository().UpdatePrices(
		ctx,
		accountIndex,
		domain.Prices{
			BasePrice:  req.Price.BasePrice,
			QuotePrice: req.Price.QuotePrice,
		},
	)
}

// UpdateMarketStrategy changes the current market making strategy,
// either using an automated market making formula or a pluggable price feed
func (o *operatorService) UpdateMarketStrategy(
	ctx context.Context, req MarketStrategy,
) error {
	if err := req.Market.Validate(); err != nil {
		return err
	}

	_, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, req.BaseAsset, req.QuoteAsset,
	)
	if err != nil {
		return err
	}

	if accountIndex < 0 {
		return ErrMarketNotExist
	}

	requestStrategy := req.Strategy

	return o.repoManager.MarketRepository().UpdateMarket(
		ctx,
		accountIndex,
		func(m *domain.Market) (*domain.Market, error) {
			switch requestStrategy {
			case domain.StrategyTypePluggable:
				if err := m.MakeStrategyPluggable(); err != nil {
					return nil, err
				}

			case domain.StrategyTypeBalanced:
				if err := m.MakeStrategyBalanced(); err != nil {
					return nil, err
				}

			default:
				return nil, ErrUnknownStrategy
			}

			return m, nil
		},
	)
}

func (o *operatorService) GetFeeFragmenterAddress(
	ctx context.Context, numOfAddresses int,
) ([]AddressAndBlindingKey, error) {
	return o.generateAddressesForAccount(
		ctx, domain.FeeFragmenterAccount, numOfAddresses,
	)
}

func (o *operatorService) ListFeeFragmenterExternalAddresses(
	ctx context.Context,
) ([]AddressAndBlindingKey, error) {
	return o.listExternalAddressesForAccount(ctx, domain.FeeFragmenterAccount)
}

func (o *operatorService) GetFeeFragmenterBalance(
	ctx context.Context,
) (map[string]BalanceInfo, error) {
	return getAccountBalanceFromExplorer(
		o.repoManager, o.explorerSvc, ctx, domain.FeeFragmenterAccount,
	)
}

func (o *operatorService) FeeFragmenterSplitFunds(
	ctx context.Context, maxFragments uint32, millisatsPerByte uint64,
	chRes chan FragmenterSplitFundsReply,
) {
	defer close(chRes)

	chRes <- FragmenterSplitFundsReply{
		Msg: "fetching fee fragmenter funds",
	}

	accountIndex := domain.FeeFragmenterAccount
	utxos, err := getAccountUtxosFromExplorer(
		o.repoManager, o.explorerSvc, ctx, accountIndex,
	)
	if err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf(
				"error while fetching fee fragmenter funds: %s", err,
			),
		}
		return
	}
	if len(utxos) <= 0 {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("no funds detected for fee fragmenter"),
		}
		return
	}

	o.splitFeeFragmenterFunds(ctx, utxos, maxFragments, int(millisatsPerByte), chRes)
}

func (o *operatorService) WithdrawFeeFragmenterFunds(
	ctx context.Context, addr string, millisatsPerByte uint64,
) (string, error) {
	return o.withdrawFragmenterAccount(
		ctx, domain.FeeFragmenterAccount, nil, addr, millisatsPerByte,
	)
}

func (o *operatorService) GetMarketFragmenterAddress(
	ctx context.Context, numOfAddresses int,
) ([]AddressAndBlindingKey, error) {
	return o.generateAddressesForAccount(
		ctx, domain.MarketFragmenterAccount, numOfAddresses,
	)
}

func (o *operatorService) ListMarketFragmenterExternalAddresses(
	ctx context.Context,
) ([]AddressAndBlindingKey, error) {
	return o.listExternalAddressesForAccount(ctx, domain.MarketFragmenterAccount)
}

func (o *operatorService) GetMarketFragmenterBalance(
	ctx context.Context,
) (map[string]BalanceInfo, error) {
	return getAccountBalanceFromExplorer(
		o.repoManager, o.explorerSvc, ctx, domain.MarketFragmenterAccount,
	)
}

func (o *operatorService) MarketFragmenterSplitFunds(
	ctx context.Context, market Market, millisatsPerByte uint64,
	chRes chan FragmenterSplitFundsReply,
) {
	defer close(chRes)

	if err := market.Validate(); err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("invalid market: %s", err),
		}
		return
	}

	mkt, accountIndex, err := o.repoManager.MarketRepository().GetMarketByAssets(
		ctx, market.BaseAsset, market.QuoteAsset,
	)
	if err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("failed to retrieve market: %s", err),
		}
		return
	}
	if accountIndex < 0 {
		chRes <- FragmenterSplitFundsReply{
			Err: ErrMarketNotExist,
		}
		return
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "fetching market fragmenter funds",
	}

	utxos, err := getAccountUtxosFromExplorer(
		o.repoManager, o.explorerSvc, ctx, domain.MarketFragmenterAccount,
	)
	if err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf(
				"error while fetching market fragmenter funds: %s", err,
			),
		}
		return
	}
	if len(utxos) <= 0 {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("no funds detected for market fragmenter"),
		}
		return
	}

	o.splitMarketFragmenterFunds(ctx, mkt, utxos, int(millisatsPerByte), chRes)
}

func (o *operatorService) WithdrawMarketFragmenterFunds(
	ctx context.Context, addr string, millisatsPerByte uint64,
) (string, error) {
	return o.withdrawFragmenterAccount(
		ctx, domain.MarketFragmenterAccount, nil, addr, millisatsPerByte,
	)
}

// ListMarkets a set of informations about all the markets.
func (o *operatorService) ListMarkets(ctx context.Context) ([]MarketInfo, error) {
	markets, err := o.repoManager.MarketRepository().GetAllMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketInfo := make([]MarketInfo, 0, len(markets))
	for _, market := range markets {
		balance, err := getUnlockedBalanceForMarket(o.repoManager, ctx, &market)
		if err != nil {
			return nil, err
		}
		marketInfo = append(marketInfo, MarketInfo{
			AccountIndex: uint64(market.AccountIndex),
			Market: Market{
				BaseAsset:  market.BaseAsset,
				QuoteAsset: market.QuoteAsset,
			},
			Tradable:     market.Tradable,
			StrategyType: market.Strategy.Type,
			Price:        market.Price,
			Fee: Fee{
				BasisPoint:    market.Fee,
				FixedBaseFee:  market.FixedFee.BaseFee,
				FixedQuoteFee: market.FixedFee.QuoteFee,
			},
			Balance: *balance,
		})
	}

	return marketInfo, nil
}

// ListTrades returns the list of all trads processed by the daemon
func (o *operatorService) ListTrades(
	ctx context.Context, page *Page,
) ([]TradeInfo, error) {
	var trades []*domain.Trade
	var err error
	if page == nil {
		trades, err = o.repoManager.TradeRepository().GetAllTrades(ctx)
	} else {
		pg := page.ToDomain()
		trades, err = o.repoManager.TradeRepository().GetAllTradesForPage(ctx, pg)
	}
	if err != nil {
		return nil, err
	}

	return tradesToTradeInfo(trades, o.marketBaseAsset, o.network.Name), nil
}

func (o *operatorService) ListTradesForMarket(
	ctx context.Context, market Market, page *Page,
) ([]TradeInfo, error) {
	var trades []*domain.Trade
	var err error
	if page == nil {
		trades, err = o.repoManager.TradeRepository().GetAllTradesByMarket(
			ctx, market.QuoteAsset,
		)
	} else {
		pg := page.ToDomain()
		trades, err = o.repoManager.TradeRepository().GetAllTradesByMarketAndPage(
			ctx, market.QuoteAsset, pg,
		)
	}
	if err != nil {
		return nil, err
	}

	return tradesToTradeInfo(trades, market.BaseAsset, o.network.Name), nil
}

func (o *operatorService) ListUtxos(
	ctx context.Context, accountIndex int, page *Page,
) (*UtxoInfoList, error) {
	info, err := o.repoManager.VaultRepository().
		GetAllDerivedAddressesInfoForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}

	var allUtxos []domain.Unspent
	if page == nil {
		allUtxos, err = o.repoManager.UnspentRepository().
			GetAllUnspentsForAddresses(ctx, info.Addresses())
	} else {
		pg := page.ToDomain()
		allUtxos, err = o.repoManager.UnspentRepository().
			GetAllUnspentsForAddressesAndPage(ctx, info.Addresses(), pg)
	}
	if err != nil {
		return nil, err
	}

	unspents := make([]UtxoInfo, 0)
	spents := make([]UtxoInfo, 0)
	locks := make([]UtxoInfo, 0)
	for _, u := range allUtxos {
		if u.Spent {
			spents = appendUtxoInfo(spents, u)
		} else if u.Locked {
			locks = appendUtxoInfo(locks, u)
		} else {
			unspents = appendUtxoInfo(unspents, u)
		}
	}

	return &UtxoInfoList{
		Unspents: unspents,
		Spents:   spents,
		Locks:    locks,
	}, nil
}

// ReloadUtxos triggers reloading of unspents for stored addresses from blockchain
func (o *operatorService) ReloadUtxos(ctx context.Context) error {
	vault, err := o.repoManager.VaultRepository().GetOrCreateVault(
		ctx, nil, "", nil,
	)
	if err != nil {
		return err
	}

	addressesInfo := vault.AllDerivedAddressesInfo()
	_, err = fetchAndAddUnspents(
		o.explorerSvc,
		o.repoManager.UnspentRepository(),
		o.blockchainListener,
		addressesInfo,
	)
	return err
}

func (o *operatorService) ListDeposits(
	ctx context.Context, accountIndex int, page *Page,
) (Deposits, error) {
	var deposits []domain.Deposit
	var err error
	if page == nil {
		deposits, err = o.repoManager.DepositRepository().ListDepositsForAccount(
			ctx, accountIndex,
		)
	} else {
		pg := page.ToDomain()
		deposits, err = o.repoManager.DepositRepository().ListDepositsForAccountAndPage(
			ctx, accountIndex, pg,
		)
	}
	if err != nil {
		return nil, err
	}

	return Deposits(deposits), nil
}

func (o *operatorService) ListWithdrawals(
	ctx context.Context, accountIndex int, page *Page,
) (Withdrawals, error) {
	var withdrawals []domain.Withdrawal
	var err error
	if page == nil {
		withdrawals, err = o.repoManager.WithdrawalRepository().ListWithdrawalsForAccount(
			ctx, accountIndex,
		)
	} else {
		pg := page.ToDomain()
		withdrawals, err = o.repoManager.WithdrawalRepository().ListWithdrawalsForAccountAndPage(
			ctx, accountIndex, pg,
		)
	}
	if err != nil {
		return nil, err
	}

	return Withdrawals(withdrawals), nil
}

func (o *operatorService) AddWebhook(
	_ context.Context, hook Webhook,
) (string, error) {
	if o.blockchainListener.PubSubService() == nil {
		return "", ErrPubSubServiceNotInitialized
	}

	topics := o.blockchainListener.PubSubService().TopicsByCode()
	topic, ok := topics[hook.ActionType]
	if !ok {
		return "", ErrInvalidActionType
	}

	return o.blockchainListener.PubSubService().Subscribe(
		topic.Label(), hook.Endpoint, hook.Secret,
	)
}

func (o *operatorService) RemoveWebhook(
	_ context.Context, hookID string,
) error {
	if o.blockchainListener.PubSubService() == nil {
		return ErrPubSubServiceNotInitialized
	}
	return o.blockchainListener.PubSubService().Unsubscribe("", hookID)
}

func (o *operatorService) ListWebhooks(
	_ context.Context, actionType int,
) ([]WebhookInfo, error) {
	pubsubSvc := o.blockchainListener.PubSubService()
	if pubsubSvc == nil {
		return nil, ErrPubSubServiceNotInitialized
	}

	topics := pubsubSvc.TopicsByCode()
	topic, ok := topics[actionType]
	if !ok {
		return nil, ErrInvalidActionType
	}

	subs := pubsubSvc.ListSubscriptionsForTopic(topic.Label())
	hooks := make([]WebhookInfo, 0, len(subs))
	for _, s := range subs {
		hooks = append(hooks, WebhookInfo{
			Id:         s.Id(),
			ActionType: s.Topic().Code(),
			Endpoint:   s.NotifyAt(),
			IsSecured:  s.IsSecured(),
		})
	}
	return hooks, nil
}

func (o *operatorService) generateAddressesForAccount(
	ctx context.Context, accountIndex, numOfAddresses int,
) ([]AddressAndBlindingKey, error) {
	if numOfAddresses <= 0 {
		numOfAddresses = 1
	}

	vault, err := o.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		return nil, err
	}

	list := make([]AddressAndBlindingKey, 0, numOfAddresses)
	for i := 0; i < numOfAddresses; i++ {
		info, err := vault.DeriveNextExternalAddressForAccount(accountIndex)
		if err != nil {
			return nil, err
		}

		list = append(list, AddressAndBlindingKey{
			Address:     info.Address,
			BlindingKey: hex.EncodeToString(info.BlindingKey),
		})
	}

	if err := o.repoManager.VaultRepository().UpdateVault(
		ctx,
		func(_ *domain.Vault) (*domain.Vault, error) {
			return vault, nil
		},
	); err != nil {
		return nil, err
	}

	return list, nil
}

func (o *operatorService) listExternalAddressesForAccount(
	ctx context.Context, accountIndex int,
) ([]AddressAndBlindingKey, error) {
	allInfo, err := o.repoManager.VaultRepository().
		GetAllDerivedExternalAddressesInfoForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}

	addresses, keys := allInfo.AddressesAndKeys()
	res := make([]AddressAndBlindingKey, 0, len(addresses))
	for i, addr := range addresses {
		res = append(res, AddressAndBlindingKey{
			Address:     addr,
			BlindingKey: hex.EncodeToString(keys[i]),
		})
	}

	return res, nil
}

func (o *operatorService) claimDeposit(
	ctx context.Context,
	accountInfo domain.AddressesInfo,
	outpoints []TxOutpoint,
	market *domain.Market,
) error {
	accountIndex := domain.FeeAccount
	if market != nil {
		accountIndex = market.AccountIndex
	}
	// Group all addresses info by script
	infoByScript := make(map[string]domain.AddressInfo)
	for s, i := range groupAddressesInfoByScript(accountInfo) {
		infoByScript[s] = i
	}

	// For each outpoint retrieve the raw tx and output. If the output script
	// exists in infoByScript, increment the counter of the related account and
	// unblind the raw confidential output.
	// Since all outpoints MUST be funds of the same account, at the end of the
	// loop there MUST be only one counter matching the length of the given
	// outpoints.
	counter := 0
	unspents := make([]domain.Unspent, 0, len(outpoints))
	deposits := make([]domain.Deposit, 0, len(outpoints))
	unconfirmedTxs := make([]string, 0)
	for _, v := range outpoints {
		confirmed, err := o.explorerSvc.IsTransactionConfirmed(v.Hash)
		if err != nil {
			return err
		}

		tx, err := o.explorerSvc.GetTransaction(v.Hash)
		if err != nil {
			return err
		}

		if len(tx.Outputs()) <= v.Index {
			return ErrInvalidOutpoint
		}

		txOut := tx.Outputs()[v.Index]
		script := hex.EncodeToString(txOut.Script)
		if info, ok := infoByScript[script]; ok {
			counter++

			unconfidential, ok := BlinderManager.UnblindOutput(
				txOut,
				info.BlindingKey,
			)
			if !ok {
				return errors.New("unable to unblind output")
			}

			unspents = append(unspents, domain.Unspent{
				TxID:            v.Hash,
				VOut:            uint32(v.Index),
				Value:           unconfidential.Value,
				AssetHash:       unconfidential.AssetHash,
				ValueCommitment: bufferutil.CommitmentFromBytes(txOut.Value),
				AssetCommitment: bufferutil.CommitmentFromBytes(txOut.Asset),
				ValueBlinder:    unconfidential.ValueBlinder,
				AssetBlinder:    unconfidential.AssetBlinder,
				ScriptPubKey:    txOut.Script,
				Nonce:           txOut.Nonce,
				RangeProof:      make([]byte, 1),
				SurjectionProof: make([]byte, 1),
				Address:         info.Address,
				Confirmed:       confirmed,
			})

			deposits = append(deposits, domain.Deposit{
				AccountIndex: info.AccountIndex,
				TxID:         v.Hash,
				VOut:         v.Index,
				Asset:        unconfidential.AssetHash,
				Value:        unconfidential.Value,
			})

			if !confirmed {
				unconfirmedTxs = append(unconfirmedTxs, v.Hash)
			}
		}
	}

	if counter == len(outpoints) {
		if market != nil {
			if err := verifyMarketFunds(market, unspents); err != nil {
				return err
			}
			log.Infof("funded market with account %d", accountIndex)
		}

		go func() {
			addUnspentsAsync(o.repoManager.UnspentRepository(), unspents)
			count, err := o.repoManager.DepositRepository().AddDeposits(
				ctx, deposits,
			)
			if err != nil {
				log.WithError(err).Warn("an error occured while storing deposits info")
			} else {
				log.Debugf("added %d deposits for account %d", count, accountIndex)
			}
			if market == nil {
				if err := o.checkFeeBalance(accountInfo); err != nil {
					log.Warn(err)
					return
				}
				log.Info("fee account funded. Trades can be served")
			}
		}()

		// Start watching for those funds that are not yet confirmed.
		go func() {
			for _, txid := range unconfirmedTxs {
				var mkt Market
				if market != nil {
					mkt = Market{
						BaseAsset:  market.BaseAsset,
						QuoteAsset: market.QuoteAsset,
					}
				}
				o.blockchainListener.StartObserveTx(txid, mkt)
			}
		}()

		return nil
	}

	return ErrInvalidOutpoints
}

func verifyMarketFunds(
	market *domain.Market, unspents []domain.Unspent,
) error {
	outpoints := make([]domain.OutpointWithAsset, 0, len(unspents))
	for _, u := range unspents {
		outpoints = append(outpoints, u.ToOutpointWithAsset())
	}
	return market.VerifyMarketFunds(outpoints)
}

func (o *operatorService) checkFeeBalance(accountInfo domain.AddressesInfo) error {
	feeAccountBalance, err := o.repoManager.UnspentRepository().GetBalance(
		context.Background(),
		accountInfo.Addresses(),
		o.network.AssetID,
	)
	if err != nil {
		return err
	}

	if feeAccountBalance < o.feeAccountBalanceThreshold {
		return errors.New(
			"fee account balance for account index too low. Trades for markets " +
				"won't be served properly. Fund the fee account as soon as possible",
		)
	}

	return nil
}

func (o *operatorService) getAllUnspentsForAccount(
	ctx context.Context, accountIndex int,
) ([]explorer.Utxo, error) {
	info, err := o.repoManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}

	unspents, err := o.repoManager.UnspentRepository().GetAvailableUnspentsForAddresses(
		ctx,
		info.Addresses(),
	)
	if err != nil {
		return nil, err
	}

	utxos := make([]explorer.Utxo, 0, len(unspents))
	for _, u := range unspents {
		utxos = append(utxos, u.ToUtxo())
	}
	return utxos, nil
}

func (o *operatorService) splitFeeFragmenterFunds(
	ctx context.Context,
	utxos []explorer.Utxo, maxFragments uint32, millisatsPerByte int,
	chRes chan FragmenterSplitFundsReply,
) {
	lbtc := o.network.AssetID
	if maxFragments == 0 {
		maxFragments = DefaultFeeFragmenterFragments
	}
	if millisatsPerByte == 0 {
		millisatsPerByte = domain.MinMilliSatPerByte
	}
	vault, _ := o.repoManager.VaultRepository().GetOrCreateVault(
		ctx, nil, "", nil,
	)
	mnemonic, _ := vault.GetMnemonicSafe()

	amountPerAsset := make(map[string]uint64)
	for _, u := range utxos {
		amountPerAsset[u.Asset()] += u.Value()
	}
	if _, ok := amountPerAsset[lbtc]; !ok {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf(
				"no LBTC funds found for the ephemeral wallet. In case you sent funds " +
					"of other asset types you MUST abort the fragmentation and send back " +
					"all to an address of yours",
			),
		}
		return
	}
	if len(amountPerAsset) != 1 {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf(
				"found funds with asset different from LBTC. Use a recover address " +
					"to get them back",
			),
		}
		return
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "calculating fragments for LBTC funds",
	}

	totalAmount := amountPerAsset[lbtc]
	fragmentedAmounts := feeFragmentAmount(
		totalAmount, DefaultFeeFragmenterAmount, maxFragments,
	)

	chRes <- FragmenterSplitFundsReply{
		Msg: fmt.Sprintf(
			"detected %d fund(s) of total amount %d that will be split into %d "+
				"fragments", len(utxos), totalAmount, len(fragmentedAmounts),
		),
	}

	feeAddresses := make([]string, 0, len(fragmentedAmounts))
	for range fragmentedAmounts {
		info, err := vault.DeriveNextExternalAddressForAccount(domain.FeeAccount)
		if err != nil {
			chRes <- FragmenterSplitFundsReply{
				Err: fmt.Errorf("failed to derive new address for fee account: %s", err),
			}
			return
		}
		feeAddresses = append(feeAddresses, info.Address)
	}
	accountInfo, err := vault.AllDerivedExternalAddressesInfoForAccount(domain.FeeAccount)
	if err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("failed to retrieve fee account info: %s", err),
		}
		return
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "crafting fee deposit transaction",
	}

	changeInfo, err := vault.DeriveNextInternalAddressForAccount(domain.FeeFragmenterAccount)
	if err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("failed to generate change address: %s", err),
		}
		return
	}
	changePathByAsset := map[string]string{
		lbtc: changeInfo.DerivationPath,
	}
	outs := createOutputs(fragmentedAmounts, nil, feeAddresses, pair{baseAsset: lbtc})
	outputs, outBlindKeys, _ := parseRequestOutputs(outs)
	txHex, err := sendToManyWithoutFeeTopup(sendToManyWithoutFeeTopupOpts{
		mnemonic:            mnemonic,
		unspents:            utxos,
		outputs:             outputs,
		outputsBlindingKeys: outBlindKeys,
		changePathsByAsset:  changePathByAsset,
		inputPathsByScript:  vault.Accounts[domain.FeeFragmenterAccount].DerivationPathByScript,
		milliSatPerByte:     millisatsPerByte,
		network:             o.network,
		subtractFees:        true,
	})
	if err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("failed to craft fee deposit transaction: %s", err),
		}
		return
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "broadcasting transaction",
	}

	cb := circuitbreaker.NewCircuitBreaker()
	iTxid, err := cb.Execute(func() (interface{}, error) {
		return o.explorerSvc.BroadcastTransaction(txHex)
	})
	if err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("failed to broadcast transaction: %s", err),
		}
		return
	}
	txid := iTxid.(string)

	chRes <- FragmenterSplitFundsReply{
		Msg: fmt.Sprintf("fee account funding transaction: %s", txid),
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "claiming deposits for fee account",
	}

	outpoints := createOutpoints(txid, len(fragmentedAmounts))
	if err := o.claimDeposit(ctx, accountInfo, outpoints, nil); err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("failed to claim deposits: %s", err),
		}
		return
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "fragmentation succeeded",
	}

	go func() {
		if err := o.repoManager.VaultRepository().UpdateVault(
			ctx, func(_ *domain.Vault) (*domain.Vault, error) {
				return vault, nil
			},
		); err != nil {
			log.WithError(err).Warn("an error occured while updating vault")
		}
	}()
}

func (o *operatorService) splitMarketFragmenterFunds(
	ctx context.Context,
	market *domain.Market, utxos []explorer.Utxo, millisatsPerByte int,
	chRes chan FragmenterSplitFundsReply,
) {
	vault, _ := o.repoManager.VaultRepository().GetOrCreateVault(
		ctx, nil, "", nil,
	)
	if millisatsPerByte == 0 {
		millisatsPerByte = domain.MinMilliSatPerByte
	}

	amountPerAsset := make(map[string]uint64)
	for _, u := range utxos {
		amountPerAsset[u.Asset()] += u.Value()
	}

	var assetValuePair pair
	for k, v := range amountPerAsset {
		if k == market.BaseAsset {
			assetValuePair.baseAsset = k
			assetValuePair.baseValue = v
		} else if k == market.QuoteAsset {
			assetValuePair.quoteAsset = k
			assetValuePair.quoteValue = v
		} else {
			chRes <- FragmenterSplitFundsReply{
				Err: fmt.Errorf(
					"fetched funds of asset different from the market pair. " +
						"Recover all the funds and abort this fragmentation",
				),
			}
			return
		}
	}

	if assetValuePair.baseValue == 0 {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf(
				"missing base asset funds",
			),
		}
		return
	}
	if assetValuePair.quoteValue == 0 {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf(
				"missing quote asset funds",
			),
		}
		return
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "calculating fragments for market asset pair",
	}

	baseFragments, quoteFragments, _ := marketFragmentAmount(
		assetValuePair, FragmentationMap,
	)
	numIns := len(utxos)
	numOuts := len(baseFragments) + len(quoteFragments)

	chRes <- FragmenterSplitFundsReply{
		Msg: fmt.Sprintf("detected %d funds", numIns),
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: fmt.Sprintf(
			"splitting base asset amount %d into %d fragments",
			assetValuePair.baseValue, len(baseFragments),
		),
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: fmt.Sprintf(
			"splitting quote asset amount %d into %d fragments",
			assetValuePair.quoteValue, len(quoteFragments),
		),
	}

	addresses := make([]string, 0, numOuts)
	for i := 0; i < numOuts; i++ {
		info, err := vault.DeriveNextExternalAddressForAccount(market.AccountIndex)
		if err != nil {
			chRes <- FragmenterSplitFundsReply{
				Err: fmt.Errorf("failed to derive new address for market: %s", err),
			}
			return
		}
		addresses = append(addresses, info.Address)
	}
	accountInfo, err := vault.AllDerivedExternalAddressesInfoForAccount(
		market.AccountIndex,
	)
	if err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("failed to retrieve market account info: %s", err),
		}
		return
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "crafting market deposit transaction",
	}

	mnemonic, _ := vault.GetMnemonicSafe()
	changePathsByAsset := make(map[string]string)
	for asset := range amountPerAsset {
		info, err := vault.DeriveNextInternalAddressForAccount(domain.MarketFragmenterAccount)
		if err != nil {
			chRes <- FragmenterSplitFundsReply{
				Err: fmt.Errorf("failed to generate change address: %s", err),
			}
			return
		}
		changePathsByAsset[asset] = info.DerivationPath
	}
	outs := createOutputs(baseFragments, quoteFragments, addresses, assetValuePair)
	outputs, outBlindKeys, _ := parseRequestOutputs(outs)
	txHex, err := sendToManyWithoutFeeTopup(sendToManyWithoutFeeTopupOpts{
		mnemonic:            mnemonic,
		unspents:            utxos,
		outputs:             outputs,
		outputsBlindingKeys: outBlindKeys,
		changePathsByAsset:  changePathsByAsset,
		inputPathsByScript:  vault.Accounts[domain.MarketFragmenterAccount].DerivationPathByScript,
		milliSatPerByte:     millisatsPerByte,
		network:             o.network,
		subtractFees:        true,
	})
	if err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("failed to craft market deposit transaction: %s", err),
		}
		return
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "broadcasting transaction",
	}

	cb := circuitbreaker.NewCircuitBreaker()
	iTxid, err := cb.Execute(func() (interface{}, error) {
		return o.explorerSvc.BroadcastTransaction(txHex)
	})
	if err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("failed to broadcast transaction: %s", err),
		}
		return
	}
	txid := iTxid.(string)

	chRes <- FragmenterSplitFundsReply{
		Msg: fmt.Sprintf("market account funding transaction: %s", txid),
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "claiming deposits for market account",
	}

	outpoints := createOutpoints(txid, numOuts)
	if err := o.claimDeposit(ctx, accountInfo, outpoints, market); err != nil {
		chRes <- FragmenterSplitFundsReply{
			Err: fmt.Errorf("failed to claim deposits: %s", err),
		}
		return
	}

	chRes <- FragmenterSplitFundsReply{
		Msg: "fragmentation succeeded",
	}

	go func() {
		if err := o.repoManager.VaultRepository().UpdateVault(
			ctx, func(_ *domain.Vault) (*domain.Vault, error) {
				return vault, nil
			},
		); err != nil {
			log.WithError(err).Warn("an error occured while updating vault")
		}
	}()
}

func (o *operatorService) withdrawFragmenterAccount(
	ctx context.Context, accountIndex int, market *Market,
	addr string, millisatsPerByte uint64,
) (string, error) {
	net, err := address.NetworkForAddress(addr)
	if err != nil {
		return "", fmt.Errorf("address is invalid")
	}
	if net.Name != o.network.Name {
		return "", fmt.Errorf("address is not for network %s", o.network.Name)
	}

	utxos, err := getAccountUtxosFromExplorer(
		o.repoManager, o.explorerSvc, ctx, accountIndex,
	)
	if err != nil {
		return "", err
	}

	vault, _ := o.repoManager.VaultRepository().GetOrCreateVault(
		ctx, nil, "", nil,
	)
	mnemonic, _ := vault.GetMnemonicSafe()
	lbtc := o.network.AssetID
	if millisatsPerByte == 0 {
		millisatsPerByte = domain.MinMilliSatPerByte
	}

	amountPerAsset := make(map[string]uint64)
	for _, u := range utxos {
		amountPerAsset[u.Asset()] += u.Value()
	}
	if _, ok := amountPerAsset[lbtc]; !ok {
		return "", fmt.Errorf(
			"send more LBTC funds to the fragmenter to make it able paying for " +
				"network fees. Extra amount will be sent back to your address",
		)
	}

	outs := make([]TxOut, 0, len(amountPerAsset))
	for asset, amount := range amountPerAsset {
		outs = append(outs, TxOut{
			Asset:   asset,
			Address: addr,
			Value:   int64(amount),
		})
	}

	changePathByAsset := make(map[string]string)
	for asset := range amountPerAsset {
		info, err := vault.DeriveNextInternalAddressForAccount(accountIndex)
		if err != nil {
			return "", err
		}
		changePathByAsset[asset] = info.DerivationPath
	}
	outputs, outBlindKeys, _ := parseRequestOutputs(outs)
	txHex, err := sendToManyWithoutFeeTopup(sendToManyWithoutFeeTopupOpts{
		mnemonic:            mnemonic,
		unspents:            utxos,
		outputs:             outputs,
		outputsBlindingKeys: outBlindKeys,
		changePathsByAsset:  changePathByAsset,
		inputPathsByScript:  vault.Accounts[accountIndex].DerivationPathByScript,
		milliSatPerByte:     int(millisatsPerByte),
		network:             o.network,
		subtractFees:        true,
	})
	if err != nil {
		return "", err
	}

	cb := circuitbreaker.NewCircuitBreaker()
	iTxid, err := cb.Execute(func() (interface{}, error) {
		return o.explorerSvc.BroadcastTransaction(txHex)
	})
	if err != nil {
		return "", err
	}

	return iTxid.(string), nil
}

func tradesToTradeInfo(trades []*domain.Trade, marketBaseAsset, network string) []TradeInfo {
	tradeInfo := make([]TradeInfo, 0, len(trades))
	chInfo := make(chan TradeInfo)
	wg := &sync.WaitGroup{}
	wg.Add(len(trades))

	go func() {
		wg.Wait()
		close(chInfo)
	}()

	for i := range trades {
		trade := trades[i]
		go tradeToTradeInfo(trade, marketBaseAsset, network, chInfo, wg)
	}

	for info := range chInfo {
		tradeInfo = append(tradeInfo, info)
	}

	// sort by request timestamp
	sort.SliceStable(tradeInfo, func(i, j int) bool {
		return tradeInfo[i].RequestTimeUnix < tradeInfo[j].RequestTimeUnix
	})

	return tradeInfo
}

func tradeToTradeInfo(
	trade *domain.Trade,
	marketBaseAsset, net string,
	chInfo chan TradeInfo,
	wg *sync.WaitGroup,
) {
	if wg != nil {
		defer wg.Done()
	}

	if trade.IsEmpty() {
		return
	}
	// to maintain backward compatibility, since trade.MarketBaseAsset has been
	// introduced only in versions above v0.7.1.
	mktBaseAsset := trade.MarketBaseAsset
	if len(mktBaseAsset) == 0 {
		mktBaseAsset = marketBaseAsset
	}

	info := TradeInfo{
		ID:     trade.ID.String(),
		Status: trade.Status,
		MarketWithFee: MarketWithFee{
			Market{
				BaseAsset:  mktBaseAsset,
				QuoteAsset: trade.MarketQuoteAsset,
			},
			Fee{
				BasisPoint:    trade.MarketFee,
				FixedBaseFee:  trade.MarketFixedBaseFee,
				FixedQuoteFee: trade.MarketFixedQuoteFee,
			},
		},
		Price:            Price(trade.MarketPrice),
		RequestTimeUnix:  trade.SwapRequest.Timestamp,
		AcceptTimeUnix:   trade.SwapAccept.Timestamp,
		CompleteTimeUnix: trade.SwapComplete.Timestamp,
		SettleTimeUnix:   trade.SettlementTime,
		ExpiryTimeUnix:   trade.ExpiryTime,
	}

	if req := trade.SwapRequestMessage(); req != nil {
		info.SwapInfo = SwapInfo{
			AssetP:  req.GetAssetP(),
			AmountP: req.GetAmountP(),
			AssetR:  req.GetAssetR(),
			AmountR: req.GetAmountR(),
		}
	}

	if fail := trade.SwapFailMessage(); fail != nil {
		info.SwapFailInfo = SwapFailInfo{
			Code:    int(fail.GetFailureCode()),
			Message: fail.GetFailureMessage(),
		}
	}

	if trade.IsSettled() {
		_, outBlindingData, _ := TransactionManager.ExtractBlindingData(
			trade.PsetBase64,
			nil, trade.SwapAcceptMessage().GetOutputBlindingKey(),
		)

		var blinded string
		for _, data := range outBlindingData {
			blinded += fmt.Sprintf(
				"%d,%s,%s,%s,",
				data.Amount, data.Asset,
				hex.EncodeToString(elementsutil.ReverseBytes(data.AmountBlinder)),
				hex.EncodeToString(elementsutil.ReverseBytes(data.AssetBlinder)),
			)
		}
		// remove trailing comma
		blinded = strings.Trim(blinded, ",")

		baseURL := "https://blockstream.info/liquid/tx"
		if net == network.Regtest.Name {
			baseURL = "http://localhost:3001/tx"
		}
		info.TxURL = fmt.Sprintf("%s/%s#blinded=%s", baseURL, trade.TxID, blinded)
	}

	chInfo <- info
}

func groupAddressesInfoByScript(info domain.AddressesInfo) map[string]domain.AddressInfo {
	group := make(map[string]domain.AddressInfo)
	for _, i := range info {
		group[i.Script] = i
	}
	return group
}

func appendUtxoInfo(list []UtxoInfo, unspent domain.Unspent) []UtxoInfo {
	return append(list, UtxoInfo{
		Outpoint: &TxOutpoint{
			Hash:  unspent.TxID,
			Index: int(unspent.VOut),
		},
		Value: unspent.Value,
		Asset: unspent.AssetHash,
	})
}

type sendToManyWithoutFeeTopupOpts struct {
	mnemonic            []string
	unspents            []explorer.Utxo
	outputs             []*transaction.TxOutput
	outputsBlindingKeys [][]byte
	changePathsByAsset  map[string]string
	inputPathsByScript  map[string]string
	milliSatPerByte     int
	network             *network.Network
	subtractFees        bool
}

func sendToManyWithoutFeeTopup(opts sendToManyWithoutFeeTopupOpts) (string, error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: opts.mnemonic,
	})
	if err != nil {
		return "", err
	}

	// Default to MinMilliSatPerByte if needed
	milliSatPerByte := opts.milliSatPerByte
	if milliSatPerByte < domain.MinMilliSatPerByte {
		milliSatPerByte = domain.MinMilliSatPerByte
	}

	// Create the transaction
	newPset, err := w.CreateTx()
	if err != nil {
		return "", err
	}
	network := opts.network

	// Add inputs and outputs
	updateResult, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64:         newPset,
		Unspents:           opts.unspents,
		Outputs:            opts.outputs,
		ChangePathsByAsset: opts.changePathsByAsset,
		MilliSatsPerBytes:  milliSatPerByte,
		Network:            network,
		SubtractFees:       opts.subtractFees,
		WantChangeForFees:  !opts.subtractFees,
	})
	if err != nil {
		return "", err
	}

	inputBlindingData := make(map[int]wallet.BlindingData)
	index := 0
	for _, v := range updateResult.SelectedUnspents {
		inputBlindingData[index] = wallet.BlindingData{
			Asset:         v.Asset(),
			Amount:        v.Value(),
			AssetBlinder:  v.AssetBlinder(),
			AmountBlinder: v.ValueBlinder(),
		}
		index++
	}

	// Update the list of output blinding keys with those of the eventual changes
	outputsBlindingKeys := opts.outputsBlindingKeys
	for _, v := range updateResult.ChangeOutputsBlindingKeys {
		outputsBlindingKeys = append(outputsBlindingKeys, v)
	}

	// Blind the transaction
	blindedPset, err := w.BlindTransactionWithData(
		wallet.BlindTransactionWithDataOpts{
			PsetBase64:         updateResult.PsetBase64,
			InputBlindingData:  inputBlindingData,
			OutputBlindingKeys: outputsBlindingKeys,
		},
	)
	if err != nil {
		return "", err
	}

	// Ddd the explicit fee amount
	blindedPlusFees, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64: blindedPset,
		Outputs:    transactionutil.NewFeeOutput(updateResult.FeeAmount, network),
		Network:    network,
	})
	if err != nil {
		return "", err
	}

	// Sign the inputs
	signedPset, err := w.SignTransaction(wallet.SignTransactionOpts{
		PsetBase64:        blindedPlusFees.PsetBase64,
		DerivationPathMap: opts.inputPathsByScript,
	})
	if err != nil {
		return "", err
	}

	// Finalize, extract and return the transaction
	txHex, _, err := wallet.FinalizeAndExtractTransaction(
		wallet.FinalizeAndExtractTransactionOpts{
			PsetBase64: signedPset,
		},
	)

	return txHex, err
}
