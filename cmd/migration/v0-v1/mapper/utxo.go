package mapper

import (
	"fmt"
	"time"

	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0-v1/v1-domain"
	v0domain "github.com/tdex-network/tdex-daemon/old-v0"
)

func (m *mapperService) FromV0UnspentsToV1Utxos(
	utxos []*v0domain.Unspent,
) ([]*v1domain.Utxo, error) {
	res := make([]*v1domain.Utxo, 0, len(utxos))
	for _, v := range utxos {
		utxo, err := m.fromV0UnspentToV1Utxo(*v)
		if err != nil {
			return nil, err
		}

		if utxo != nil {
			res = append(res, utxo)
		}
	}

	return res, nil
}

func (m *mapperService) fromV0UnspentToV1Utxo(
	utxo v0domain.Unspent,
) (*v1domain.Utxo, error) {
	if !utxo.Confirmed && !utxo.Spent {
		return nil, nil
	}

	_, accountIndex, err := m.v0RepoManager.GetVaultRepository().
		GetAccountByAddress(utxo.Address)
	if err != nil {
		return nil, err
	}

	accountName := fmt.Sprintf(nameSpaceFormat, accountIndex)

	lockTimestamp := int64(0)
	LockExpiryTimestamp := int64(0)
	if utxo.IsLocked() {
		lockTimestamp = time.Now().Unix()
		LockExpiryTimestamp = time.Now().Add(time.Minute).Unix()
	}

	confirmedStatus := v1domain.UtxoStatus{}
	if utxo.Confirmed {
		confirmedStatus = v1domain.UtxoStatus{BlockHeight: 1}
	}

	spentStatus := v1domain.UtxoStatus{}
	if utxo.Spent {
		spentStatus = v1domain.UtxoStatus{BlockHeight: 1}
	}

	return &v1domain.Utxo{
		UtxoKey: v1domain.UtxoKey{
			TxID: utxo.TxID,
			VOut: utxo.VOut,
		},
		Value:               utxo.Value,
		Asset:               utxo.AssetHash,
		ValueCommitment:     []byte(utxo.ValueCommitment),
		AssetCommitment:     []byte(utxo.AssetCommitment),
		ValueBlinder:        utxo.ValueBlinder,
		AssetBlinder:        utxo.AssetBlinder,
		Script:              utxo.ScriptPubKey,
		Nonce:               utxo.Nonce,
		RangeProof:          utxo.RangeProof,
		SurjectionProof:     utxo.SurjectionProof,
		AccountName:         accountName,
		LockTimestamp:       lockTimestamp,
		LockExpiryTimestamp: LockExpiryTimestamp,
		ConfirmedStatus:     confirmedStatus,
		SpentStatus:         spentStatus,
	}, nil
}
