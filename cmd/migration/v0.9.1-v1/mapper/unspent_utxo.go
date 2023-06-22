package mapper

import (
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
		res = append(res, unspent)
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

	market, err := m.v091RepoManager.MarketRepository().GetMarketByAccount(accountIndex)
	if err != nil {
		return nil, err
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
		AccountName:         market.AccountName(),
		LockTimestamp:       0,                     //TODO
		LockExpiryTimestamp: 0,                     //TODO
		SpentStatus:         v1domain.UtxoStatus{}, //TODO
		ConfirmedStatus:     v1domain.UtxoStatus{}, //TODO
	}, nil
}
