package wallet

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/psetv2"
)

type Service struct {
	wallet     ports.WalletService
	staticInfo ports.WalletInfo

	txNotificationHandlers   txNotificationQueue
	utxoNotificationHandlers utxoNotificationQueue
}

func NewService(wallet ports.WalletService) (*Service, error) {
	if wallet == nil {
		return nil, fmt.Errorf("missing ocean wallet service")
	}

	info, err := wallet.Wallet().Info(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ocean wallet: %s", err)
	}
	txHandlers := txNotificationQueue{
		&sync.Mutex{}, make([]func(ports.WalletTxNotification) bool, 0),
	}
	utxoHandlers := utxoNotificationQueue{
		&sync.Mutex{}, make([]func(ports.WalletUtxoNotification) bool, 0),
	}

	svc := &Service{wallet, info, txHandlers, utxoHandlers}
	go svc.listenToTxNotifications()
	go svc.listenToUtxoNotifications()

	return svc, nil
}

func (s *Service) Wallet() ports.Wallet {
	return s.wallet.Wallet()
}

func (s *Service) Account() ports.Account {
	return s.wallet.Account()
}

func (s *Service) Transaction() ports.Transaction {
	return s.wallet.Transaction()
}

func (s *Service) Notification() ports.Notification {
	return s.wallet.Notification()
}

func (s *Service) Network() network.Network {
	net := network.Liquid
	if s.staticInfo.GetNetwork() == network.Testnet.Name {
		net = network.Testnet
	}
	if s.staticInfo.GetNetwork() == network.Regtest.Name {
		net = network.Regtest
	}
	net.AssetID = s.staticInfo.GetNativeAsset()
	return net
}

func (s *Service) NativeAsset() string {
	return s.staticInfo.GetNativeAsset()
}

func (s *Service) SendToMany(
	account string, outs []ports.TxOutput, msatsPerByte uint64,
) (string, error) {
	ctx := context.Background()
	txManager := s.wallet.Transaction()
	accountManager := s.wallet.Account()
	changeAmountPerAsset := make(map[string]uint64)
	inputs := make([]ports.TxInput, 0)
	outputs := append([]ports.TxOutput{}, outs...)

	for asset, amount := range totOutputAmountPerAsset(outs) {
		utxos, change, _, err := txManager.SelectUtxos(ctx, account, asset, amount)
		if err != nil {
			return "", err
		}
		if change > 0 {
			changeAmountPerAsset[asset] = change
		}
		for _, u := range utxos {
			txid, _ := elementsutil.TxIDToBytes(u.GetTxid())
			var scriptSigSize, witnessSize int
			if len(u.GetRedeemScript()) > 0 {
				scriptSigSize = 35
				witnessSize = 223
			}
			inputs = append(inputs, input{
				txid, u.GetIndex(), u.GetScript(), scriptSigSize, witnessSize,
			})
		}
	}

	if numOfAddress := len(changeAmountPerAsset); numOfAddress > 0 {
		addresses, err := accountManager.DeriveChangeAddresses(
			ctx, account, numOfAddress,
		)
		if err != nil {
			return "", err
		}

		i := 0
		for asset, amount := range changeAmountPerAsset {
			info, _ := address.FromConfidential(addresses[i])
			outputs = append(outputs, output{
				asset, amount, hex.EncodeToString(info.Script),
				hex.EncodeToString(info.BlindingKey),
			})
			i++
		}
	}

	dummyFeeAmount, err := txManager.EstimateFees(
		ctx, inputs, outputs, msatsPerByte,
	)
	if err != nil {
		return "", err
	}
	// 150 is an over estimation of an extra confidential output (change).
	dummyFeeAmount += 150
	lbtc := s.staticInfo.GetNativeAsset()
	feeUtxos, change, _, err := txManager.SelectUtxos(
		ctx, domain.FeeAccount, lbtc, dummyFeeAmount,
	)
	if err != nil {
		return "", err
	}

	for _, u := range feeUtxos {
		txid, _ := elementsutil.TxIDToBytes(u.GetTxid())
		inputs = append(inputs, input{txid, u.GetIndex(), u.GetScript(), 0, 0})
	}
	feeAmount := dummyFeeAmount
	if change > 0 {
		addresses, err := accountManager.DeriveAddresses(ctx, domain.FeeAccount, 1)
		if err != nil {
			return "", err
		}
		info, _ := address.FromConfidential(addresses[0])
		outputs = append(outputs, output{
			lbtc, change, hex.EncodeToString(info.Script),
			hex.EncodeToString(info.BlindingKey),
		})
		feeAmount, err = txManager.EstimateFees(
			ctx, inputs, outputs, msatsPerByte,
		)
		if err != nil {
			return "", err
		}

		changeOut := outputs[len(outputs)-1]
		changeOutScript := changeOut.(output).script
		changeOutBlindKey := changeOut.(output).blindKey
		diff := dummyFeeAmount - feeAmount
		amount := changeOut.GetAmount() + diff
		outputs[len(outputs)-1] = output{
			changeOut.GetAsset(), amount, changeOutScript, changeOutBlindKey,
		}
	}

	outputs = append(outputs, output{lbtc, feeAmount, "", ""})

	pset, err := txManager.CreatePset(ctx, inputs, outputs)
	if err != nil {
		return "", err
	}

	blindedPset, err := txManager.BlindPset(ctx, pset, nil)
	if err != nil {
		return "", err
	}
	txhex, err := txManager.SignPset(ctx, blindedPset, true)
	if err != nil {
		return "", err
	}
	txid, err := txManager.BroadcastTransaction(ctx, txhex)
	if err != nil {
		return "", err
	}
	return txid, nil
}

func (s *Service) CompleteSwap(
	account string, swapRequest ports.SwapRequest, msatsPerByte uint64,
	feesToAdd bool,
) (string, []ports.Utxo, int64, error) {
	ctx := context.Background()
	txManager := s.wallet.Transaction()
	accountManager := s.wallet.Account()
	inputs := make([]ports.TxInput, 0)
	existingInputs := make([]ports.TxInput, 0)
	existingOutputs := make([]ports.TxOutput, 0)

	ptx, _ := psetv2.NewPsetFromBase64(swapRequest.GetTransaction())
	for _, in := range ptx.Inputs {
		var scriptSigSize, witnessSize int
		if len(in.RedeemScript) > 0 {
			// values for 2of2 native bare multisig inputs
			scriptSigSize = 223
		}
		if len(in.WitnessScript) > 0 {
			// values for 2of2 native or wrapped segwit multisig inputs
			if scriptSigSize > 0 {
				scriptSigSize = 35
			}
			witnessSize = 223
		}
		existingInputs = append(existingInputs, input{
			in.PreviousTxid, in.PreviousTxIndex, hex.EncodeToString(in.GetUtxo().Script),
			scriptSigSize, witnessSize,
		})
	}
	for _, out := range ptx.Outputs {
		existingOutputs = append(existingOutputs, output{
			hex.EncodeToString(elementsutil.ReverseBytes(out.Asset)),
			out.Value, hex.EncodeToString(out.Script), hex.EncodeToString(out.BlindingPubkey),
		})
	}

	amountR := swapRequest.GetAmountR()
	if swapRequest.GetFeeAsset() == swapRequest.GetAssetR() && !feesToAdd {
		amountR -= swapRequest.GetFeeAmount()
	}

	utxos, change, unlockTime, err := txManager.SelectUtxos(
		ctx, account, swapRequest.GetAssetR(), amountR,
	)
	if err != nil {
		return "", nil, -1, err
	}

	for _, u := range utxos {
		txid, _ := elementsutil.TxIDToBytes(u.GetTxid())
		var scriptSigSize, witnessSize int
		if len(u.GetRedeemScript()) > 0 {
			scriptSigSize = 35
			witnessSize = 223
		}
		inputs = append(inputs, input{
			txid, u.GetIndex(), u.GetScript(), scriptSigSize, witnessSize,
		})
	}

	addresses, err := accountManager.DeriveAddresses(ctx, account, 1)
	if err != nil {
		return "", nil, -1, err
	}
	info, _ := address.FromConfidential(addresses[0])
	amountP := swapRequest.GetAmountP()
	if swapRequest.GetFeeAsset() == swapRequest.GetAssetP() && feesToAdd {
		amountP += swapRequest.GetFeeAmount()
	}
	outputs := []ports.TxOutput{
		output{
			swapRequest.GetAssetP(), amountP,
			hex.EncodeToString(info.Script), hex.EncodeToString(info.BlindingKey),
		},
	}
	if change > 0 {
		addresses, err := accountManager.DeriveChangeAddresses(ctx, account, 1)
		if err != nil {
			return "", nil, -1, err
		}
		info, _ := address.FromConfidential(addresses[0])
		outputs = append(outputs, output{
			swapRequest.GetAssetR(), change, hex.EncodeToString(info.Script),
			hex.EncodeToString(info.BlindingKey),
		})
	}

	allInputs := append(existingInputs, inputs...)
	allOutputs := append(existingOutputs, outputs...)
	dummyFeeAmount, err := txManager.EstimateFees(
		ctx, allInputs, allOutputs, msatsPerByte,
	)
	if err != nil {
		return "", nil, -1, err
	}

	// 150 is an over estimation of an extra confidential output (change).
	dummyFeeAmount += 150
	lbtc := s.staticInfo.GetNativeAsset()
	feeUtxos, change, _, err := txManager.SelectUtxos(
		ctx, domain.FeeAccount, lbtc, dummyFeeAmount,
	)
	if err != nil {
		return "", nil, -1, err
	}

	for _, u := range feeUtxos {
		txid, _ := elementsutil.TxIDToBytes(u.GetTxid())
		inputs = append(inputs, input{txid, u.GetIndex(), u.GetScript(), 0, 0})
	}
	feeAmount := dummyFeeAmount
	if change > 0 {
		addresses, err := accountManager.DeriveChangeAddresses(
			ctx, domain.FeeAccount, 1,
		)
		if err != nil {
			return "", nil, -1, err
		}
		info, _ := address.FromConfidential(addresses[0])
		outputs = append(outputs, output{
			lbtc, change, hex.EncodeToString(info.Script),
			hex.EncodeToString(info.BlindingKey),
		})

		allInputs := append(existingInputs, inputs...)
		allOutputs := append(existingOutputs, outputs...)
		feeAmount, err = txManager.EstimateFees(
			ctx, allInputs, allOutputs, msatsPerByte,
		)
		if err != nil {
			return "", nil, -1, err
		}

		changeOut := outputs[len(outputs)-1]
		changeOutScript := changeOut.(output).script
		changeOutBlindKey := changeOut.(output).blindKey
		diff := dummyFeeAmount - feeAmount
		amount := changeOut.GetAmount() + diff
		outputs[len(outputs)-1] = output{
			changeOut.GetAsset(), amount, changeOutScript, changeOutBlindKey,
		}
	}

	outputs = append(outputs, output{lbtc, feeAmount, "", ""})

	pset, err := txManager.UpdatePset(
		ctx, swapRequest.GetTransaction(), inputs, outputs,
	)
	if err != nil {
		return "", nil, -1, err
	}

	blindedPset, err := txManager.BlindPset(
		ctx, pset, swapRequest.GetUnblindedInputs(),
	)
	if err != nil {
		return "", nil, -1, err
	}

	signedPset, err := txManager.SignPset(ctx, blindedPset, false)
	if err != nil {
		return "", nil, -1, err
	}

	utxos = append(utxos, feeUtxos...)

	return signedPset, utxos, unlockTime, nil
}

func (s *Service) RegisterHandlerForTxEvent(
	handler func(ports.WalletTxNotification) bool,
) {
	s.txNotificationHandlers.pushBack(handler)
}

func (s *Service) RegisterHandlerForUtxoEvent(
	handler func(ports.WalletUtxoNotification) bool,
) {
	s.utxoNotificationHandlers.pushBack(handler)
}

func (s *Service) Close() {
	s.wallet.Close()
}

func (s *Service) listenToTxNotifications() {
	for notification := range s.wallet.Notification().GetTxNotifications() {
		toRepeat := make([]func(ports.WalletTxNotification) bool, 0)
		numHandlers := s.txNotificationHandlers.len()
		for i := 0; i < numHandlers; i++ {
			handler := s.txNotificationHandlers.pop()
			done := handler(notification)
			if !done {
				toRepeat = append(toRepeat, handler)
			}
		}
		for _, handler := range toRepeat {
			s.txNotificationHandlers.pushBack(handler)
		}
	}
}

func (s *Service) listenToUtxoNotifications() {
	for notification := range s.wallet.Notification().GetUtxoNotifications() {
		toRepeat := make([]func(ports.WalletUtxoNotification) bool, 0)
		numHandlers := s.utxoNotificationHandlers.len()
		for i := 0; i < numHandlers; i++ {
			handler := s.utxoNotificationHandlers.pop()
			done := handler(notification)
			if !done {
				toRepeat = append(toRepeat, handler)
			}
		}
		for _, handler := range toRepeat {
			s.utxoNotificationHandlers.pushBack(handler)
		}
	}
}
