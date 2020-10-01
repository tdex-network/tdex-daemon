package explorer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTxFromJSON(t *testing.T) {
	tests := []struct {
		tx         string
		hash       string
		version    int
		locktime   int
		numInputs  int
		numOutputs int
		size       int
		weight     int
		fee        int
		confirmed  bool
	}{
		{
			tx:         `{"txid":"80008ac81a6b7999101912d263108906fd138fdbb01a5a291fff16297c0cd4be","version":2,"locktime":102,"vin":[{"txid":"8b902c36a0acdc699f33f98a06746d8036a42dd5d490acb81bcc95b5a47ae136","vout":0,"prevout":{"scriptpubkey":"a914dcb88f37e229a69076b9222dac623a411a7017e787","scriptpubkey_asm":"OP_HASH160 OP_PUSHBYTES_20 dcb88f37e229a69076b9222dac623a411a7017e7 OP_EQUAL","scriptpubkey_type":"p2sh","scriptpubkey_address":"XXUJbmC2gCbKGNz9Ck6bKtTUEPesk6rsxp","valuecommitment":"094f65e14988e0395ecb185213986ebb788706879b19c8455aee148aa090ff570d","assetcommitment":"0a04caa2c82d32fc177ba9ae7c883d195043b5f8d8f4c54f28c49073686e6f42ec"},"scriptsig":"160014f03e88e8f028c2cbf3070c1a25dfa5ba24ceb317","scriptsig_asm":"OP_PUSHBYTES_22 0014f03e88e8f028c2cbf3070c1a25dfa5ba24ceb317","witness":["3044022057e34780e1df2538608926c9fd09336e50a221c70a36bb56ab0ecf788fbda6520220756ea488da91f903fdf9cfefdc610988c036598ba7a51ef80d8f6a931b6db05201","0312452c17f38c85c1155fcd38af63eae517e33442f7962b9f3d10b5dfcece1199"],"is_coinbase":false,"sequence":4294967293,"inner_redeemscript_asm":"OP_0 OP_PUSHBYTES_20 f03e88e8f028c2cbf3070c1a25dfa5ba24ceb317","is_pegin":false}],"vout":[{"scriptpubkey":"a91485fe9e652da065eb94b0a397974fa8f1a6518cac87","scriptpubkey_asm":"OP_HASH160 OP_PUSHBYTES_20 85fe9e652da065eb94b0a397974fa8f1a6518cac OP_EQUAL","scriptpubkey_type":"p2sh","scriptpubkey_address":"XPZjjssao7v3oDSz6KoNd4MNuEwdwMh1GD","valuecommitment":"0892eff1656a04b2457e5b2ca7f9a4758634dd1e1b43fc41196176bdfcf60a1122","assetcommitment":"0a28434da13aeb8efc5f4f4cf7e51735087c9d586fa16f472968e3b5c6442a5265"},{"scriptpubkey":"76a9143e1632a93247c2c304e3f2ef1dd0579562168e6b88ac","scriptpubkey_asm":"OP_DUP OP_HASH160 OP_PUSHBYTES_20 3e1632a93247c2c304e3f2ef1dd0579562168e6b OP_EQUALVERIFY OP_CHECKSIG","scriptpubkey_type":"p2pkh","scriptpubkey_address":"2df62skQ9ojniFjRg5NQZQbHwKQivcxYsDZ","value":100000000,"asset":"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"},{"scriptpubkey":"","scriptpubkey_asm":"","scriptpubkey_type":"fee","value":2802,"asset":"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"}],"size":4670,"weight":5603,"fee":2802,"status":{"confirmed":false}}`,
			hash:       "80008ac81a6b7999101912d263108906fd138fdbb01a5a291fff16297c0cd4be",
			version:    2,
			locktime:   102,
			numInputs:  1,
			numOutputs: 3,
			size:       4670,
			weight:     5603,
			fee:        2802,
			confirmed:  false,
		},
		{
			tx:         `{"txid":"8b902c36a0acdc699f33f98a06746d8036a42dd5d490acb81bcc95b5a47ae136","version":2,"locktime":101,"vin":[{"txid":"12a51d3e6441abbdb3a5b35983260dec3ff457a301a24a4e76d41b0c0adfb99b","vout":0,"prevout":{"scriptpubkey":"51","scriptpubkey_asm":"OP_PUSHNUM_1","scriptpubkey_type":"unknown","value":2100000000000000,"asset":"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"},"scriptsig":"","scriptsig_asm":"","is_coinbase":false,"sequence":4294967293,"is_pegin":false}],"vout":[{"scriptpubkey":"a914dcb88f37e229a69076b9222dac623a411a7017e787","scriptpubkey_asm":"OP_HASH160 OP_PUSHBYTES_20 dcb88f37e229a69076b9222dac623a411a7017e7 OP_EQUAL","scriptpubkey_type":"p2sh","scriptpubkey_address":"XXUJbmC2gCbKGNz9Ck6bKtTUEPesk6rsxp","valuecommitment":"094f65e14988e0395ecb185213986ebb788706879b19c8455aee148aa090ff570d","assetcommitment":"0a04caa2c82d32fc177ba9ae7c883d195043b5f8d8f4c54f28c49073686e6f42ec"},{"scriptpubkey":"a91477c2a1230defd10f1951f12f54dcd2b9d9c4b5a087","scriptpubkey_asm":"OP_HASH160 OP_PUSHBYTES_20 77c2a1230defd10f1951f12f54dcd2b9d9c4b5a0 OP_EQUAL","scriptpubkey_type":"p2sh","scriptpubkey_address":"XNGUR4dKyZmQXyZfJ6t7phu94SnR1iY1Jx","valuecommitment":"085857b0fee41980e61ec576614b8e3f353695b1810640783cc5ed27ed617fe2d4","assetcommitment":"0be5cfedebb40a094a2a59a1a0465f0aa71aa889dba75c79e69aa1f4bc4ab7a9cd"},{"scriptpubkey":"","scriptpubkey_asm":"","scriptpubkey_type":"fee","value":4932,"asset":"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"}],"size":8838,"weight":9864,"fee":4932,"status":{"confirmed":true,"block_height":102,"block_hash":"6cfd770948cd3b860cbc9fabf104257fde5cceaf6c20d0fbe3ab71d8d2cf902b","block_time":1601550776}}`,
			hash:       "8b902c36a0acdc699f33f98a06746d8036a42dd5d490acb81bcc95b5a47ae136",
			version:    2,
			locktime:   101,
			numInputs:  1,
			numOutputs: 3,
			size:       8838,
			weight:     9864,
			fee:        4932,
			confirmed:  true,
		},
		{
			tx:         `{"txid":"a90bf01b0f96b8d58fd14c07e19eb9c587de043de249b19efd24e3d133a69fca","version":2,"locktime":0,"vin":[{"txid":"6e4da106a2f687f987729e3ce7df0c956dcf739d34b26b67c66106ba807db017","vout":1,"prevout":{"scriptpubkey":"76a9147a98c4960c517f551799b1788c6c009aa877398b88ac","scriptpubkey_asm":"OP_DUP OP_HASH160 OP_PUSHBYTES_20 7a98c4960c517f551799b1788c6c009aa877398b OP_EQUALVERIFY OP_CHECKSIG","scriptpubkey_type":"p2pkh","scriptpubkey_address":"2dkbyqRV6RCyWGb5en1pQX2Rr3gvvPt12zi","value":100000000,"asset":"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"},"scriptsig":"473044022061f578cf4fee29b9bfdaf4ac8704ccf856906846cbd0976acbbe8a11663bc827022008ed46133d256ea3f17b41151675897ae9696187097a097f63cba5aaf461b703012102798e7acd42966afa198fad764cfe581e7279e69138c6a9c6d5ed30b138adb50b","scriptsig_asm":"OP_PUSHBYTES_71 3044022061f578cf4fee29b9bfdaf4ac8704ccf856906846cbd0976acbbe8a11663bc827022008ed46133d256ea3f17b41151675897ae9696187097a097f63cba5aaf461b70301 OP_PUSHBYTES_33 02798e7acd42966afa198fad764cfe581e7279e69138c6a9c6d5ed30b138adb50b","is_coinbase":false,"sequence":4294967295,"is_pegin":false}],"vout":[{"scriptpubkey":"0014f10870e5c8fbff446f7567e52be333a1509d3541","scriptpubkey_asm":"OP_0 OP_PUSHBYTES_20 f10870e5c8fbff446f7567e52be333a1509d3541","scriptpubkey_type":"v0_p2wpkh","scriptpubkey_address":"ert1q7yy8pewgl0l5gmm4vljjhcen59gf6d2psfymt4","value":99999500,"asset":"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"},{"scriptpubkey":"","scriptpubkey_asm":"","scriptpubkey_type":"fee","value":500,"asset":"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"}],"size":268,"weight":1072,"fee":500,"status":{"confirmed":true,"block_height":106,"block_hash":"b9f3bff002e61d1bbeceb706849abed8d3d68ff2326f497671108a16297a6398","block_time":1601560260}}`,
			hash:       "a90bf01b0f96b8d58fd14c07e19eb9c587de043de249b19efd24e3d133a69fca",
			version:    2,
			locktime:   0,
			numInputs:  1,
			numOutputs: 2,
			size:       268,
			weight:     1072,
			fee:        500,
			confirmed:  true,
		},
	}
	for _, tt := range tests {
		trx, err := NewTxFromJSON(tt.tx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, tt.hash, trx.Hash())
		assert.Equal(t, tt.version, trx.Version())
		assert.Equal(t, tt.locktime, trx.Locktime())
		assert.Equal(t, tt.numInputs, len(trx.Inputs()))
		assert.Equal(t, tt.numOutputs, len(trx.Outputs()))
		assert.Equal(t, tt.size, trx.Size())
		assert.Equal(t, tt.weight, trx.Weight())
		assert.Equal(t, tt.fee, trx.Fee())
		assert.Equal(t, tt.confirmed, trx.Confirmed())
	}
}
