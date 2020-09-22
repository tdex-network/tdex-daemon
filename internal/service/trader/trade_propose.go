package tradeservice

import (
	"context"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/domain/trade"
	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"github.com/tdex-network/tdex-daemon/internal/storageutil/uow"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	"github.com/vulpemventures/go-elements/address"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TradePropose is the domain controller for the TradePropose RPC
func (s *Service) TradePropose(req *pb.TradeProposeRequest, stream pb.Trade_TradeProposeServer) error {
	_, marketAccountIndex, err := s.marketRepository.GetMarketByAsset(
		context.Background(),
		req.GetMarket().GetQuoteAsset(),
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// get all unspents for market account along with private blinding keys and
	// signing derivation paths for respectively unblinding and signing them later
	marketUnspents, marketBlindingKeysByScript, marketDerivationPaths, err := s.getUnspentsBlindingsAndDerivationPathsForAccount(marketAccountIndex)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// ... and the same for fee account (we'll need to top-up fees)
	feeUnspents, feeBlindingKeysByScript, feeDerivationPaths, err := s.getUnspentsBlindingsAndDerivationPathsForAccount(vault.FeeAccount)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	var reply *pb.TradeProposeReply

	// try to accept the incoming proposal in a transactional way, by committing
	// changes to different storages only if the trade is accepted at the very end.
	// This process causes changes that affect different domains so we need to
	// update all or none of them in case any errors occur.
	unit := uow.NewUnitOfWork(s.tradeRepository, s.unspentRepository)

	if err := unit.Run(func(u uow.Contextual) error {
		var mnemonic []string
		var tradeID uuid.UUID
		var selectedUnspents []explorer.Utxo
		var outputBlindingKeyByScript map[string][]byte
		var outputDerivationPath, changeDerivationPath, feeChangeDerivationPath string

		// derive output and change address for market, and change address for fee account
		vaultCtx := u.Context(s.vaultRepository)
		if err := s.vaultRepository.UpdateVault(vaultCtx, nil, "", func(v *vault.Vault) (*vault.Vault, error) {
			mnemonic, err = v.Mnemonic()
			if err != nil {
				return nil, err
			}
			outputAddress, outputScript, _, err := v.DeriveNextExternalAddressForAccount(marketAccountIndex)
			if err != nil {
				return nil, err
			}
			_, changeScript, _, err := v.DeriveNextInternalAddressForAccount(marketAccountIndex)
			if err != nil {
				return nil, err
			}
			_, feeChangeScript, _, err := v.DeriveNextInternalAddressForAccount(vault.FeeAccount)
			if err != nil {
				return nil, err
			}
			marketAccount, _ := v.AccountByIndex(marketAccountIndex)
			feeAccount, _ := v.AccountByIndex(vault.FeeAccount)

			outputBlindingKeyByScript = blindingKeyByScriptFromCTAddress(outputAddress)
			outputDerivationPath, _ = marketAccount.DerivationPathByScript(outputScript)
			changeDerivationPath, _ = marketAccount.DerivationPathByScript(changeScript)
			feeChangeDerivationPath, _ = feeAccount.DerivationPathByScript(feeChangeScript)

			return v, nil
		}); err != nil {
			return err
		}

		tradeCtx := u.Context(s.tradeRepository)
		// parse swap proposal and possibly accept
		if err := s.tradeRepository.UpdateTrade(tradeCtx, nil, func(t *trade.Trade) (*trade.Trade, error) {
			if err := t.Propose(req.GetSwapRequest(), req.GetMarket().GetQuoteAsset(), nil); err != nil {
				return nil, err
			}
			tradeID = t.ID()

			acceptSwapResult, err := acceptSwap(acceptSwapOpts{
				mnemonic:                   mnemonic,
				swapRequest:                req.GetSwapRequest(),
				marketUnspents:             marketUnspents,
				feeUnspents:                feeUnspents,
				marketBlindingKeysByScript: marketBlindingKeysByScript,
				feeBlindingKeysByScript:    feeBlindingKeysByScript,
				outputBlindingKeyByScript:  outputBlindingKeyByScript,
				marketDerivationPaths:      marketDerivationPaths,
				feeDerivationPaths:         feeDerivationPaths,
				outputDerivationPath:       outputDerivationPath,
				changeDerivationPath:       changeDerivationPath,
				feeChangeDerivationPath:    feeChangeDerivationPath,
			})
			if err != nil {
				return nil, err
			}

			if err := t.Accept(
				acceptSwapResult.psetBase64,
				acceptSwapResult.inputBlindingKeys,
				acceptSwapResult.outputBlindingKeys,
			); err != nil {
				return nil, err
			}

			reply = &pb.TradeProposeReply{
				SwapAccept:     t.SwapAcceptMessage(),
				ExpiryTimeUnix: t.SwapExpiryTime(),
			}

			return t, nil
		}); err != nil {
			return err
		}

		selectedUnspentKeys := getUnspentKeys(selectedUnspents)
		unspentCtx := u.Context(s.unspentRepository)
		if err := s.unspentRepository.LockUnspents(unspentCtx, selectedUnspentKeys, tradeID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if err := stream.Send(reply); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (s *Service) getUnspentsBlindingsAndDerivationPathsForAccount(account int) (
	[]explorer.Utxo,
	map[string][]byte,
	map[string]string,
	error,
) {
	derivedAddresses, blindingKeys, err := s.vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(context.Background(), account)
	if err != nil {
		return nil, nil, nil, err
	}

	scripts := make([]string, 0, len(derivedAddresses))
	for _, addr := range derivedAddresses {
		script, _ := address.ToOutputScript(addr, *config.GetNetwork())
		scripts = append(scripts, hex.EncodeToString(script))
	}
	derivationPaths, _ := s.vaultRepository.GetDerivationPathByScript(context.Background(), account, scripts)

	unspents := s.unspentRepository.GetAvailableUnspentsForAddresses(context.Background(), derivedAddresses)

	utxos, err := s.getUtxos(derivedAddresses, blindingKeys)
	if err != nil {
		return nil, nil, nil, err
	}

	availableUtxos := make([]explorer.Utxo, 0, len(unspents))
	for _, unspent := range unspents {
		unspentKey := unspent.GetKey()
		for _, utxo := range utxos {
			if unspentKey.TxID == utxo.Hash() && unspentKey.VOut == utxo.Index() {
				availableUtxos = append(availableUtxos, utxo)
				break
			}
		}
	}

	blindingKeysByScript := map[string][]byte{}
	for i, addr := range derivedAddresses {
		script, _ := address.ToOutputScript(addr, *config.GetNetwork())
		blindingKeysByScript[hex.EncodeToString(script)] = blindingKeys[i]
	}

	return availableUtxos, blindingKeysByScript, derivationPaths, nil
}

func (s *Service) getUtxos(addresses []string, blindingKeys [][]byte) ([]explorer.Utxo, error) {
	chUnspents := make(chan []explorer.Utxo)
	chErr := make(chan error, 1)
	unspents := make([]explorer.Utxo, 0)

	for _, addr := range addresses {
		go s.getUtxosForAddress(addr, blindingKeys, chUnspents, chErr)

		select {
		case err := <-chErr:
			close(chErr)
			close(chUnspents)
			return nil, err
		case unspentsForAddress := <-chUnspents:
			unspents = append(unspents, unspentsForAddress...)
		}
	}

	return unspents, nil
}

func (s *Service) getUtxosForAddress(addr string, blindingKeys [][]byte, chUnspents chan []explorer.Utxo, chErr chan error) {
	unspents, err := s.explorerService.GetUnSpents(addr, blindingKeys)
	if err != nil {
		chErr <- err
		return
	}
	chUnspents <- unspents
}
