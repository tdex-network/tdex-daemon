package mapper

import (
	"time"

	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"
	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-domain"
)

func (m *mapperService) FromV091UnspentsToV1Utxos(
	unspents []*v091domain.Unspent,
) ([]*v1domain.Utxo, error) {
	res := make([]*v1domain.Utxo, 0, len(unspents))
	for _, v := range unspents {
		unspent, err := m.fromV091UnspentToV1Utxo(v)
		if err != nil {
			return nil, err
		}

		if unspent != nil {
			res = append(res, unspent)
		}
	}

	return res, nil
}

func (m *mapperService) fromV091UnspentToV1Utxo(
	unspent *v091domain.Unspent,
) (*v1domain.Utxo, error) {
	_, accountIndex, err := m.v091RepoManager.GetVaultRepository().
		GetAccountByAddress(unspent.Address)
	if err != nil {
		return nil, err
	}

	accountName, err := m.getLabel(accountIndex)
	if err != nil {
		return nil, err
	}

	lockTimestamp := int64(0)
	LockExpiryTimestamp := int64(0)
	if unspent.IsLocked() {
		lockTimestamp = time.Now().Unix()
		LockExpiryTimestamp = time.Now().Add(time.Minute).Unix()
	}

	spentStatus := v1domain.UtxoStatus{}
	confirmedStatus := v1domain.UtxoStatus{}
	if unspent.Spent {
		return nil, nil
	}

	return &v1domain.Utxo{
		UtxoKey: v1domain.UtxoKey{
			TxID: unspent.TxID,
			VOut: unspent.VOut,
		},
		Value:               unspent.Value,
		Asset:               unspent.AssetHash,
		ValueCommitment:     []byte(unspent.ValueCommitment),
		AssetCommitment:     []byte(unspent.AssetCommitment),
		ValueBlinder:        unspent.ValueBlinder,
		AssetBlinder:        unspent.AssetBlinder,
		Script:              unspent.ScriptPubKey,
		Nonce:               unspent.Nonce,
		RangeProof:          unspent.RangeProof,
		SurjectionProof:     unspent.SurjectionProof,
		AccountName:         accountName,
		LockTimestamp:       lockTimestamp,
		LockExpiryTimestamp: LockExpiryTimestamp,
		SpentStatus:         spentStatus,
		ConfirmedStatus:     confirmedStatus,
	}, nil
}
