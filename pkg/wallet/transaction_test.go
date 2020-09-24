package wallet

import (
	"encoding/hex"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
)

func TestCreateAndUpdateSwapTx(t *testing.T) {
	wallet, err := NewWalletFromMnemonic(NewWalletFromMnemonicOpts{
		SigningMnemonic:  strings.Split("quarter multiply swarm depth slice security flight glad arrow express worth legend wasp mobile anchor dinner mutual six sure wear section delay initial thank", " "),
		BlindingMnemonic: strings.Split("okay door hammer betray reason zero fiction rigid vivid scorpion thunder crucial focus riot cancel wear autumn rely kangaroo rug raven mystery ability stem", " "),
	})
	if err != nil {
		t.Fatal(err)
	}

	psetBase64, err := wallet.CreateTx()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, len(psetBase64) > 0)

	ptx, _ := pset.NewPsetFromBase64(psetBase64)
	assert.Equal(t, 0, len(ptx.Inputs))
	assert.Equal(t, 0, len(ptx.Outputs))

	opts := UpdateSwapTxOpts{
		PsetBase64:           psetBase64,
		Unspents:             mockUnspents(),
		InputAmount:          60000000,
		InputAsset:           "1adcc1e8564a6f01c957a0f7fcb8badce9c126d790550e6d6817aa752369ae5f",
		OutputAmount:         100000000000,
		OutputAsset:          network.Regtest.AssetID,
		OutputDerivationPath: "0'/0/1",
		ChangeDerivationPath: "0'/1/0",
		Network:              &network.Regtest,
	}

	updatedPsetBase64, selectedUnspents, err := wallet.UpdateSwapTx(opts)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, len(updatedPsetBase64) > 0)
	assert.Equal(t, 1, len(selectedUnspents))

	ptx, _ = pset.NewPsetFromBase64(updatedPsetBase64)
	assert.Equal(t, 1, len(ptx.Inputs))
	assert.Equal(t, 2, len(ptx.Outputs))
}

func TestFailingUpdateSwapTx(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}

	psetBase64, err := wallet.CreateTx()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		opts UpdateSwapTxOpts
		err  error
	}{
		{
			opts: UpdateSwapTxOpts{
				PsetBase64:           "",
				Unspents:             mockUnspents(),
				InputAmount:          6000000000,
				InputAsset:           "1adcc1e8564a6f01c957a0f7fcb8badce9c126d790550e6d6817aa752369ae5f",
				OutputAmount:         100000000000,
				OutputAsset:          network.Regtest.AssetID,
				OutputDerivationPath: "0'/0/1",
				ChangeDerivationPath: "0'/1/0",
				Network:              &network.Regtest,
			},
			err: ErrNullPset,
		},
		{
			opts: UpdateSwapTxOpts{
				PsetBase64:           psetBase64,
				Unspents:             []explorer.Utxo{},
				InputAmount:          6000000000,
				InputAsset:           "1adcc1e8564a6f01c957a0f7fcb8badce9c126d790550e6d6817aa752369ae5f",
				OutputAmount:         100000000000,
				OutputAsset:          network.Regtest.AssetID,
				OutputDerivationPath: "0'/0/1",
				ChangeDerivationPath: "0'/1/0",
				Network:              &network.Regtest,
			},
			err: ErrEmptyUnspents,
		},
		{
			opts: UpdateSwapTxOpts{
				PsetBase64:           psetBase64,
				Unspents:             mockUnspents(),
				InputAmount:          0,
				InputAsset:           "1adcc1e8564a6f01c957a0f7fcb8badce9c126d790550e6d6817aa752369ae5f",
				OutputAmount:         100000000000,
				OutputAsset:          network.Regtest.AssetID,
				OutputDerivationPath: "0'/0/1",
				ChangeDerivationPath: "0'/1/0",
				Network:              &network.Regtest,
			},
			err: ErrZeroInputAmount,
		},
		{
			opts: UpdateSwapTxOpts{
				PsetBase64:           psetBase64,
				Unspents:             mockUnspents(),
				InputAmount:          6000000000,
				InputAsset:           "",
				OutputAmount:         100000000000,
				OutputAsset:          network.Regtest.AssetID,
				OutputDerivationPath: "0'/0/1",
				ChangeDerivationPath: "0'/1/0",
				Network:              &network.Regtest,
			},
			err: ErrInvalidInputAsset,
		},
		{
			opts: UpdateSwapTxOpts{
				PsetBase64:           psetBase64,
				Unspents:             mockUnspents(),
				InputAmount:          6000000000,
				InputAsset:           "1adcc1e8564a6f01c957a0f7fcb8badce9c126d790550e6d6817aa752369ae5f",
				OutputAmount:         0,
				OutputAsset:          network.Regtest.AssetID,
				OutputDerivationPath: "0'/0/1",
				ChangeDerivationPath: "0'/1/0",
				Network:              &network.Regtest,
			},
			err: ErrZeroOutputAmount,
		},
		{
			opts: UpdateSwapTxOpts{
				PsetBase64:           psetBase64,
				Unspents:             mockUnspents(),
				InputAmount:          6000000000,
				InputAsset:           "1adcc1e8564a6f01c957a0f7fcb8badce9c126d790550e6d6817aa752369ae5f",
				OutputAmount:         100000000000,
				OutputAsset:          "",
				OutputDerivationPath: "0'/0/1",
				ChangeDerivationPath: "0'/1/0",
				Network:              &network.Regtest,
			},
			err: ErrInvalidOutputAsset,
		},
		{
			opts: UpdateSwapTxOpts{
				PsetBase64:           psetBase64,
				Unspents:             mockUnspents(),
				InputAmount:          6000000000,
				InputAsset:           "1adcc1e8564a6f01c957a0f7fcb8badce9c126d790550e6d6817aa752369ae5f",
				OutputAmount:         100000000000,
				OutputAsset:          network.Regtest.AssetID,
				OutputDerivationPath: "",
				ChangeDerivationPath: "0'/1/0",
				Network:              &network.Regtest,
			},
			err: ErrNullOutputDerivationPath,
		},
		{
			opts: UpdateSwapTxOpts{
				PsetBase64:           psetBase64,
				Unspents:             mockUnspents(),
				InputAmount:          6000000000,
				InputAsset:           "1adcc1e8564a6f01c957a0f7fcb8badce9c126d790550e6d6817aa752369ae5f",
				OutputAmount:         100000000000,
				OutputAsset:          network.Regtest.AssetID,
				OutputDerivationPath: "0'/0/1",
				ChangeDerivationPath: "",
				Network:              &network.Regtest,
			},
			err: ErrNullChangeDerivationPath,
		},
		{
			opts: UpdateSwapTxOpts{
				PsetBase64:           psetBase64,
				Unspents:             mockUnspents(),
				InputAmount:          6000000000,
				InputAsset:           "1adcc1e8564a6f01c957a0f7fcb8badce9c126d790550e6d6817aa752369ae5f",
				OutputAmount:         100000000000,
				OutputAsset:          network.Regtest.AssetID,
				OutputDerivationPath: "0'/0/1",
				ChangeDerivationPath: "0'/1/0",
			},
			err: ErrNullNetwork,
		},
	}

	for _, tt := range tests {
		_, _, err := wallet.UpdateSwapTx(tt.opts)
		assert.Equal(t, tt.err, err)
	}
}

func TestUpdateTx(t *testing.T) {
	wallet, err := NewWalletFromMnemonic(NewWalletFromMnemonicOpts{
		SigningMnemonic:  strings.Split("quarter multiply swarm depth slice security flight glad arrow express worth legend wasp mobile anchor dinner mutual six sure wear section delay initial thank", " "),
		BlindingMnemonic: strings.Split("okay door hammer betray reason zero fiction rigid vivid scorpion thunder crucial focus riot cancel wear autumn rely kangaroo rug raven mystery ability stem", " "),
	})
	if err != nil {
		t.Fatal(err)
	}

	psetBase64, err := wallet.CreateTx()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		unspents             []explorer.Utxo
		outputs              outputList
		changePathsByAsset   map[string]string
		expectedIns          int
		expectedOuts         int
		expectedBlindingKeys int
	}{
		{
			nil, // no unspents
			outputList{
				{
					network.Regtest.AssetID,
					1,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			nil, // no changed derivation map
			0,
			1, // outputs
			0,
		},
		{
			mockUnspentsForUpdateTx(),
			outputList{
				{
					network.Regtest.AssetID,
					1,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			map[string]string{
				network.Regtest.AssetID: "0'/1/1",
			},
			2,
			2, // outputs + LBTC change
			1,
		},
		{
			mockUnspentsForUpdateTx(),
			outputList{
				{
					"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c",
					1000,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			map[string]string{
				"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c": "0'/1/0",
				network.Regtest.AssetID: "0'/1/1",
			},
			2,
			2, // outputs + LBTC change
			1,
		},
		{
			mockUnspentsForUpdateTx(),
			outputList{
				{
					"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c",
					500,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			map[string]string{
				"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c": "0'/1/0",
				network.Regtest.AssetID: "0'/1/1",
			},
			2,
			3, // outputs + asset change + lbtc change
			2,
		},
		{
			mockUnspentsForUpdateTx(),
			outputList{
				{
					"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c",
					1000,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
				{
					network.Regtest.AssetID,
					0.5,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			map[string]string{
				"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c": "0'/1/0",
				network.Regtest.AssetID: "0'/1/1",
			},
			2,
			3, // outputs + lbtc change
			1,
		},
		{
			mockUnspentsForUpdateTx(),
			[]output{
				{
					"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c",
					400,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
				{
					"3be6cc6330799ea0a1ae2b7a950ba983e88f41b75a0cb36342e7a039903e7d55",
					300,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
				{
					network.Regtest.AssetID,
					0.5,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			map[string]string{
				"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c": "0'/1/0",
				"3be6cc6330799ea0a1ae2b7a950ba983e88f41b75a0cb36342e7a039903e7d55": "0'/1/1",
				network.Regtest.AssetID: "0'/1/2",
			},
			3,
			6, // outputs + asset1 change + asset2 change + lbtc change
			3,
		},
	}

	for _, tt := range tests {
		opts := UpdateTxOpts{
			PsetBase64:         psetBase64,
			Unspents:           tt.unspents,
			Outputs:            tt.outputs.TxOutputs(),
			ChangePathsByAsset: tt.changePathsByAsset,
			MilliSatsPerBytes:  100,
			Network:            &network.Regtest,
		}
		res, err := wallet.UpdateTx(opts)
		if err != nil {
			t.Fatal(err)
		}
		ptx, _ := pset.NewPsetFromBase64(res.PsetBase64)
		assert.Equal(t, tt.expectedIns, len(ptx.Inputs))
		assert.Equal(t, tt.expectedOuts, len(ptx.Outputs))
		assert.Equal(t, tt.expectedIns, len(res.SelectedUnspents))
		assert.Equal(t, tt.expectedBlindingKeys, len(res.ChangeOutputsBlindingKeys))
		if len(tt.unspents) > 0 {
			assert.Equal(t, true, res.FeeAmount > 0)
		} else {
			assert.Equal(t, uint64(0), res.FeeAmount)
		}
	}
}

func TestFailingUpdateTx(t *testing.T) {
	tests := []struct {
		unspents           []explorer.Utxo
		outputs            outputList
		changePathsByAsset map[string]string
		milliSatsPerByte   int
		network            *network.Network
		err                error
	}{
		{
			unspents: mockUnspentsForUpdateTx(),
			outputs: outputList{
				{
					"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c",
					500,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			changePathsByAsset: nil,
			milliSatsPerByte:   100,
			network:            &network.Regtest,
			err:                ErrNullChangePathsByAsset,
		},
		{
			unspents: mockUnspentsForUpdateTx(),
			outputs: outputList{
				{
					"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c",
					500,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			changePathsByAsset: map[string]string{
				"3be6cc6330799ea0a1ae2b7a950ba983e88f41b75a0cb36342e7a039903e7d55": "0'/1/0",
				network.Regtest.AssetID: "0'/1/1",
			},
			milliSatsPerByte: 100,
			network:          &network.Regtest,
			err:              nil,
		},
		{
			unspents: mockUnspentsForUpdateTx(),
			outputs: outputList{
				{
					"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c",
					500,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			changePathsByAsset: map[string]string{
				"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c": "0'/1/0",
			},
			milliSatsPerByte: 100,
			network:          &network.Regtest,
			err:              nil,
		},
		{
			unspents: mockUnspentsForUpdateTx(),
			outputs: outputList{
				{
					"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c",
					500,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			changePathsByAsset: map[string]string{
				"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c": "0'/1/0",
				network.Regtest.AssetID: "0'/1/1",
			},
			milliSatsPerByte: 50,
			network:          &network.Regtest,
			err:              ErrInvalidMilliSatsPerBytes,
		},
		{
			unspents: mockUnspentsForUpdateTx(),
			outputs: outputList{
				{
					"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c",
					500,
					"0014595a242dc9f345268b40cbe669e5d5f746301bb9",
				},
			},
			changePathsByAsset: map[string]string{
				"be54f05c6ec9e9b1886b862458e76cf9f32c0d99b73b980e7a5a700292bd1a2c": "0'/1/0",
				network.Regtest.AssetID: "0'/1/1",
			},
			milliSatsPerByte: 100,
			err:              ErrNullNetwork,
		},
	}

	wallet, err := NewWalletFromMnemonic(NewWalletFromMnemonicOpts{
		SigningMnemonic:  strings.Split("quarter multiply swarm depth slice security flight glad arrow express worth legend wasp mobile anchor dinner mutual six sure wear section delay initial thank", " "),
		BlindingMnemonic: strings.Split("okay door hammer betray reason zero fiction rigid vivid scorpion thunder crucial focus riot cancel wear autumn rely kangaroo rug raven mystery ability stem", " "),
	})
	if err != nil {
		t.Fatal(err)
	}
	psetBase64, err := wallet.CreateTx()
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		opts := UpdateTxOpts{
			PsetBase64:         psetBase64,
			Unspents:           tt.unspents,
			Outputs:            tt.outputs.TxOutputs(),
			ChangePathsByAsset: tt.changePathsByAsset,
			MilliSatsPerBytes:  tt.milliSatsPerByte,
			Network:            tt.network,
		}
		_, err := wallet.UpdateTx(opts)
		if tt.err != nil {
			assert.Equal(t, tt.err, err)
		}
	}
}

func mockUnspents() []explorer.Utxo {
	script, _ := hex.DecodeString("0014e12851aa15a3eea94587a6f11d818d1e7809764c")

	nonce0, _ := hex.DecodeString("02c75ccdfba5d722359b1f65686bfbf25e02dbb778823ca6979716510d6811f67e")
	surjectionproof0, _ := hex.DecodeString("010001788e56d63b93aa445e952a41d0090b04f4b48837ba3114188db4805bbcd651590a76c1e63e9638ebd9f03ae7adefabfddbf14d7a22b95a02c6e7c2097ef9a2b4")

	nonce1, _ := hex.DecodeString("02b24930f7fa092709117a73bec36dda942f5235c0cdcab76d48aca162e3eedb9c")
	surjectionproof1, _ := hex.DecodeString("02000397d6eff18da59df0723c5ac95900e4cbc7b460976ea5cf3454a7a9d98fd16b3381c928b5d75407e0fc6080928d0fc5e39bf1bbbadab7826694fa49da57f31bb97411cf37a17b2986dcb09f2541d8ed98d926e757fabb41a78752229aac63590a")

	rangeproof0, _ := hex.DecodeString("603300000000000000019e89d000c40cc1e3b0e7facd27a916295b4c8138b8e9f99daaa7efb9e0549743c036d6693928db88d903e4f95f9c345947e6126fc8dfa1c049c849b86f3281b8f2c7ddcd59cfccaa686a678ce3f25e2d3f5a083d1c577608a2fb2f61787a84d235efb3b0ec32fad96e9c8e2fc2b1e882393f48b84c7a1ecf70fe2a6a62c54ed7218372111ca9136911d7ff64dff7ce5ec8ac0aa550b460d86930a72dd0378d33f6a0d402788a57d25b23af510c9e25f52014103192d50ffe61673ea420565f479c1c21183ea99886f2ddbee8320d6cbdfdad298c3d89da55204c3ccc18c31c4c51e4c2e7e3588863e1ab1a1efac75d4c2618081d47295cf3aa35f091f05d6a4651b0dd098ab1ba65ba005db467e15c87d9aa869375d655c9dd5cc5bf45f46bc4de6aa46dfbc4e6e8aab842cc14e9534b902af1ef5b57507678b8a45c3d6d3f0d537ac16cefc284f9e63e67e014c75b712b10a257fb4e5dcf1e12109aace6852b63f62f58f9aa025e3a1c8c9b0130d3496e022b4e9b3b00f1f5b49f5e7dff3e610aa0b1c1f6dc35e27c9753f2e8d33a06cc3658d7d2cf78e7003fbac30a9e3e14f89ea981100bbd5a028618c2cf76376991a3e7e1fee5a11200e0c7296e417ba26597d2c3a0ba6aa7a49042a6ce66a73f02192b521925695742b32e40a0f18c3fb03faf10ca7ad83f3033207ccc7c96e95fe7bc988b24bf61263d1f94b6ce1a58b70e85e56de92a4d37f5a70def4437ed3d5fcdd64286baf292bac515b610f75a88a1dabe24e6ce2b1ddeb28968adbc06c9070e267c7ae297d3a327a691dea431ae91673063ae9a754934e54ada49628bc4206183aafef6006300cad542cb3a9bfe8a2b8bd79071f10254ae2c3d0388b94c2613a3dda81d2039e9aa0bd50f30cc9a4c92d7ad188ff24e19c735019f9ce65d24f624537a938a9752ffc813a38d4167f215e69ca0540ce68f6d2facc38a7a35133ce68a2cab7e40584f90a04855b4c9efc231a7066baf7540464b4465a3891c06e910d220090f5de68fe52613afa2ca3783afe2c543b328242124489bb5f82c6e804fb12bb24557213a3f21819db9d6b540560fd5fef8c7040985be1edc9ec4d2b688c92ca0e02fb61239da159cfd78c3a42ca21f4f730e67e4bcaf00f8f6b920496e3c65e52691171eb81c8fa3277a1b615ef4d0ad297c81846635c70f4e07ef8c0d24fe4d32f55070e8f74747c61d3e495af57bd3a86e8e14705da1a103ba123135d13a14e1a35851fe09a5be8ac9c8c9dfeaaf697231ae841b4d1d3e06ca3357463f1418958ec741fd075930d433a50b91eb2061a12d631907d508dacadcaedf919111d8d3ef53d182c4801d29666b915939093c94bcd3f7c459abcc0a2bffbd065879dcd2639e3cc210b51c2f4d1c401154247817e4b15310abd7a6ded639aee1ecdbd6992bf555294a63cd836cc86c32af468aa916009dd20a52958021a8ed5aae9777ce379adf0720ab1154675a17e57ac2c437d4fdbd2834b45b76dee896d6e042c91e36556621e831be9f1278888bebe16c7655e9a0ce16780d269a58b59f102969cba4b03cc0c4bbe498cb8817f6e6dc9e90437f6a22395392727acae4d9cdd52148a333e5751cebcaa5a65deda638f89e9c906d580931676f4ea130a69b890e0818e386e64183f58bd05297a2676b45108cd932a25d9c9d1f21c3e4093b10262dfa534524738d92915b0fdf03eab8d1d40d2e3095d6dc993480c195b02d414db85880f61a2524a9cee78b43b555868a2018e4c6699b9a472baadadc80e09b3c5b5acbb281f548f3b2a2d5c750dedad47e7dce268f967f8992576e42425650aa2afc9952786c2e33ef637c7325722cefa8ffefe82e8cd8e200bbe68e74ea4c3832fc9c53b3c1018a594a594649dd4e457562933bf9f5398b02ce9b4c83dbae464453b6adf1b9769f610305ea9ade690485724d2a0aa50e6fe2d48778801aaf06ef096466b65d80907c4e3606c1ee74d77c1a55fd3a9221ab529361edf4283987cd626d15f2f44b4c6febcaef740f46804b703898379ad2abfbb6c527525298acd500c9ab3e60a02aed49b9c64fa55d10796d1299bd53f34dc8c3884a1ef3ed4baa087f05f6df71f99cf890b2f61f70b7919937d0852bacecdffa7941236ca255a26105257c8dbea127108db618d80affae4bd2093bd83e64e121a8dc79dfa611e3c1987f514006ebd139e07506a896aece5a58aec1e968e1e088fc89303ab1c0142a81e0058c31bbbb31eb192e92c29590e3cfd21d21c7114507a3900ee11a7b69d9ce544a2bd29a18a241b473d31a49bdf93efe33cdc020caf0c78d5921e67e592aef9b1d825ea44f8f3915bb12f04698e6aa5b0d962b8dc58053b4642fb36ddaa6d46af54869e38184d53871b2a596bc36769b6e3cda5880017b14e9e478b8caa3b6e5c394a2e2ddfa7b713914bc1b2a6dcbe0c8094df13191c93938e8fed0ba5ba67a4add8a1e77bdebe4696b5d8bb19686baabb5a700672a6385ff3d02ce7829abbaa80fd17efdbd22227342c0473301660d1adc1d5b4d9209cca49dab3c53e52aee48fe9da4ed54d2769e5adfcb7fb2329dce64cd94ccee2a03fd519155ed272423179fb23383250e2f091aa87737f8a1690ed214cb21db41f8f3e891d8c1d31de83011ff9664ee01ca20f5875826fbff1794345ae7f712cb69fbe22a7cc8d3b23233b3cbef957cb071b9672727c0c2d443667aaf5d1bf844c9be75fc22c8ed0c951415267ef9e223e138b9f22cf349c2468560ee9db091accf741681830ece912ac6ea602140ef5a578b8f4133ca46a0fd117b9eb8327874d8ca4ef7a789e2d2c5cb4bb5caea5d359c6a0c088cb1c85695e5d3b396fa1be32e8a39f9d3a90266736a4e047bac1cad4870a7a695f9e210d167448d0b3c0d59fc2387f5a014aa689e2282aebdab51708306b2c995ff3b733358c61bc63f5874b913c939413b30d7d2b2eca08e0475db8f4636ea5ef8359d97055b8364531c47030ed3d8c68dbe5b00275a7675536486a40f3af4b368cea3a91e3b4f274f148fbd94c3a307f7151bdcbd63fe941b85e5751c4347287b567cfebc650a91700d04c0be9ea249209f2471e646a84cb8c71289771e2fc71d064b1d3a73a3274b147122e725951197ac7b999c160e76b49d5b317792e9965be9769e76337f49f35c2fc5a844bdae328daea4e26f185dc838c6028ddd573abdc4db08276a48493d80a33298e6e6ffd99f65dbcdcb9366682b2b5258d8d50fd70d1c9b64aa5e4d3ba667c1fcb5144af751efab60cf5f7d5c7a45c85d09cf8406fde9282c7ae3ce7c560fb2225dc8f614e189f4bf48a3da7e35097eda053f1db1d3649e4abdc8ff8d28bae8e278ed96228b4c554f7264b4452ff06ec2f93c9c83e2809c045ffe21f4ae23b19545ddcb6f85df2841650996d501c34a1bf78180a40a1efe2cf43cd13ca986621495b91f14f6d9e121360932ccced396ba169cebccb16f74c0a1da97456fe6cb67c96351099d161251d56b5e1c0095f43c6223741ada510a2e7b6cbcd4fa80158121389d8b49e22edf87daf4e6643ae2ce09aee2d8220616393dd3a298fb494472663020d8cc8e042c3818c155ef0de95e95c42531c8df613732f0725d71ae79a6e020da26a0946691e6e67bfdcfa8e1f8d536ab1421159fd7fb0453b07f224223b7efd8692256aa878821e397b02d0261b800d7f79429615edec22ce2224309fb5502ea81433cf16ea1ad400244c173b71ababe2e9b91b90d2e505f3c941414712f77b35c46025f7ffba92c1b90c4434408579a9060c9600432e9d6b25d0746833c3488335c8ceb6e357e93606b4f78215419a3ab36ee0ecc0f8bb38f3bdeb503db2c5483a722ebb3791bf497b2820803d4a10ce060f31de35f49e8b5611158c07c89ec6319ebecc0b5692cc4a8c1d68ec71d78c9c582462d2830cad170e027ea21f30b100aa4e7ffe16d43228a17f439e7f69a255a017102078a65dd8a9281c0ed192dbad3f8a2616f4a06267eb46a7a9c4a6baa6e1428acb302998ffc06e1dc7dc02d2f9de4546a427c773ec88754a590bc08ffc8870a23cf8ec39822c083fa44dcb17b670dd935536030d17610588c766b6b0a3de86770093327174f0bfc188769725e14ef4c33f2f97d372e5fe05c1302483b822a03636e5c987bb4fb5bc4cadcf6d94592c42ef903211753c5b5178b94a0489b7a259697a3883d85a83bc6c77eb3d09857b0bb635b520f42719a8a5de5817c52ff191b9c1d8437b343f7de97d05870a53598cbcc64cbb13f31dd4b23354d02c5e70b9a993000388d176b584426c8fd0b7e889b61154accc278f5de6073ed6a94f2ab7f7e7fdd2bd1aa7a9d9053995559117e1bda392cfb68ddd9f655536da5a357a1c03664b2b6f7bce5434e496825cbdee1681b215e01dd486b2d61bc9a4ee83fa9a5ea4b8af8bf698681d29391d320fff26fe23eae511468d3b51c53594c98ec1806ad6edac8e365725b90d47c21260aeb7f7b5155ac42229952aebb3b6cce348fc10dbf541dfa0cdc87d2819c3d1ece9affedd69b84a669ef5b5da957d85b02fecfd2b9fa80f3acbdfc5340610bc23225ac867ca34bf50243376423c352c7d0ec7aa5bb19a419b26a08d673e3c919b43d7c09519b8d7588b21c51f1c386bef5ee75a5b1f2ace10d457523975711a3056de282fc20968bf9e67c2b60fc14f9e5e6ce370fea3345de62e6d31553b2261a7f8077511b685db3d8cf4bee7821e0bc18fcd7a5bdf8f2c254982b0433c7b8002d4d718c3d32f1ed80ec2a4d6f9b5714037bffc598ae67ac75ef0fc4e4a0a60e7536cf95d6dedfa7866987f36b6fe3c77460548ea6d56da2d2d7f7157f34a736a208cad8cd27bdfd01c384d299334085927922ba78232a3ee3e46d4f47737f36acb0412b29324a995df1d1ba19532ad73e3ae3092db8db570a3dec7f5cc6b0a8fecb07c50016af964de84ea3c162a48f126ecd8511c201e29178a0cfd5a7b08955e4a1921e5466d4dbc0e19fb6e731850fc484686f7b3699ec8c208a31c71a7bd2ffba4990d7ba595f6efb6ffacad16662b4cf14ec86240c1157ea039cdff7040fb48793f0e7161a7ac08b1582a216cb7dd48ebc6826e928b4c5a8aea144d8feaad8fc56ba3cb1e01e88bc3159ae373b94c241f556ffb6b47ece7244cb25943f87f80d10f1d4ef20467fed10cca71f9f870363483b4492b73bc23d60b534f0cf3e7f5c6f3a8e9bf2ec1f95a125a357b13a362cb1a1962c4e0e9d59472878058472d564aae8425507265b8793c6ce7913947ab626266c6107cdc74c0f6375b68fea1cd32bf7257d7de783812c5350df774d690ce6e30705a114d3e8b795a31ce43cd107eb9e1aa9b0c3662432447c96cef227fbb4bbe97cbd2284f3d21ee6d8b00fa9c2c5c077d1a1584a585f9cba0d059629de61a1aa2511889384ccb34b6ad5aa0aef652202a46b3a26271c0659ae0e5740f971f51238b6271d43f8f2ae97517cbe8538067527e662d70288f9ea81eceecab60d43ec0775f6180a34a59b6dce061101b143f44d742de094a1a207b76b173b9efcb13a67aeead18fb762950e958d52ae6dd31e29fb8b8b4389ae013e7f6d805ab4d0bdbec5238490febd4130869d59afdd9ea8cab5c2fde16aa6ecfb1d8402d7eff0c822f70760a3428cc36d21bc85581ae19c4657ec601b55eef3473c44e80c70dff34d730b0a0a6a2c3c51f19da1d43b9d31f57287fcd69b090218708eead9023cdbd0046030420663897b9265a545d3daaebc71006b8b8b9e3310933ca250c183b6bd7cfca512cfd3c20eac80036b14e3dabe5b1459cc90fdeba8227b8b88bb41cdae920a245b7ab1b87c3537a60ac7a532b34dd8e117")
	rangeproof1, _ := hex.DecodeString("60330000000000000001e5cb4b01f29a0dd7400b242bf96fb6156aa72a5f6cf380c93555ba1c00fbc8f39ade68866741edef76aa981e05f59d70b3448d8a7eb3ea38a15d692dbc898d10c81067b5b43681cb66e47f5f15b7e62ea28156078212ddd9a361b6c1ac597a7cda188f4e93d9b67d20ada6b69dda64f2ccaea2e0677d49cad3df925de863cf4b6d9424c42ed96c7fa24a030101d1f5c16ce61a11c7ca18c74247553d85f82602bac164b6b0e0a02633e30c867b46e68f2788a908d9f5e038b5c237e7652d239206c78edb236b2fc49bbd0003d1a23863cf1167f1aa4bd6ccb8f5265f4745a46c6321b00169c8d7ebcb0cb2bbcc9aea1c3c5133b106a864f9114e94c0bf719eb0cf1af4addb5507c28a9340e4066993f6dcaddb23d09f96c89baf38048d1c3af00b96966fb70fc4a307421b3f854c496721dc370e9be2f44f8b64dcb4d20fb5e77e02d9af3e71b336b76fe27dd76901206c3706f3f0214d9f524eed3b982a25f2a718184e3ba1912215e83aba2f2d459f36cbb98b00946a5770eb09f673885733c9ae9016b7121d6eb64747a3b114cf2d8a9a43f6e82a14bdd5f79f4bbb62d744809edbc08e3f928558f604d0a9fc6600f699d5a0d31b2bbf8b3f93241eb455b02ec5145d57e918b0fccdc33ab82b38f9f57bef9cc87e588a6e763415316cb338217df16a68a2a697694945071a9c955524115f39fdbbafe948fa8cc16e0b6c02b80eddac05d33b4422f7f0be4cd49bf3502da1fd1c046010850afe489988a273dc1075ffe37b2d03147dfd68847640f5ea6b4c529492578b1b83dbd6f6588c37496daf9de9ce8aa27850f044e8384f7d2664465c1c132dba222d6da9d7330d6449a01451a569dca5837ee76cc8cf0065d29ee2aaef854887021b84bd4e3b4a0f2388efb35971ec94d38ad7f1b8c1f1301cab7b19301687d8f09d668cacc24e17210fa01071382e5d7097efe04c50b28d4a00c831db8f79bf9f6891cdd2b24948768ecfa9217ac04c45e8ca549f262b9f4eb7037184305074de943ac0a2d182f1ac6fdd0e12003c5829760a75c67da0c1f691644e580f3a1ccdd711ce66f148943ca03a51e03ea3c127ebb192204222318718bba69e935e5781648d2f9111484c2eadd33bc917717d3c847521f6798fb11eaf127d8b20f789d65d423360ee63f8fe03a850f68f5c3bb991d25104e2ee33afb21fd5846998f0c43c88e8e6996e18e37ba46a28c839a349c69f1ef4c18e290d54d68142eab1e9cd166b80549d951bc8d7fc64391b4a9e636289cca1c3e27289ca4ce980f1ac44695af48e2220349f3720618add5066fbce8c1183862f104bf89afe511e1b34630c1f2e56e4246f6f29d1a391e76ba8a13f3efbd4cc4aff98639eeb48da69f7630db1056df979dfba83ef7569f27d34837029007db2544be1142e8bc437ed92db78b4b12a05e1b9781c5d0b14527a0c2a7aee912310289b7e1564157e6271db9bf75dd390dceb50dee39dc3dcdb730284eb28a3d7913cc6c1de43951a7dfb42d79e6ec2054060264f45922110029394e1a31ac7a6a991b06bb0b15df625f939fb3014e18dff5f3c3dde4704887cd45688effb03cc014cc582ed6e6887bc1a9a80962b90f2abad569ff2c84e61703c23950a503d0b8dd95bcb6c1cb51412a692110fdfd2c5f4d46fd151844070c8c3f1cb867b128faae069ef8158647e71fa07899774bbaf52b190e9d6f80b59e40ea6826cfe0b700d3a21c360ee84fbba8b0c1754268650d1a40b82b5a0e6503565b4a8cb481944d161e81269d54bc8760d35ba348d4b53eb550927fbef077cbb97ff8346e92f3504981add2cbfecd844b39514d686ce9708431ddb4f067ed1c9112c31a73cfce4596fdd36832a4fa2c3eb8dbeaa5d871b52eaa369a20fcc543e07536439cfe3f05dc519e27aa6e4d9e73b9d71cb779f1224f5da3b984472d51adc24e99751da43aa7d901d9664649d05f950b350e423d1952ee4ab9a208d3abe4d1f56912cea1339a9e4d769c332e3f9aadb129aa1f4ccabe096ea73848981ca611bad776abd6d5f7c12de247c2318e68d40082286a9e6840fab1f23849443dd95b2261eeb7bcc0bb1c01354a1fdfeb75428938cbc8a7d2e9a4bfb9b2a25ceb1c2992c51ae3aaa68e3740149cf2ef8b2029503c06a21e1a4e7eac1ba659b4f51efe07b4809a5d059dad621bb4df3682edfcc0e5d927112924ae537bb2dcdab40fd0b960b02d6c64c235dddbefcb4777cc8338b1d220e89fe1e9b26e9bead72d8cb64d57ea6a3b5f1136896370f327ef859c6093ed90c40bcc9ebf353a61ff3e1c0049a76cd86d87e540e4cb09ead71dea6f7db27543affca79f7d695992419c57c3bc15c2a81e3e012a79357318b74c98fe0dee462802670ca1680e59b84c37a97a2ba8321b9a82566f53f44fa0a1fe14e243185f724539cc63bd2eb687a699a77b2a52c9a177b7e31092a1909555d8a16ecc1984d20566d56c7a12abe9b4d89d9c0e9ca10f571de37c10f1b2a8ee4173dc75c1fe5043105bc21b753aed3e08f4e8c6d4ee1fec81316d2d0bf0012c95a3ffe4b338e43e04b407fbfea5f579f451e52666d5b42cd52a39b68dc10d7c1c89600b38b7f88cd5d04df58093e6f2a71949f3df41e9f9ad3986c58e3f7082b1975363db43b159522fd6a05042026369c7cdd1a73a692396094c5934229ed0ea53f7c86233001b01baf95819639a1e5a79221cef9d8255163c30df5a8deb0d5e5611914ee0e42755ecec0078bf6decfed09ec81da65cbe44ff07eae33d3841cda1791daef24656247159dec86ed2d60e84a0b8cc5ddd7a170f27f7bf534bd0c79eab4278aa4cddcd954a3468b16a7f1aabeada96e5c709df7f2c75a263680d150e70afdf4d42358bd710c2dde9378fd8066722959ea1713384783abad501ebd5746cceb1451f70da960db84c695ca682b81c8eab6ef2860d92dc970ccb5daea5aaab4033a581211d7d437cd378a95e6fd99b675cd9398f58cce75d80d1266538578ff1ad3872cab8623640cd55f56fd27a2cb979a32d65049735d7d3e169db72da8eff0e9691dcaf91fed8e41742c2fe3992df396c5e589154b98243334bfe1d1646799e29700b3d5dfd98df9880effb2750692fb6f175832348f9fb1a6522a174b986c25d43dfe3043863a0560da5e7569022d1227ee827c5a6c6d33802f128c9822f580744e4212ccdd709a9e69be4e848f7bb47a878607b3edbed850da18d13faf643a40a7f0046e1a4abf360ac936163f3b40dd5a07e7ed915dc686a4ab446ec58d719bafbdaaa0478def7cd99313baa97434a5060a23db12c65f64ae026f58ac1ec25414f56fedcb3d1b7d30086fc6928a1461b6bce153985927b53256be3873c213a607fdba308677f839b8f7e5eaa387cc3f11258e96b59d1fce798473ccad5d23b6f9bad40df5079546e3520ae8a34e8dbf1ba75ff799aed489898bf7fa5bf70bc5375881da7d3fb3f655108393de8ee526d70749f5bae1336050d688a32ed65f2737990b1007e14a50d108c2af01ded7e92a2178b2eb18bab62bc69dc9eda2e8a891016c5c7b0584e0ea294da97d43efbda61997548da1ef738cfdc20c5cf8adbd3b0a7bf238d267cef9ad195e10b80764ffbfcd801e147cfc82e64b98515b78281673181930c2a7e30ff1b812bfae2101a3ded2491b777467005370a33bdcce3b566ae5e822841502230b9b44cee0083bf2ce498ec9a5d48880d158118878f02a44b4d6c29e55783eb2b8a90caf1f066097bba5ab744f9cb52d405f90215d44c79e35760974825d72ee62da733dbc0c90be223217c68c8402f7df5bfb71ce263c4705a59dc2358c29c7183feec51b59f4daa8fbc9a1717c4823f0a26da622b1ed5a18c951af2931dcbc8607369b2f45e35a40cb4ddca17b19efeb994cba49d055300f1b375e5440cede4d3179403f4dbfc4e0edf56441e7c1ea2beda0c3815aa917fd12f40841863d85e3545a50101273ca051c544bf594a6273cda4344f3d898f27c6de2275fdc88c34c7d8960156ded5b1302a00f881144218f48b0810346fd6a540010e0c46cf9d4e0691c3b8ded1d07b9737a9864fe62c72c98a88fc8073fa7e245a63a0d0d4baeb664cbd3e18fcec1919d02253d34061dfa8b351975551419c798cb41c1a6b65cc4689547e53a67c84047b0aef1fa88e513fde4ce809274b66fac3835fb58cf15da82559c2401e6006fef1cbfc4e90bc1f9e4a3024c5263e0db3c596b75d366ec0f356feaf2fafbe981687a7762816b4d4c1a7d81ec7608fcfab87d83c7da122749e72ec64e1f9352a846f6a5aba57f0146cca2094d9126fd96dfd91e568fcb2071d751eaa1c10ba29fb676d572c1bac9cb0b93dafc3134992b2626efa14538289c97b0dd6e4da517498ae9cb1ca0131dbb8fff1d5062b2d2003af5c183c4506a3f5fe416fc3a7a850240b4f653b482709d75734a083817ce2ac179575ddb4802fe371dbdc665198050411a4697ae74849b8931a773e692db016df8aae092b0631354011fc8f7cff550db925cf0835caa911ce560022d21bde7671f6d8c100aa1e0874130001ca514732e258cde4df68aca02e364b0d7fb0b3ecc53dc740cee19c3968d96412d6b88284e0969521c83443bb352cd4a8f429211abf15f94824a950d81ff7c81993bb1e15945d12ef8869881c015dd3ab7e139a5fedbe27624601e99fdde58c837d5bf3a2a19109a7e30770d573e6b824c6b8ca5ef70cba7e06ab29c56ea482a973f217215bdfbff76b8897698a2e24f06be067668a58a71ba6dea869f5cfa9df58d1a0b65b1e18fd1294e6262eef57c8df698592cc31cae940b844553e6e1ebe2601d33c2c9acfad5507d746e867ef2206f7fcef2b71e7209ad23a7ef9de8928adc8c5c4aca4191e7443642eac4821baf80cb470155bc0ec727c75adab892ab24a48483f7d895e2ed221980f31d0be3f69640595f552d0b9ea651cd8dfb308a7ec84e9c9b289efd9c83d769109d7a93489221c99f8a7d638d4218d2003258eac8f831fd4334b0f626881039c6f3a95efeb04789796f720a058040bafa9d88cccd42bd2009a0351e45a69480f47e8eefc1b6d3995ea41946717dcd40e801c717dfff8557435197417da05cf969854fdadaec17abfc2d011dfebaea6ad0d18b37b104dc2ba2a717498e130b4c1033e65a4f357eb9ea78a6e24fd78c41c44ef98528d59fa01d2fd3c52bbfe4adab8fa98c267928638e2f80c21bac0f466e3d40fa1c3fa9aa606ee87f28384c3a0771eb4f1ae8e69f99d67e4009f1e15bf675e518359aa3efa6b2c307c2968dc67fc0563c4beac177912923d09c1dca3e956ca89739f02608986bd5941308e2b8789e3c1b5c08aff7b0bd3890e985f05a2ae917a372bc11fd3bdbd50c1039c87ae6d4565722ed0d9331635cf2e500c19828a44f3120d7933dd51e44ee6093364de8458789c946eed8f71e59ca65958661cbc247e27b9399edc3b9dfef2339177315071653e25cb223cd83495c75443081bdc2586e1bfe78d2ded8e04e07269bc0670e53031d3d94c72f5e35e052f61c949a4207522722e756305c98fb6c662c903f07ed29d1c86c63648adba025d7d9eea49bf7394f8a838cb5b7e5ffd854598625b589079e6acfda8b29e9f6c11b424d8773f8fb7948f51528b797ee64f6e2d9afe96017e897e225951408f397b56536f735c733062abce7533dd69430af7f15dcea9d9b24f30f7de753b466b99a816d0d79c509ad2f1746c9aeacb2d11d19a55cc17954eb1a43eb3d161b60c21c3631ab1e34888da4128a5f7b14bbaac34efb7a9cad35fb3e7b2eee365cf4c04433503d37dc7c51aa77d2acb294ec2db1be7be3c6d97f71506dccdd8dd0f105f7c936")

	utxo0 := explorer.NewConfidentialWitnessUtxo(
		"b83173f53cc0b0bb6de9faa2a11f0311040e61ded5529cecdb07a8efa7aa84db", // hash
		0, // index
		"09f46d737bc6290e6303d125501924501c55ee77d127767134f9c9e2d8d6828a99", // value commitment
		"0bd6b320a165a2fe477779012d7b87c7fc1898de9f4cd3a5821957c15a7037424d", // asset commitment
		script,
		nonce0,
		rangeproof0,
		surjectionproof0,
	)
	utxo1 := explorer.NewConfidentialWitnessUtxo(
		"91255f4f9c3ce88ea256611482c87e604f4ad290b24c78bd1d4a79b97aea45f7",
		1,
		"09f2ec2ba35eaee539c13cefdcafce794573249e13aebd625140c7a2a440b2a4c3",
		"0a503773d44425bb6cb503d9612c7284ce82d161d01d3c4259ab38493d6fb77dec",
		script,
		nonce1,
		rangeproof1,
		surjectionproof1,
	)

	return []explorer.Utxo{
		utxo0,
		utxo1,
	}
}

func mockUnspentsForUpdateTx() []explorer.Utxo {
	mocks := []struct {
		hash            string
		index           uint32
		valuecommitment string
		assetcommitment string
		script          string
		nonce           string
		rangeproof      string
		surjectionproof string
	}{
		{
			"a5b1c1393e19debcbcdcf3c4b4e6c6024b5fadee9d2d3160af8bd6d65071f488",
			1,
			"095be3195e450198cc53d7a0b3a21796cc40d24e34f58fc4003ffbb6a4db315d38",
			"0a83d8fd5dd96b4c5342d4609e36026af1d75a66a596876cbac63f8fb4676bacb1",
			"0014e12851aa15a3eea94587a6f11d818d1e7809764c",
			"03efab495a7afa3c4b8af99a1b663bac7b9f3d785dc7908e4e5baa292c8075f45a",
			"60330000000000000001b78b6600b30aca56e013a6f8dad552050d0ebc822bf1b213c290a77280e47efeb68dc4d6e232513b697631de5d5e27b318a4e486fb1af4e0611ec813936b22e9407a2b2a42b8ee160ab1fcd45cb1004badf664357d7405c78fa3e115fbee9dca79a090d662df4548af5047a83bd822cc5afa7ea698a489735c861bab70fee1b75749b4ab40365c0d339bb15fc5acc1137540bd6656d5f483edb5fcbeb05f13b91389dc245b79490af707c8ccd3b5df865df60b1b27062a0a3a6d67f0cc1dee00bc69f1fb4ba06fded5148b6e1d1396070b7348134c93e8443a56b6d750b9b915db4f2e95a0ba76a57503656852b7069defe05e64bec56c9d36dc3defed3af0a6857ec4d87f49d813cc30489a33e20af7846fc6de58cfe3402ca89ff642c68da032bedad9679438618d7c44d57a23a4b9cb2d02d64e89489856ee5d8735e07c5fa36d5c43588757b93e119d9cd431055ebc8f54ae837401e7b5304b3ff4062954f751d2768be60dce372e61da5237f8727261b3c6689317827dde737624f3ef0dcc29e4aac2f6c22fb8f90b980bcdd0655e08965fed9498aa13bdad7d7ceacc43975d173dee990851a8db926a57f5c02c7f849f5d502355e220c8d2c7ab60381600bd2b25fd7a626015ecc3ebc8c8ca532340f26ae6eec2633b0e5311da818dc469f62f4d8bd9ba5d33f7bbd3648cdc5ad646f79cbf8de762221378b2feb734f3c7d0889ec232a03d7e7f08c27730530e983e8e20acabed5529416f8458a5d45268e0f8ed6547b366a84e799bd50fbe180b2e9073adfd68fc32a395e8a7f9d8e8bf1f48cccec578aa25297bd4c32026e838e29e8b2f4ac2663c9ca7197fa098a08cf079700727a8a2e9d612c11ff78d3cc62b86fe9ade35a52de72e98bd50184f6b9e224366b5f03ac3b3c9b37105e9dfffb84630fc23fefb5810f5648062d590454f28bd91b57f30f205057f2156c8e69c8b7b448e0c4df779e2065c56b404b1cb928fdff72af178fe04143b1522828f2cc681ddfc68f5dd86813578d66908812adedfb0a5765092f3ea6fd68855564fdd4c1344d662b82e7e321cd4e424f6af0ca7c03f2bbc86becd79c8b12e28c5023a3ed62cb8824bfe68beb680b9db6b0e9530da6a9b28b9051b4e4dbfbeda9841b8bdb611d358712dd2a73354c146e5ee1938775897643a51513101ebb826bfb6bac844b095bbe979bc896494a2cc173bc6661d6f8ed2dec3ce29185c69d8104533f278a93a9d7815265b1aa1fd4af79a2f6e5965c5354c779033194ecddc59d00c1c3b4e53a67bf936bffac7709e520416bff3fe168f4d21b0de95e35c101f0ee2cf03c668ed0e3805c59ddefbb285261b324a158d630e61ee792c36f118692f4853e740d7d164dc12d754a04682010792dc0d2e7fec1f92f07fe2f0f53086bafd0fa557c35e5bd3caa1f5d97672a0a1b0fd9514e6b6af8c461ea09cdd74e31e42eb3d3aa0463260ad37f5174a88f7eea01777a333134acfa9e9f564097ea4ea9f020f46af54c63f998429da00d567d2293c738c53d313cb7c4c5d1288bf5dec61d853eb1e30a974a06653e47301ed09b613a6ded0f88b46adbd3f46aff017f33a5464e0e4c546e5be572d75e9e6c8d2f0316843d0d2dbfd6d85dcb7eb2afab48bb13cdd6aaca59fd28c7be1e77c963c75b9d051fd6b66f892bc78840266246108357db3c049434273be29a1113c82a375e2391e9b9251ba7c72798b4e25486afe462bd43aba333cfc3a8a75d6d9ffc51cd18c2b30dbed06a3559c61fa0fa6a02649237fcfe48aa4979dc572b1aa964e0c53fec4936fb07b9d7710a66658444cc4883fa8173a232960e2b56c337751ff5399ad87c24023bb489bd817ece85b832a81fed4cb1a9fe8ee280c9b5a631363f1fa8a2bf15a00e22c523a9d2631ec020e56796221d400ec8a19d9bc05f14b8472476de09a8147ce32d0bf0d85864328c49352bd65a15e1651f1fb24152e2283440fcdd2b520e10551fa356af0d866c318c120d958e551725d0262ea72d62829d7a2f3bd961d44d6f911efedab1fd2b938a499ca9e49a708fda016bed26c731b1c9dd63ba3a397ef816951f4b44b6d246990122c97ffbff042182ddd28307c8ff82e3e6013484311a6fe5dbbfc950fdf97038f410557ec7c376314970e24e02e3dee4f20232bf065e5e78429a4df260711f3dc7f3ae2cf52b6624fc277e0fa9f7e64453e806b8d7a86185fe37b664399e102a909b41eff23b36256c7ec76656e423aa6b7622f4b4340e8491ec07c4508a779b69e3b75b696ed4370e5c630af34b57b47c6f5e622bca39a58a2c9eb67d1f7b96e3b8a79b400c365660afa16b66b431ad96ae877e64d85d39ce73b8186d477a9074a9c45d2221b6e6789efdafbd729c17a64c5a09f64fe2469258cd956781911c97656f697259246ce39c474511e60bdc444c769e2392c1e9c90ab81ff791307aeadb0a855a3bb72abf89d0fdf3fe40ed75069ef276c7607b44bfdd0ed4ce733f0276892d7291916a1a04a1183040e36b738acf69f0f9305c3c48807eb80a129e6e95f997fd3ea31bcb022a195dd0bd9a677aed292141a492471f17a74ec163f88f59623c74218bac458345fda38bc4c5baf78a57a970c8c3c48eaa2f96fc7afc7ec5950a29f88ff1b9546dd62d60c164ddb36213bbceb9a16f1f9802fcdfaa3eb50e0b0e339d09698c5f3f341c3fc4d746d55c46d8e09d1f18ce3bf64c28669fe915294cc5b10fb36d56c9c259c7c48034693edd5b2c8540ab062d4dc89a7d7710b2314af74b5dbecc3b5c30a81cf82a246ea7ea5323554222bbf61305401cb78341ab6b5b3dcb19e585f1524f34a78697c7bdbb5861dad36415bdd012a77d6bee636f3bcaedc984fcc38b93c6e93946895d4930832765ad207e5dd1953a60abcfd5f7237935aa8f24c60b5059fce006cc405da9fce769c713f721edf64c1987d1291a7d4e3ff5966e4fcedefd27876a0b41a55d08ac38b3f656651838f1b0bebf88f53c03547c5ded37a160501217f9cde80c2de0b9c2709727bd7a15f45e1681783c978d1f491fce9633b0c1a7c3435eff0c962ad91a6ffe79b82e40ddf961b83649702770e13678032e935b12a57769c91b83c1745ffbc2788c40b04f5924013f76864f0bed0d6a5934d59a5e8121ce50d1470154024c8ba5ab73f4b01b31cf19967f007b76b30a026bfcab48baedc8e343115e6a50f66711a285c1c6e9e4a54ee9010d41538b7ac2cb4bb894df1d66e831c97eff7c263b42364c8d4143076bd0c94af2878ea5f295f0ac7e41e43d6d835276d5b11aec0a7cf095754e176013414c1a256e9d108d6f29732d7901587e04b49ba7fa08515de62c620798395e52b3f79217f1346811f6ad87bb20f1f57476897da50d4afe3e78fa31b283b5f6a3866b6e484763d8f0dbe1b9430403c97ea3bf2da183dfd3b1c3cd7841f26d6ac3762652aeb39e762270448a0f66f1cb4b4759092d1b242c9dadbe7b0c2bd199ca8320c663ac1011d1ff4ac6467463b771597f1160f98ebc285abf4e9697ddc8f61d0f55f90d182a8a268859c05b921b2a0cce2a04a6569179f4eb2a857754f5ae969e9836e6a3490e1a065d9f5479016c6b0682a0cce0ed0ae1e2d7efcdb3618097c1844fb51f98c6a41b9a9946e75c17e35db4de50fe0d1ec1bc51b9d58fd20b8659269343fd9951637de7d750d9dcf2f82c5d3cd2fa9d9deb845e2a9dbbb9c90e1b0521dfea3c98b84a64fe9d1b541884691964bd443edc9143d2a25ce48245046cb1f84839b392b925b3d85cc43e7e71e98e6a2e10f6ece17f3521eeedf9b62ea3579eda9c69832d04134eacbdc32e3a4adb9b92df1d88f2fb368305c696e935a4d959b245327d1a697d8bcc9d4849f20dc2ba1ab48cbecb4fa4debda5460a7c72b61605084eafe87e538c84f1487566b7628f869618d61713adb12836493bca4f3c8d781910e4f56ee7a3bdd6a953ccd31a6194ac49ad4525e49ba880e76e2637fcf95f588d870ef555e00d43f68708c3d019fe98fcdeeafbd7a2cd247dce3f1853cd0427bb8285873717bb5f22a88cad44b9fd7278c07c38186928df21f02b631fd762513e292e363748925ac3bacd5471a0156943daab5931a94e029d4b601216446173045af1551df829371d360bcfccf79f8acd0f4a58a17df03a27a843bab8280884810effb21cdb1ad307a538850483afb4c6d0167bf1bfaa9fa395950f0fb7491938032f1ca7a60c1e16c51f82e7a59ea945ba0fb732b79204209a185e7da4a2587673cb454c54151eac903825887967bf89703e8f136484b0ba46c7270e3e5053cc68617b2466702fbb3464c87ba8fd77bdc83327dfff718f52e3174e57746c3e8e190e419df149189fe12b0a2f429010ed3fc9053151aa645d0676ec060038b4c1d7d505027a09376ad39562f29ea542d77717ce5db96c3508d77f4c49ec48e2109c91ee134a4f448c1ea44cd7eddce15f875ffc98ea2cbcb987c30ef9c87e5902697fd1bb3a9065d29b42dea4e5637dc09df5c02df40c9d84c4295476cd7c2fef0100d2b42dcb25541847f43b376ff3df9fa20f0361bef3a9f9caed0d0ef17294c0c5e30cefaf1793877e28d2155cdb12ad9e460c640ee23eded74603e337be3423b8394b469744bc4a7528b1f87c1647fc4f296608cac9d1758dc977c6aeeac0a94ff0192c4a75c156118b2a75465cae398db6c330ebeb6096bc507de5158a0691fe36c8c6b4d091cccda897a22575fb1d92ed09a5f9d1ebfcc65c970d16036ec12ecb619fa9c0d36414ed3c43d066dbc996192fc6c9933be2c31130d61cfbe1f3fbd0f89d596b24b04eeeb6e34556f266cc3e89ab1a8d290540de42d7799f95217daff3572dce95ed34445b05874002e7de3598e4b0fdb41b9bb6629533e79eafac9ebfdf09696d0a76b2be9990680d13846424ad73d97fe43443d99ecd533da6d3c5ae03c79e3fb8164e97f77391ce2d6dc7471eaa8df99e8f899adb129753437218b16d48df5c424a82056bf4fde83ea64ca6bf5d954b1f033c4e32c53a8e288518b43cd645ad601b9c035d3a6f7fd7c13cea0f24af982687f06afdb177f2fa7f5ada705c8c1b6c9af8fda372db96ae3e985f81657f69c9227539758961123005c21d85f9f8ee0f697897ff6bce65c60a820c2ea0bd97d6a13d37556c9fd0e06fbc779cee45f044d7185d5e399a636529e8da6587ee0f7d7e80bae95942e4d142dde135dd624537050601323b6bc6e16c639c3fb96a7bb651ee4d601db19b3cf6f0920c5efcad56b67b16deb13796af83ce658484e40f271e8afbdbb7a2f26767f4e8e9cbf6e308cb3668f47757688dce2cedfb6d278e2ed4059b861217d360d1fc04702023fc6dbdd207fc803e997309ed351255e2d49baa0f6872c972be66898f011b38292be35d556acd709dcd4de5453ecb8c5b5d543cca29283d475c03f2534cd1ab7fc05328beb138ecd6823f06acaef3dfc04a7ca4301bfb0dd0db6cb80fa1b5146bbf0cb14ca8bac0088c47224017d90def3583e29d19a9701b15b540d70e64420764a57bfc21d271f7d12ad78afee39044077e3665a43c0c14e6930a16f3f5f1407a94ed8c701d2961858ddb8222c014256677691fd0020bddc6c6b8c4d8aa13749f42a16793b6ce68b7027613d0ebc0c86abb450f8610f0aa2098ab9cfbd4ef2881ed70cbee7a99275ca484c3b190032e8135202aed16a6b0007624cf3de078c0ca1cfea377953eda50db5fa30fd3b640b9ef549d44385be501e6afd94ff98bb7e02da4a64ac909436cf059da1d7454a1f3bf1572f21ffd0b110d9eb5668334e6f9baa47c16baadc4a9eeedfce4d03a9fa018048d8ae6cab37ac19389bf417b2b2fc56",
			"010001e37e3aae3ae53cef3ae248f7f52a0e85a06e6e3e60b3efb1ef24fdd2334614c15d63fa9baf48afe8d5942c6ed44e483428cc1331e5cb5ae18463140344e768c1",
		},
		{
			"750bb8a79238619efae02e2ee8e112a555ef61c403602181e96ff364860e2285",
			0,
			"0823029f63c4ff427793a6098b3605aba76f15238cc35c562809483e3d64f33954",
			"0b8b8b7f2c3f516e88169705d025e5c1b7f2821ef8a0328f0832145e69667ed0b6",
			"0014e12851aa15a3eea94587a6f11d818d1e7809764c",
			"02ccc83d2cf1ef296a48541ceb8108eed630bff15e273e14375d0b7aadcefb12c8",
			"60330000000000000001a4fbdb01135c0943ac5ffbdca791a9412c06385528aad19aaaaa050cd03128b73ed3ca64f2467d82b24a265cfcf62d37824d2e122a8f777513b6b791500a547a8790b9069ab935068b34d544e7bce58e8a38b9ac55873f4114dee1dfd513866d5fd6167c38ddec0aa9fdf71d49aef23a7d05d41f3594d64683b9aae480f27f3feaa34b094ff43bd8ed766a7364f7635de75f3a698608ffda4b6f8cc6d0d2ff80963b10478e8698f4f3940d4d35b6a479c1eff941a53d80bc00f2557bdfc61eb0e4b712a3b19dd6ed554d4bb5becccb4175ce4c2e2dffac6096daa7304a4306db40d6fff05a58e32560a015a33bf3dd02fbbbb6e7ff0312e85159f9a9ea50dbca16c85ab5e0659c375957460e7f35a66a3c487ede8a8b141a0dfcc8f4eae089115564db2262853e14f0abf9afaf2374705a5c4d89257e29021622e4696c5cba8f32388615ac2125eff9ad32956a2b19e1bb57ec53f14d0378dc2f921643c6074852d477c0a3f104c2fed9f098bf72bce994709113726190af2f8c33ce6180885395fa1b5407742d9e87db740b64caa9e4740792a83c4641a6d26d1b08fd4df5f590660255a8a6beb74ba5c6d58c065296120639e004a85056ebf834002c9c24f3b9e85b9010f2d5cf1e50dd0a5d1582b3b2b441f877e4a9ef2bd50279df647cd2bcc03a53646dcca2a62a862b126a4641c2a09aca075a55732f06c4b3cb77cbfda6c2947505c02358b85c0dfdbd6dc75e9ec52fff9f1e52b7bf5ef555b0601bf0e7097ecaf642a9b74422fa5d5d8f94cc4f886f838d4ebb98b9e6e942281e1178dc7f67f91cd84da0d64392622c37b3afd47d5ee4bef44a679fa84240ec7212bcf863fc44e5b4fcf885364cbeb4ba6695b479c0d8ddeb75c379b16e4388d633cb030406f94592f5ad05e871f95f84e1aeddbb7cb1af5395a97a82bee00ff13ab1a170fd7d924672121ed40c941930c3abd7cda7081b3ad6d74dc167f769565b7e972ebf5dfb74e26c4491b22708b84bad7b4f27fc463bc2b94a801a7f5741f0ada43be9f888b30023493972bf212bf6390365756d39bf7e9832e9aac214db6429b2bc38bbb2f0c64ea8b8b717343391f5ce8b889d6e9c1f5af64b2fb073f72fdbfeb56a043d30489b86efb0cb0a01bc47e54ce0756b6c807dec970cad069b604f3735320ead8d34d5e34e87c8e9236fdf7f4bf33ea5af1a797a27171bf9db0032b0391d60c9bb46f3cfaeac242f227ba9bf46b6e08012e66c198b8d9c21219716ef8b0ab6cd78a2139a99b74b4544872680306b3d8b20e277f831a7eab8e88b6731e322dbd6211576be467cdd31899a15562e12a3e2d4d31f904f991c2ca35c35cbd318ed4ff8eeea8f1189603644db99fd00bbb5447238db974681cc8fb22c380f18933710ff2802b6fb7ead8696581f56f1a7487bf36b36efbd32f9b099a9968a48339191dbb4e4067702b85e8a008b3c803277ecce8612bb58275578d2c3cb8c1d19ae8cbd72870f9b86ec4c6160dc96c6b26754633dfd9e7bfd65ddce7d53f42f363861b918b8064deb5c2bf329bee907b76828dd42fc77bd8109e813491e38a7963d417a871042204a05f59af24bd11b5574906bd364cc2c8cb1b6b71488f6d7e8903b89f1afeb6eac9747efeaa5f59493d8b3aedcbc4db502607ffbd050a5b0ad9eb44f43a510a4dd51e81f4190c1204ff1b3ebeb457ad647299d5c43622fde75c533017e28a5a9f144d671f2c7778ffcc489ef7075c046bcc459774e8c7175c3a5601f3d50496881dbbfab99f60f956b56c8f78d6151565fa4e0606744509664295aa87a9a2f63afe6dd6985b372a7f207f27388f0f9abd24eeb09e14bbcfc9ea2c0936807ac03b3f30b87a1182738d9978a399610e819b5c5b0c506f5e04fef2ba0e35895834ca03a9bd30dceb3fb7019edbea4968a2ad2157b54df40f5864f6b8075f723634a1c14422f8cf88feba368abf2d7f79682a49f4c784bec927c5f9c71d52d81864eb50120233920732009bb2104604bcdb4cffc18fbaf388af248950190c46843c21547281f316bf9ae9b15c4b97b8cb53930134b13e813e515851cb2740d560ae1eb911c8bf160d7b70666ee71cb301998cc68936d7962dd96d1f0840192a3ad04fe25031aecdd773c660e8735916458266beb9ad486e5423a7f54e6840241d9e5677a6ce94ff5d9f3264b312119889e7fafd722e40d02f7917c35d9bba408c0b21fb090c74ae2a0b660f30a90e1006c8a4e7dd7665729556294bc174b11d5e266b7662c155bc429f65d2f1183118185fc18aecc538d04c86b178f054844912008ae8fc06da1802ef1870e99bfe0d805682921fb9f634166f4e1270bf149d62744c01d68958183851856632c6d918d4126e70c98db9be59b80f358ca050cc6f378779a2714c615b312f831955da2545bb7ecafce177899e55d56018fcf592c36b30c27e0c8a098c8cf77b9306ee58323e44bde9886a4f5e94ca7963ec11f30501e7357850391fb4f626c4baaa5da2a2a72741ebdb05075962d546820b4fccdb2895f5bbc0b138160c281894c7e251d6625e00785683ec87ead324cc0605681e4752001b307b4aa6000d9eed674491ac80a8b08d37acb499e8e975608a7bd279e6551fbc895d4fc55952496c55ebb3c497b6c9d4f33df54c3b3296cac1a74d9fdd816d603f7a3826ce7b88e763bad580dc4a98257c1b74095361da33f7d7f307b6dacf72e95c2e234272188b57b3348458243356afccbe88d8e3306bcaeda47253735fd0bcaf59938227d07b9852da147aba89b0eebd41fce68451a5d46dcac3da05710b6f2b440f1d03a03371f8e26a8ae82b8e3f4a967a7d1911eb44c07764d2b0f52eb891342d31a08f7e48581cf883a52d23dc6f493067a7a8cec38bb62023ff41cab6aa5dfba04af72fd7a7474be4d0036133b9f1468a906618ce5db6a57630fed76f11f69bba5c19c51ddc3caf9092f8338e96afe613c53b839d843283adeecfc5309f964209cc618de3cd17f02efec1783e9dfca5a0a8a2b20b183d703aefea52beaa38ee3f40eecb5dbcbb2c6afdef2a0266f23793272e3924bacef61a9936cb9db265d9f4a7bad9205cd1d2fa46fa21c050a3bf1005e6fcd5b61c76fed6f2d676bface532ac31857b27e1fadd624aad7635d801381b74056cace9acfcd02f3398415b39727f1b8af9e4e3a060a9c3c76a0508b287583f3ffc362b2e4bf0c65f0259ef00fb7fbb990a0ebabc3c219259f8eca2c8d8e0cfe78f0885009a99122bc8759fc92138e50ae05d466c42fe1cecef46276b9b1a51db4b8b6b5a5ddea1fd65b75dfc080a2ae37c728f060d48050e0d0d0eb3a7f6c11479029a7ae5f550e035ac69a8dccd17448fc9a04ed87551108aecfdc3226436f7c0d06d1ddb97b26f7c2bde8812db0db887038458c38b58507aade420c8ae6450a23f53c8045f2ffaba8692235a9d0eb27a9055b67dd855de14a7874c7f85e20e1589cab2402158add007573c173c3be836513fdad6636fafb3ab0869166d953442ad400efa3b78eb18499c59201a85d7bde2c3957654935f03a55476b1603f1a337205bdcd4254855222a5105d789316b146e7e4c25acb7f1b0ec3e3e222a95522d7b8e79a01ff19c1e1cb783f6babf368fb3149d406488001e22dff4078d44803f9130a98a62ce43d5aff20aae56f66b96dfcd9bea1d8ef5158a6ff15aec0d663d85f1b07b8266c03705029bf46ddc045d3e2c08ad36c54d513e2f829c17a108682a8a2717220a712a42b35b1c37676d37979f21dae2246dad26e4c323960da846df050123952d3e25093f5d68b8c0e40bf2954fabd8cca77d08c9629a323c162edfbc8d6d6333e1be1d32b39bb138e9f5ff3ebfe3d41c12d0156d4ae17ca445c12b48e8d598bdcf0a0fc9ea9c641a37e041d246684a690423dc354765d2d11b66735b0db83eefb26d4d96e8eb0b37e9eabb655f322daf528a624b1ad686a22c1364671e25c7bd883a25796ee03403a46744d39d9b8d2bbd966a64163d484ed2b653e0388bba16592860e4ca19a93de13c49fefeb62796f1dc7a3341569bdede71a3aa21ba6b74e1e6a48318fc0f657bd77ba84789d2dc8677d6b4c75f44830f33ac19f2f0e1376ba07af6c42cf92596db1561a4459ca954292c5605f0d06681811dd8a3fa55198ebd78e90dd767c3c3c06f89987f52703bccff0cd1d666b43fd03472cb982ab6177a5f1d2ae0ce1cf0297508ad4a134f3f12286f471fcbc114cfaa6ab4068f2cd945d8b88464b2c8f42757b13919f6a0943f75ae25b6ab6de22e011be297f57c1210b17aef5c5fda7ec38af500a35584d244ff42210bb58b994f01b860b1063b83050ca772dc71535aee87bbf04d1d18b0194b1a1b0c5cc814a73efa0a6c9ea6022021f76e416cd36906cda005e4f47deb318ce573a5e4da416e0bc56cb3d0b140e25891f8d704523ff342aa52494544cac86957c2e12e7f14766fa01829b9e67c642621b59d1c08111f20d30de0527425b567ded506c26483142a6efb9e5aa95de52ef3bd66457b68189c1f7a950633ac1891af2a030cfba443f1f6d4e2db10975be1580b5ca53f3cf9dc49606efc78f408a8eb14760961b4349ccdb7715cb1d1b7ab3e6d4b1c006f17e04c48c57b9e07f2587c07ebf4b8c3c76152b53910aaaac6fae9f4799cd4dd3ed9580fa7e6e9c0e0c4f3110db91063d2aa256b8e075f3216614b0797fa1a8be4fbd66e623cf822c7c6907fdcc8f9c2fad31af7e4a7c90f7e915389011627ffa921987bc37baec413ff5b4af18ba86e4b1b277a0ca36cbd33af691855295b4aece013b7513b532bf345399cca9f8dd209d2477e881d692b31fc5f4369ec1cad838b89a8e7ec0299e69586e07063816c1d923ae3fb370fdabd97c48652630e5f02d96cb0797910fd538c0b7a490c1c428f1a885e824d1e567468cb0d20b0b3441257c832831abfa949f2368534847e3bc368ce9c6f744f53e1cbf8c55d812eea1751b4915f88b37aa9a8ab89bed6eeec1d26323ffc4ddba7f84ff09a5579dc3a6938ebd30583aa0f576dfee14858b8a6a7b44e9277b65c40036bb4906e2b40b12227a1538aff35f981131b3c62434059684723badc7aecc5ffcdb1ef10ef2cc4b66137fa8fd893f3c0866acf6d40acfeb793a5c3d21df5ba922f7e8dd8a1e8ce03b1a3f9f8bddae5e48a08e2d6f320d801da42fcfb8f93c50e15b4e74420edd695fd41ba4f188f0682e646ad04d1a815525cc59139f9690a286845029e91a894f2a71a49cb6bd10bd6c8ec75e0e831bde42e2f9ed1848f14c33cecc743f0292328ea708bd723578e57db8751cbb889150ba9b86aad0ab1d11d0dc56eccebda24865c126127c4690b3985c20579ba10a3891cb2ac1a58f68af137624ce659e42e5bb3575289c1577b9d767c6ae9052f59aed947774c287fa61719e72228e0b210555e5179f931709af6faf9135ad61e9eda94e846369cf7b249b65cd97ea0d453b6f85fc3b0b151c058ceb1064919e87809c3d6090f4d994196c73a488b7114177bc0b355be520a3035389c531c76c29f783a60fe67849ae78be0bbcae8e1fbe3e0b559c3a56ef80880c688d6b45ddd1f8801fec529fd6c4ed6277a5e0d8f191e4c65c0bb56bef58e35d3d076be6bdcdc1231532371fd2434074f8321abe749fc321826f411a3c9eb54cb327125e675db71e1e718a30c8f92ea084ce40d41096931f1d5e5c418bb229be11ee3b8868741819035b4361e58fd738f92b314668db7d55a8d5d100ac95e1a5664ae6de1f48109904e9229130fbf25c0325cf1f6158103bb01a5a5cf1085892f3aea1e21bf153309960ede5ed550b74a070479290f1284405006bd4d2b9ec9687eb146c79d",
			"0200032c7e99b57d17616f8cee4532063d66a2fbeb145f77f38f52c7c1cca8bf5584f44cca6b539f90f2d09f7676248896a39dbfe5716387e9f103dc9ee748947c91092aedd742fa446149b43f08ffb4ad229ba80cc31a5d9da8fea820c1dd67587cc0",
		},
		{
			"4ff258931034e53ce30ac153498e116b3838e9dd760b1b928c6fab5e2af3064c",
			1,
			"08046f447728b1840af7da5f08d3d54901b7f4a4d8ca28ffcb7a8ed05d531fa26b",
			"0bc99c7ed08034d8b453fe3d5c23eefa063963238e08100474bc9ddff6118b2543",
			"0014c2c87d69d673b21c99af8fc9c2b87ccf66b82885",
			"02145bd2e9fd3a36abac3d0534c4503555065e3df0e0393165cb4d5c816d7ef8a6",
			"603300000000000000016617f900513a65885a75c453de74db0451c0a95353297180ab55523c0078ed413b0a6a281539a4663f97558569c36b6994c6fc8d690d564a63d6f16637fe594ce323a235e875b932bee6c4f0f47b000147ea895e138c172a7be3da3810db158d6717fdecc587dcacfefcf0b4942267c5e16ce33c53e6e4dbfce259c5952f43ae55a61093fc40e3f177e6eaf18ad910bd85f65ceac7ac8f4216e523ceb403c21b91dd70c6880d1c76153ba800123876ed83472cab2b63182a0607997cd00ee7cb1a80eb442b22831c75ccc64b79270b6bb44be2e8c9f5bde6ce1d728dc880db3fcb484a589813336d014cf54cd7dac63cbdc093c30e39cca30da49d727f17b902115d84ab3a2b64bd961a2dbabea8e40af97b27eb3ebf31f04fc4cb057699b077ebb2349662bc0560752da4ad3eed3f622b36f8af15b64c970f4b4fa351df1a189b59dba6772a62c4f0d23e568cf34a074a95be5e3d08d631045a297d344795b0f78027f8258f368edbfe8c2cc097274889288c860f04fbba9a68c12af91d806e99fce3c680023bb65d7175915eeeba199bf8d5192605b36e34c5bc482326b765ccefdbc6dc73a5d7641e0d6fb7ad515deebc5eed642bad537974273ef8dd17cfe0392abce454628fc4c2dade00e3a6adfa84f707d8a677ffb9cdbd370fc44946c99682dbb39d42ff40c1f9ce280da471b7f2f8a5ddd8cc8ab0361fe4b22b91cc725b0ecc15ceff014f9996dbb6fadfc6cbd135d3c36ce04a651621e0c4211222bd15ed32bbc96baa08b65cd885bdbe112de058fc66eadf70642e2d69f072f5022004499f9f1222e14fc5b7be10f2f45480399027dd459fca48b6b6bebca42dfab069956b75cabd8eec3aecb46b1931070de1bf38e4d0ac689fc1b47eb961a0aec8ba61cba18784efa55686a2585d945f696780730db42987d2bc669216a7e981a4c8f0783a8b65d850291044d83c1afe2fd54c048f1715a16bc892442304315ffdb7d84513eadc3c7e9de285b068f5da396a766575c8eb79fa7a786b577e06f93883d9c495092ec4953ca52344eb3c9e9a7627a38c16e4447a89c5c5858ee1b061c207d7f44bbb5832c92b8fe6cd8ea5cca7be01226d2ac69466d953791be9069d9aac307647b35b600321c761f54f8369087a2e922879476f965f0ee9dd324821df082696eb4fe91e851d9a8f2297b787298567ffcfeda91e2da5628f87f34de8ba6657a39a543a9d77a4ccbb0a93ed91e1624f0b4bea38f15b6f9636c78ff711a5a6ba7a8eff4f9b5d9d08d062b39e624814c4b60a6cfb8674514ea93383fd3a5342f21e6f840950726708ed8f91170805ca65fc2f02c27532f989b8a4ecb2e0f645f799fcc34a67a473da22044a5e966178ecf90721215f91cf74ca1d5b96d0954b00b21843e209873d0f46cbf2794bb305e1ab7dfff6e00dd1fa101b5fc93ee66f65f9db151dd3392bc13871ca4085907fa41d094c5455e4425899e1b9c58ffa9d3a23cc9623378918ff80738f41082214a819e99f73a6c95d9ee94a4690864bebd025075f9358acb23f6aa49894bbf3b9a84b77fd8fc99f100db7b7c50ef54be565a06f5a8ca30735f9331e7ba81652fdd3af5e97e92bc7e97f5a95c9045e0c3f8e18080c8f1c5c9823ab9a808c2592fdbab180d282eb71c5f45097653234550e3d9efaf5489fe5dd939d406ec4d4b9a485c21bd322709008006754b2d50a8e8e3dad7d1556bbdeac2da75141d16412a2dafbb8ad2742eaf242581bea1e86b506410ac5d85da06d4fe6aa8a106b99000bcce28100373099d7ab6b5cb9c523119c30bab90c11e62a0046af50023cf570cf01e66a15b673a1a8ad62e44dfb6f44f6936865aa47b4976de05aeb053504cee380f206687609ed70a65aeb31f3d0f46fe06365cb59819ef21a8dc41abdf7fa9ac3186ed21f28ad42f6e22673647b8ad0b5205b04bf7caa586055b7e3611e7b9975f391e966711b61c9c6bafbbc9cf0589fa4226508df38c94f45613da507eaeb492a7de8cbd44993d45b9406802ccdd3e1529aebd1ee2e77d0e2eb6d3043add33ee7b006e0f3188a956c15595e62bb1d47d72abeb566c62fa4f04cb3b34dea00039a7a01244b831229ca0b3fce2a52acf60940b8d48f54c5c07091975a1ca7e31c28d3581b1a5a788fda441cd8a30fafc925e978428a0ac3d0e36f20779f6d95a50735c4719ed7cbebc2f41b8ef46a4434d6f59253abcf77db88872646e2357aac0859361a3ff36fb77fd650b7ffe7222bd592249cca2d8de6ca0bfad6e8696126b2a27c71db6ab415dff8f4bc7affcb24f4a7a281c15b61faea283f400c65c8a6cdac9537d492eab0ef9455aa7e18db437e367b8cbdbd5d231429ffec04aef4fb9c7df7ef18e17031bb638ef2c585dfe876bfacbe3a9c406f182be252b106859f6a331f5a0a5fee98f0fd5903ecb8fea28627f29577bb37b969f772ab19e48eed70f35e604d09ffaeadfb07c10ff73e9ef745f2ab7d47dac0097380bb5a7b83510f34f390c3ea1faba264cfc3823c539a2fd303c34ccb512b9dcb07d0132d7f3c903ae2283071e9c0dc89f753aa475f368314907fe8b59c37b4db9c731f3465800913e7e714677d568e9021c4f91387e1aa4c1381041d50cd5085ae71308f806979ba060c59c2fa4f213b224ca1b11c845cdc4f18c58eba92a8b1cfe09871f6f91f80b656516a6ce250c7cab8bb99d0ab29bf453719bf079866e014985b6868b707bde27cbba174461b7f608344ad745c84a894d6bee668e47c1b54769b4cd33b464e0f9c5ef7f1ffe8dfd32327a900ffae0ca40831ccb6f9d687ad270da4071c76627d10ac4dac0c692276d54308b7febcfd395c7e91ceae6f93035e4c64715eb3793ffc33edb9a208395e48779724bae1307ee67a69b8b6a94bce3275f1bfd0cd97fe1886879cebe3ea5485fcd3a45e7017c0e084badb6d27cf3f1cc8df08644db1514b37efcea85efd5ddab0652dc99bfd077424f16882ff6908b7584be1c87a8a028e3ff016e0adf28ebf1bcdd33de15e32e2f26470534faa7fc9e57e0959bc1cbbfc84b17b3f0897e245bde9a84f528068a841f720797edec5a1ea4bb0246a88531002dd1b8472098e6731b1c79a272515e57329d259ad5de35d0ae30f89fcd8f1c99a7d2a5f155e762cd9a6c02c711d5b65f487855e86c4d0aeeef4d846660d7beb0e42783baa78afb797a6664241025598b28380b5f0869d811cdeb4368fb87a5149a7fd6a3a13a9912a5e53a6e0d32955382dde550a9d599aa1cd02500b51171ac2cea3deb4375d9fbdcdc4b53982db2d43686eb6370f456d5a0ac44e8422a820b42b6e9c6cc3134a9810932459bfd11f7094b68e8a7f765eae54a305d166de5d04a44b4650e84361cafa48818f914148d55e8a9f1de61d440c7764c9be09d709b8f245076c074f9f87cc4272b98094b3fdfba47b9215273537d3192aae72f998e97ffcfca0a831800a3a00595b287452fdcc73558dfeabea52a36bca7f625256b5070ec4e2bec9f12432b19eef39aa8d12648c40501e129133096a3bfb0c9a4495f13be07d23de05ac0ab47ec147c15f086527f463e130b3279c18815056bfd456bb6dde6c1a2c23ec75b1132c313b996ead5d3ae147587c4fca663db73e4762ac2659ae524b4baffb4d760a7c4ce3c0b00c80a8ab13f06209494a19cd52323dd18eaad0b286d92e2b714a413f77466c3c9e032e6551904cd58bc9a76e264c838f9ace9cdfec67427d7eafbe6c4bd1a27d23a50966778121dad5a3508e0cf84a29d0a1d82e14ecca33f07fb5b4a9b9dbd9766536033dfdb9869d932d788f3e0a2dc3b05a25a607e265e86e34ce11deb5c5e66a98097f736e6d0967891573a72ddc52796552c260dc715323c55d1c68aba9e046ceca9189b20172e9747285f7a52833f0c4ce9b0a469257679bfcb97ae7553984b7a397c90fadea5b3c3d0391eb24b347eccbbb52ac0010f266995484c04dd74fdf8fe6e60ea360aca23e39973cb7342284fcd96d266728ba0a9dca7afb215b8cfb93a78eed43d0dd08433a1bcdc3e6d26b1d6c6156a501bd4d84e3bad68a08ff3974c3e12f40c4b9a03b179532e5ebc10f3b3d5a4133dd9125f6565781c3ee0a89ca46912d9d9e16e94042194aa1a12f1651b27de83f8d3aa912c888a57b72a56a1c21304b094380f5646aff8ae418e8abbd1cf8848e30370f8c2ce36b7e54e7f931d18679a1f04ff01e01dc2dbcc073a211892128f9dbafe27b519d431662a8ee542c572bb80baa508c791951e6638563fe65887fbf221c9fd41f0bb7dae5632c516a1627985e3f2918c3c0d4ce831174a03d9fdafcb7bae74de0f232bc4595a6818986609e2f38de28661908686805eb5ba1d297c3cac8cb77e4a48d3b496c29972bbdee6342bd30be1aa6feb822ee5ef70b39860aa5de1620d0512d0dcb16a8f1ac9eb5f6c50fac4d390096313b1424cff5ebc67afb68fdece20b69987767f1344818fd8b66982903da74cacc365ce8e41b16273201d01ccdc77e1fd7b6ec82b17dfb0d33bb1ba7df76527a221a384932d0bf5d057655f9c404587b40f87b379b2873094cfce47240cbe30708f5d8956138ba8d4d6fae7017e2402d94452d7b96f6fb2af4a121258614519b306286b3bed6ef4c8674cac90cd77cd3f98a5d61b96387bde1367be236729ae44366a5db6ce18fe1d66c768ad9b0c33ad125ebe1d12d17a511265810c44e02220074217e4ad95e028123b454def961bf4aa9846771bba0bb447501f78347d38b257dfbaab2df7f6c291b5c6a577e73690e8395e0ce1dcc98ca77751ce6bd7072c01a969abcf3f9969c871e41fae87e7febfc6d6e99c1fb14cbe39b619bff1272990313bb93002af9b8b4f7c98ee1ea06ac4391eb03cbc0f87f334b728fb4be6d56490a5f4f2e5b8bcbdff337c263f567ec984aa6d0e9e655601cf4cd2f1e33ba11a8e4cce98b2eb593c531354ce9040f694b69fba9ef337555bbfcf85af55de726810324de98ff009a03cc7799c7f9daee0596dcad4f6597a2c8af289f8fbff6ed7f482595c19c0e9ea3f73aa27eba19d9ae16f78434d107413da12416c1b33887743e911834ec6a3461d75fbc0a43be276bebe81b0052ce73e4773eaab66286fa9555e956ef9df3b32a97fa456356349075ca41fa24d57850de4227e70ee0218aab2f6a2ae6956391a6d730eb8889adb2b6a26e69d26d09d22ac491735e1798629a95bc8f1c8e744c1e893929c10a60e42b5a2f37e4a2204a59c17281cc8d89c5b30e334df323bf06e747be4049d64ea26e8e779e37ed55936f2d8a0b651e7260bb62857720b606fd0d5832449144e0133479bcc50ff09b873a04e1c9a55f6df0515572ec6343f357f8cd0b2c1d75f5d57cad3d730a0cc127f7c0bf33c6a9baf1b2d831c78cdab3b09fa71530a5a2ce3c958437d855453a7b24777f79fc35e8770ddf498c06c6b67ddb85874ffc979020ca3e30e2213cc1450438bc17d4f8053eae2adf1411383383dd64ba214f828025ab2b06af208a1e5e9db9c68c5e529ff33f97272249a9f2c953d2058e279483e961c7bd97e62eb5e8faa48779f1ccfcdf71fe0bebbc904dc7b514d8ff06bb1f6b5151c7950eae30bddb58518531a081abeecf658223d37107ccf7d7d2ff9041eeeddf671cb1aa7ec47c96bdc0dd1a3405f05f8313a15a2c94e265e49010aad31c0cf82353a5ceb1623c6d8e5fc3ba9e7d67285e4fcc56fa4c72b7a05ff0fe617a6bc781eded4ca9ce8014bdcf83d5bb3ca67f654c09e45996841cc7f77115bb6e561ef0e066d1e4659a2ac0bdd1961da1b3d0367eeeb525f288f483928e47ed159faa4e673634c27df3ed3f253535e5ebb5a65498265f159d787f25",
			"0200038fcb69d4a7bd9fff3f212c22c2338790ea3b6e5ccbb658b5723f5929116e08c7f6a0abcd895528e104b4d07b87247e3b5dacf66e6cae11989a232dd8793ad8ea981bc0628b0edab3e3491ad0ba37710a50c717f7cec9b31c9cb8a4019c207e79",
		},
		{
			"7aa572e24719c9bd33652b702c4aaceddc2ae53b4f12e59f50eab2a35af6aae1",
			0,
			"09d1fbc1af8cfc1af9f795a76ca11a726da188e36d9afddeb41322e5ecf66d713d",
			"0a5930fe31b3cda6185b6bf15f9689fc6efa94fd86460293b5548f6332306da6c3",
			"0014c2c87d69d673b21c99af8fc9c2b87ccf66b82885",
			"034ddfa2ce20d2ccdf4405e9a4d1c282a7460d390f0ebf1de74316a5b7a031ef63",
			"6033000000000000000199d028011244a303ce737c80d2ce25444ca427a49fd39fad63a651d717ed99f1a174135edee9b08788bdcd79aa9941dddca4b6bbc67916b498bd110c5bda43555b981803d02666dec8f8a03c033a8164a89fc57c61e3bd00ddf4d458899f00935970969e89d193a7fc876c59ae6bf2050bd39c121c43695a5fbf1f68fc51b647798708be30ed303e888b3cceaae39767a92fa421e3ead726817148beee2a3130f53d7b6e2b3b34975686388b86f5a1bf6e923d27cba8593eadbae662e0ebf3e94291cac3b290d4c237df9019ce86f69797da9adf6a81cc8c58550ae9743dffccad2451e9c79e634cc476b0b601cd921c08d15349882d0910030ed8271b6bb9a60a2585cdc31bc35a5d3ca0d51cb62179612b108fc678287695a54d4b02bda7bf205caf18bad512b19e0ecaec2c64ec6058271ab540c033b9ec83008b9e175affe0a16cc9a9059f52345899c7ec2794714fdd27fc84ac2cd9be0d62250637d2da3d28d2f3138af97d97377a805f0e0e4105eea0f6ed29e9cecbbf2e55a120d6661ab0458ebb22be4e89ec312dfd0debcb2dd43b617c292ea61d4e928de18253ced38e8ea5dcc76f7a0f7c519c41f556f1c44d5a81889c0aa2486a9a83d1788a8fd4d4169041e1a4373c639824289a59049910cedde09818c98bebcf08fbe9f16beb1ac8b52b3e509952ac22d397dfe067d6ea2a8b56cabeeaba19393752f7d41b2e74643119ed6ac3bbceb7f504ce39eb3546ff71d8b262c6fc9c6a741da436e34aa60c0459d7b46a93cc60772c98c79a5151552afaad0319554b3b229de2550ca2ff2f5a6a85d07da752c09ef6988a6ed71058fdb388be2f960f95b6c34c43cceea67f45561a57d53f76fa2c9db7c0ebac51d0d9e1245ad51bde6fe7c8f91482d1f28b2e6eb6346408998743b878fda32fe291cc81588f2a2c07c8888f07d66848ceb72c1d27fbbc42ffcf715ae00e697c9ac0632ec078c7fe710825f2726d3e34c8c34045718348eaa4f73175e31f30ae29f05390f4545d5e74e1a440fdff0163ce66810f8593d22bad851cbd93e2ac27a13852b1e8c63da943eab294336e4ae97bb0aab428e19a19a1fa00abc6c1b9da3879aff96d52d46910324582928fcd7216d4867a07734f96724db48a95d58a769cf8fa00ee738ab2372ac31858bff9696f959b578560466bf086ab2d2d65047ee8e977f1122ed61b9969611e1a672007adea063a5a8ed15edd6b961913eee5b892cf1ab4b8c2f6b586ae83b3c3aa962f64581e3f72dfb576f22e5546bff90a6fe89bf135b631a171d2723813b8fdbfe1c2b94d2ca90586a9bb8498a6ceb2b1774ea79edf3b81e420f29eae259cbac8e4ac01df943456c5cf21043d74d276b00776cc1047e66ca06c34d5ac80a08043f33b2ce54fcbd92f7e42a07042517ab921b2c1bd6b5f3aecc819e08e40c164ec40680e0047637f1c9a698b50bd41a6b142635694d9ceb0bdde7b86c807fcb49a5d6eac29d4d3fd897dc1aa1c2b4ac7736826c62dfcc605a19c590cc85f48b7732f069253b851e54ea7fc9f189225f74b1c922500c3323537d78140d4ae8f6e4886b081aa1e03b2f73e0521ba565f0d9d31ee451a0669dd6fa4edcce062ea613c992b8bcfed9c5a073d3e3fdf9a98eb876dfb4ebbd1e5ddb44b11b1bdabedcdd29ce57f156cff2ed3600fd2074b785ffa92ee0559d94af35856bb6b0e8ecb677b9ccbde8f1eb645adb4616967d3d7edbd7de9daf01b85ee95503820f58ccc13bdcf4f7ec569303cc0c5f5da8228417caf78e9c3f792c2d7e6ed7ff3cb91cc5fc8887fb5e3208a5291f480986286cb8afe8737d1de716f826b80d4da65d21e201b492180991cd0078b8b9c630f587c994930162e6e15717eb6036b28aaf1d6e06f2f4515d640d5c29adc7410a40c625673a2794cad5ebb808ce13754bf09af2bb2001aef995e5a2345e20a39486beec57934fbafed00438b9059a671d20621ec8c2b5ea1736bd8c164139549e8ca562ac9878b92551e59125a178c795a383f3063a91421a363e37d70e96f84439feb222741582b023a8a365fbd7ab8e58294e0a4d1b17b887d7331ae773a95fd119421e0f98e6911fb1b399265753b3d30631accf174856fe37be082d5ae28529ad4321f645cae41c8544b0a6a7a28eec0555965a241ab3fdde01998cc05433bb1ae4811df95a96c7c5a2e76dd529da61e8e223d0c12539bdac157765365abb0ecd9b53a38b4e0d7542652cd089df255f6d12b4584f467342bcfdaaf9b6d86a57c7c99321a792cd69fdbdcf6fc8aff24b5d1f0854b221948c20c55e9c7f93f3084aac3d60beccca4aa6bf36a87324e9d47a31476bcb39a0ca30cc170a12381b1bf6b3e0ecddfe4537e950a02c89e462c62209b0f9eecb0e8fbb28a5dcbab747c3a365769182c1e70b2b25448da65c6ed48aa877c3b48cbc74dac0b0fdedd0fc40022e10ae8c78739738d3f38aed897d5260b896a3510dc1e340055489fd7d562f3a79b149a49483e98907d0dad0ac5ed2842ec18b42026c5398afc5183299a9210e1b00cec7cf66e725367901a89372336524bda03f25609dd0438b1504326bf4b5f91a2d8630c4f609bc6928c1999c3c12af4a2e7c373954afdf62f317023a938fff239bbec942341d0824300d70b40ac88714f84d4fc87d2d69d252e2b19760ba5ac1058a78eb4cc26f29f3288d64321036ceecca7e2a81e1c242c57ceae90f9acddbca434ab7f07ea26c18315fd33aa98dc0450b3a1cc333d84c9203d6e5e3b73ec1dc19b762e0b2b6eee2d84fc17ef60604007fdd72bcfaaa03c4133ed7d092dc4616a191218209dde31a4de98dd6c6e256c9755bfb46abb1f35949da4f51305ddc6b31fca67ca55eae13380ae10f400125b5220f6b45230cb417f5dbff472f267bf0a1571e01e375f9e883b80c992637ae70c2796df0f3c443401bdee2c0478717088f52e4e3d1e63c3bf8c5ea17c2f71235b12059b2fe74fb7f43e4b082d8312db2174b7adc502aeb70bab120086315dcda041991a719f4fc6cac19aca6d62a694c0c19965a81fe8feed8f2770c1e233e865605696406303d5de028446b0340b6cd6a201f4ddb5535797e8433deb78e2d9ccfd84f144cd5e8fee62ec77df9bc0ed4d0059cdb9487574f477c0b7072475f444c13a538730c3a5aafd5a6f134fad6472ffe28bbd26d6b8b879d4f60a5d1712df8ac3562b682799453a9ee79323e494bed35cbee762604c0c306eb95d546e8d7df84116e4322d90be511779219876716c65ac438c31eaea8eef21b71155c9bdeb4f10b18c06469017086b7ceb2c3acb52a8452f383114de7fbe2ad5910e758c215969be432b8368f60036ff9d9093d067e09ce994a7e6246ad7e19d04c6bbeeb0781ad1b4303bf465e947522ff3ddeb25ff9f08031561f37fbb64a6f2190dadd4e4604a2e14cbd85c3277d84bf966876b89b72be51d761d7e308e316febbc2d133dc233831d22314d517b009ba5bc51ad802c184ef81d10cf4e23830bdc5fb8b0d0c11c128dd66367e8e96cd66941310b6822b84bdf2617d84764295e447d56c557c4ae6d321684fb2fd97c40c63037bf245ce69b90fdff523027aa4d855a017203a5e686f9cb3c508a9c9acbb5d9b4b59eb589a6782389c3438daa00ed320a01bb8b5e2018a2ce10ccec65f685805afdb54cf26ff07809c6239fe97494eafd931a50af8f9e2c6919733c70ea0fe93984e50cb5610a3fe0026d634cc1b197911f4f19495bbf0651ffc6a69aa6d3a7a06d03bf9d273df1a6e42825ccc34d0199fa4536971f30e5e316781383f81089a2223e08616766ee482bbac7cee8582e3d97b91af34c59e215319393a8c51c209e13de513f124bbdfa1844fa62ec8d174dde34460296ec137c13d043dbcba96c93625673160ef361fbe58479dfdc91b8b7044a26b656b94e0b18d2255f5adb49304aba8166b9cb1479e86046e22a339bbbff43e9f67b8cee50b7e9e80fc8ebb38cd7721a6fafb1a96c1ee05d47b7cfdca1ac2a9f4a396c3cde489a0129e6356fc0c7599af2b5081f46dc486a64f00fc7ad5d72483746b2678dfd45669d6a3cb614759e4b43f53fa34c829701d3accce7c3c77db5fa72e1529b8d2c6181f36bf2d936b339902cd0a2a2904f939ec4c9571ff9261e84215ab7329c25dab325b18ec1fdfcb619a7585fea4e43b133d6779cc09a78cafa2b571f834fbbe322e2b35d7c4c699c6fefe2cb506cb540c8a2831386bbed6fe9fba70b9c6c9f32c8db8d6b60c83ae11bdb36699381ccb5b0aded649ed4953b10cbfbcbda899df8637456278f013e58c780e453ff2b874570570cdd73b3f91e40481a7118965a94fcd01e7ffabf5bd1b04890a3e02e886098c8ae6fb924ef445ff6ecc8b5a94444e603570064ca660bc2930a072f4ca2ece3e6bf639d2ec159a849f3df7aed3c19cb14525374b9637621ba0c2aeeee97662dab000c111a098a6847abcef007e13a60577d37fb9c99f3a1ac0927476be1ce175e5d5dcbe67f813383de6dd8294d2aab60141ac5b475010d8fd6bcb4d3765eac6a0c914a2218b956a88480d26e67eb93d6be000a9647ba41d657f0012556b7c5634d0a5f05ed3ade0d241146d3e13509c61638b5814681576d72d4ceb00f894ec6a1a662648b1bbce5757b2fb6bf5d20e65b9d90ee22c47a1834e16816ff3993abd990de0e9f50ed8035ca8b997e8133c5c441c673c31e9288f2bda09eecc82cca42a4d4439589cee9a261622bf4b8fe7861ecd7a2b9f00e032ab6cb85916ebd5e7962cb8ebe9156fe95527ebd7725b1e145d59625171208db4c5a8a712bafa6971182a4317facba038bb01464d31cfe287140bc0502f39b20f5e29847f13a99bac3f6bd52d48d69168b95f44c3cf2be1192fddd77c99804698a8c6bd8523945db0a0c5bce796d7392c63908dd4d3bd6cf334193478a88e735dd492551873167fcad0b21ba89f3c53f060349e02e08fa8c06422fee0d4e8848b0675570c5ed4b7a983931e2127bccc669772e726bac061a368ee1e6a13fb24d3c8b7fec122f9de7a0514ef69b8ceadb466e6e36e0b1334b436b397983fde3bbf6e75d9e876f9e16689528ef2379c5758937b0c39fd07e3ca4faad4290d452b2161c4c30fdbcb773b98477a9e1291dd41c92ee22170bb54d0a044bd4b9370245ef15b014465af148abf0cca03adea6b7061bcf93b7a419c3f7b0c3d34a20357fbe72642f3a313d2e6be08a4abe41d66d3ac4747fb35dbc092af06ad97ac951a8e36e3d31f0a58cbb206c51d688127648904ed4e62c1ddcef1decfdfaf407ae3dcfbccbe8b0134d867c12e041bcdc51bf7efef5f432b683dca9aedb53f85b9cd59f7309df40a6f5f28f88423cdeb594fab8431592b2d9b95bd2fac3ffe83cfea9bc230cf82de222939c0355d850fd12803637d3371ee556a3ff4da2d00e6536e98a5e0e471f3f4943488e3443f7e3515954d3828a2da1f707e0f2cfdc7c155ebddde2cd747be19c2d8d2991fb1daf706f0eaf4a814c7b1c0c34c885b8078abc974ef400b1ef1d8f6a4b1bb04ad699770e9599c81ef032559075317feca4caa87f51ac76df9f0bdceddb60ce2b18388fe6792471eb1a395d9c56408074a7d2eef6b3beaa11505e88495cd17ed20cf287f298fb23132e16418d6cd37b73fcfa185d4c722dc4371f5382416f5625052e6e952e64e1ca9dbfb0c87596a5153b97a43cf6c3a81586ac4243432c26e522002d7dfc733208939bc09fee38878d356c749a7c557ac4d8164d967bc195fc33bb6972e17f8c0de4581c4ddb6658ba480d20a7d8d2378447e93182643489c43939953543a29e4c02cd173c6c85f1e6fadd6b5e70f5a94806349e8c052ff986a842fc",
			"010001e6171ccd1187447b6c83d13ecadb8421eb186c2f22f0a127c8aad79c54872ceac72f8cc3709848a922fdeda06d08875c6dfcb8eacb309cc57c244eb575921d23",
		},
		{
			"c4044f457ad0b46da87730b90d67a222e448804004fb05aafacd5d00df81e9e6",
			0,
			"0917b57d24c5777864d1f5befd2b0144704ec010f44244d87331ba151b76365fdd",
			"0ade81e7f969f6d13c9cab0bd99b863cd8151238f08b403ce1d9218d2cc20a018d",
			"0014621d0d58ccfcab2ee0c01879f69305b80e53a193",
			"039bdf71e07f7d1e75e8d39039f2e8167fae6788e521636735b2892b99b58b4494",
			"60330000000000000001220fcb003c09d5bb16a092de448bf88d34b92a5fe695c37b56e34e4a798e2c5aa3f82e2f977b150c0d27490fb61805f9a7bd77ca0afd6847cbbd5617a408def7874f28b1276e837a5eb4ea9a70ca5aff087f7b9c7d90f679eb63ec51bf961c2cdfa2f7827b6bfcc9f37ea5e8f6c73f4408853e1c8e795e3381fba820b044eabb38dd3c91a03718576d29457e6b878cce58184d96cab1e093a30afd157e91e72b8890ab2b672d4e5b73f1210e847747974ad603eb74bb0cfa8b1df21a56a01c5b7a1d0fa081b0570a8d1de7c1192192c43e6700031684ea29af3e1f01a208cfb1cea9a85148eab78970e193b8430e41a090b6dcd94bf5207319552cd3917708766253b526c734e086b9fd8e6436a6cd405a0941d210e88adbf9abecc91878247f498adfa920f65561c6fe0e10a317d15f5ff19838d06053f40a2c0fac96bf5a52a5d07a75b5e2de8319c996812f630c3621080331b26f628e8f44721539ed064df720e8c1d88d53ff06684456b7476efc8160d7444df63da3b7dfc8e3591ec07ef9103ebe188918f48cd52c0bbbcc6e97dbf08d6958cff42168643dd5b6db8775fc9fd441361cf7a2313ab8ea3d47a9eb9b50b7169b2a603e8d23f24e94d0f4a860c47d94c1413002576c21f7b7fe964555a49bc6eabac506f36b598c04ad7aa224424c585d04e1a90ab56caceb05d154b0456c98fc16a97d1896aafe3fde8862f5edddc894bb79c12cd85d28d3a38aad74e506860ab6fd6c4134ad7e2ca67625f4c68d9ffe4565181e4b3e058d3614790d73f81e608df8c1bae5b78e635d8f3182b868c776dc33ca09a77b30ba14337951960e2fbb5805991fe83f6522f972b74317bb3888b89e5f423dc2c3ab3ffbcce39ba606692cf510848d8b29929621d255452a179c58d9c7c7db5032da4bc00f44455356f8db1e229a5feb546436bc82e4c71e62cf3dd4f82bcc24f09fc084390ff1e1ac3e2b13c0c36c90fad2287f2ae37a4d58dd0cbab7b8a7b0f9b0a40bf2a9d916b068374ad2c7cf98a1dedd098aec90604b3d42925289328160432483fda0f7c483e01bc0d1a4af18f1c88adfbb1a0bf3b53aa25bdd2ab65d0cb63a263dc4da52ea970203dcd30d96e7e85d43ce3f8486ea8c7d557c9b054c99c7663f76fdf773be52f68e4f0f639e5b4c76a16838056e4b0440baf305e8ab6674c5c6173477195f518baa181a6a9286a898d0a0a2a7802048939d254dce30dafbad678a66165a12033a9c46af0dd9af2ed0be2c0d4a20b90ab3f950e76974e75f1dee2c4c982c29609e66e418ce29b66744176107b4ee7719d83ca3286422c6d398f16d7ba2ac1ae339725fa1313efa205c2c136769830bbbc1d9e69109fa35233ea4a6c49d4cdf4d7a62b2a60cb8e11a6547c84a5bc61e7b89b10c41d9297b28d3283ffe39092c994c88e0f10b90426fc07d8ba5977cb12b73763c8c01d15e3c0dccd78d0e0afa094c5db692a0780297bfa972cc27a9a39897919f2d3795c6a82d94502aafbe3918016b73df6f584a81147cc4062cc4d8e11a20308ab9d13465a81a074e23539fef6d2a4e016220c48683a109dbbae172d3583c70e6a44228482c4b9f526f9f52b2093782c13995fc9e4ed4d507ba45308cb78606fbd8748c21a9205394ae2c6528ac4625e670fc43f5a886fe770cfc6f5bc775a6ebce197414fefa824de8b7b235b8c44b22350a542f651474e7e2533f9c7402ab41dfa9b9cc1b4c08e327b1a49cbc8d9752dbb6bd53b93b3b2c69f83d70e14cc02dcd6d8005236aa029b6f81af4b6a5a60e602b594f39db88c9d006c506d8971a068a12e4cca0a982e9141694af7e178ee1fba90d4ca566a817b40032e964b1bdf701764e5caa61ff913b170c72c5fd94c2fa0b6ad070aee7deca78c6d8120bafee28b066d9108175f16b4f050a5a77bba0a6109530eedb6d0d450bd47ee426ce3cc61eea7b617dd5f1626bf90fbf6e684eff700528335767bd0d46dbe5bc98dcd9de06e225fabfff9cfc7c6ff42b4c02940f93af8a831ba9698488fc35eb7f400376e6257f5f08b555d5fbf561183d4a110ea5859e86bf1a968c5b5609e57c08444b6ad1455c9b67d3935a62f6d69416ba7f90234b522deea721ded5548c97f6b90ecfd373c0daf21f179f9fe9553fa9640775f42973362ced74250831e99034705f3e35e96539233a1a1d46b1eca5a943602c13bac175214201eb76ca062293234cd382cf2f15a59222a2faada82ee9218a392bc87b2abba3ffaacab8394a21198bcd5083507363780bd4b64cfdd190e89411c7c2b27f3f42fd9802dd49eee66c62f581ab67f1e4de917f7818b317077635ef90f40f35249d43429a16f33d4e8dfcb2368188dddbd1230a189d24ae308cb24503c9aa78a982eef64fdac53b1f2c570607af8ae20c7c1354359e92ffbc24d81d7c82932112a60f855849b9610becd4923b018bb52aadbbc11f0f85928f7964ad9366f05afbe4a82074aea76ce039208496d19e36f049988ddcd2cdd47aaf49be3d77f563f2de2950b4c960bf4ede19fc0f918ab03717b80880f33b705e59867e69600c2cbdc9cd9a62172091eee2e39753fe19fc321acca5caf4f8d3587cc0ce786a20a5e3a5941b62fc9a911ea552468d45c9c53aeb7ecd315a986a00a2fb64c9ea86a564cb32014682652f70c7d41e3416adfabf893f372c6a13f3c0ed188f0448fbca9df1de284c8c7ef4dade75e8eefa1d13ab482d489724660fabe418ac0aac91cacf877d93b23676a753babdcbc90ca0e61b41ef2883d676fe2a92723b1cbc9f9789340cbb8ca18e470d2f6638d947992c76fc58123f656c6229d880a271425912e5a29f77941c47d3bcf520b67c515639a48ff1dd06384e1bfd7e21b75e9385674bcc6bd8c75e855333551d1414b474e234d2e701a88268c5817f4ce461d60dcaa40206902b92d860282c79413b12fe1d8f9d8e657fce28a3fa35cb327a913c60eabb5946ca55f0d1d1d689c118954511879b3d5a2b26c1c3c80b59372397d93ade3924d6016d94fa7809317edd95bc8a4a59ab5b8fe8a6d8a72ebb40437a519f412f9804336bdc2eb0bb420df87ffc8e838ba5962deaa702989434ff30e1bd10f017f5ffb3da7ad7ed676ba254b814fb6c821ff8b2f0aa88eb80b7ca4f2f56fecbd78637affaf76bd6a6be511d0c05d0ef99048d11b5827106ca866c6b4ce519bb54c9f12380adadd4713043324ab19af5a6c81ad5b7209f355468b2b3bda7279bd2d0db418d5a244eb106cb11e21ce9c48e1724a2d01ca47c6c9d548b9b8ab2c7048eebd563b97e1dd8812756cb8297d4f7e995e38b3d5a5a96f2e1549088b1b2da582a27864c2ebea592f1d7a8b21f91b79aef52348ffdea1de746f2f196b85b81cc6350614366a2476e938c4beb9606d255609df374f3d813c12deefad8d0cfae0c27e589502a10a11fa3e76f09510d3b677f04a74cdb73806cac66655a764929502632aaf685fb60c63a3bb1cce5ff7c135e9926e3ca195f8ab28174a590c1a6714795b02f4228bac109ac8a6f81eb973f6c69c84697c2c00877c73fdee2008f82c75b69c02abf08218eef907b265e75dc1e86c1248cc09ad16613c91d33feea1fcec241fb15e63d5059eff125fce8d158d4c12db6e8a879d4dc9c28da0d90a001782daaae9ab7b67cf07897bd65253efcb77f8aafa08da9bbae8782c3bda4a9148af2c4927b69d003f856d1f0fec1a6b928d53ddcd115fc4ebaf7c9b978838a74a77e5d31c8a9260088206c6df1ddb654051b66e2b6770437501eb55dfdb29b018e98b7c45648e50ea3f24e16439382f8b21008f89df117aea5b8801913830f88fcca4305b3c8a3afd192274a4ee3e774b5d062fee97edcae95859bd26a9438d2ff8f10e26eacee35e8fad77623b45d481583fec9c7cea12cee71899c3e9f20fda3c76325b2614cb5aaf3c626366a2c5c53039752854bef44173fdbeb621f328e2a827cebe8244fa409aadbf349c1861dcb8119e9adc953133684628a2c731a5085e8d1ec04600feb9c48e209899dd4590412e81405e7c3e8ae1b3014bc93f0b4f85f062ccdb5791439ee30760b3792f2c481c65eeefee88267895f8a1edd2af39b8fb82c05cd0eb2bca25beeeaa39de7e4be51234414f30341cc5cf3764cb932ba96ad0ef11873f82652e5c5f887f411bf3ad8d7ed84d3782e4cff85f48395d625580618cf46cd5c892fa4952adaf695e70c7829263e308ff782a6601e89aff9a3fa927cde23c18b6182f0d32e9b42614223bd9c1d406c9cbab84a1be1bb6616c203a19106bfb88ed0ce636157b6735d7e92485364e8cf7d629293db03961276aa3d3ab90841ec82d1f04fc01321dd549e74a8f72f8e18b5ad5fdfa6d773bb8cfc5ba688098e05e27d514f3fb84d7fa5049e2e071211572fc53d30d8a0e77d2c73c287d60c40a98d3a8e8b8b175e600aedc7b0a364d0429207f2b68324ec930f4476386a88f0a74c90fafdb54e99fc091f81425c8d2b39112357af2d070b8411c55c7989df6f368b7b42831ecb18e56c3155a308725b660f2623a1cea9b97368c27d0176d554553382e7637ee7a464b7203b8e2cd4860df9ec745833df3d9102922cf57781996a89917fd2812648be15c6b569959e8d369782b8700a4647e491c3404dfb56036748a569eabb01d783550ee3477f809bbd31164438a2afaf3b562035545ee31063ca204ca7f0d79c18bc5f92d9ec4de8aef04ee436156b497ada5988de553d99c86cc8adbc75451e8513cf38173dd3668689f0b43fcd1536acc694d4a205687ec07538e8fd00e89fbe317a7b44601e4f9219b362f4978a7d49dbd36bfa19f0c44c1c2e04a86dd081b0b2a5aaca502e3bcf7abc7d4e122835b8b5df37d89140767d81a15d8919560e0e85342f5c33775dcc88cde3501df9c954918fdcc539f9ff06c3fc40f54e34895b873eac34c4f29c6b940bab1619bbb4e3101094228d073f87821fa9a45e3f65626d5186ed457d6c1ec40cff801e83a1157e00105272080d44a9033386712eda851ccf384facf3913b07a891e65cd099184f2b2348c85eca5674c20eecb695c2e1876370831bba39b00f81372fd001f09da9f82eb38fb1630d790da338c739b064bbf60f3af8482881c47583f5129d05a3869956c9f1c25717a2c81ffe8cdc155ce59e216ae14aeaf1060173e8d9796d6dae4690487abb3081308881c8311644f4cd16cca450e5685e7fde10212bd15e1b51fd1bccf36734c0db9f0cd471a3416061c55f7a2d9ab4d328416f1c79c40329ca2748687a27c47a217815c856c8f75bdfb8cc95740689ac850c91776c51a92eafdeb945b98cc6e9db1f596c5e25a80c48f9d023d0cf3028170d79bd8965bad0a8c5913b73655d79d1e640226015dfedab4b55ca3ea99a5b4eba9026073644aad27ed56431cce31369d3b7b7d1f34828d7b66bbb06dc1b0b354f1e240bfca82b1a1c38e09539a33a02006c536fbe05506e7c67927d4a5b8e79d511d95e6ff7b3311f6fa12b5c0bde6e5a1f8e54dd4ea489941e6b29a95456883cf88c71e2f1efa833eb280f8ce916d2c50a65fb08c1ef5da556d2e0a640e40bcd93e1eefaa502eeb928b521d5902831683b1f7d62173919a29b51839404f9cab5527842d8a7fc2524d73827a1e02f68a3b984827530ac757a3ba91a701a2bdac703cdfbafc400c83587cff7151e5d8f788462a3c441481fff072edd0c28bc182113296cd5d513ea3ba2d073b98ef3014d2a4519a15a3a03962ce0152a2d62cd62b3a5750139256696560486e2d619e08da7b9dfc45e52159484f930e423e513f9b6a43ac6329022b9b59cf0c94c208fffcc3f845ebc5e28ba3f0f737b1fb80c788f6425c873e4befc5e",
			"010001526fc3a28494cf5fcfc39de4699ee88e07980ef7b2b88f309e69ce9256de164cec8a4ba6b484b750aa68d927f526e7faec760ad8c92dea8e4876a9eb85fa2a7d",
		},
		{
			"1ff37539126dc95813c4c2664f15c3845b1946af48edc60280b01fea292182a3",
			1,
			"08900609fb657ad1c64360a6c29b9f4e5c2d0c8924e1a4458961105f792cd00874",
			"0b064976a7a1eb1fdbb7caaf1a9069e6e1a2cd6af608f8c2e216ece60ddc3ee162",
			"0014621d0d58ccfcab2ee0c01879f69305b80e53a193",
			"03f76aaa62b7850f44f8d479ed64738805511675511a5958fcdcf8ba95742c61a2",
			"6033000000000000000116051700076dce8b9e921d7ca842a98e25f77399d24f4a6bb187dfc65e29bac92b50fc108b54ec72c27bb66efad6d9a1f0678cb858cc4b58a0b8bfa89e417131103845d39f97f82277c2652bc95047d09ea443ac426dc01aa8e6153e1b714306b393ebff8362fa54e2f8566b570818ba27708e14b733115ef99254ba6bcd1b4e3b19de8fe3417ed90d768a926776a08b35d09d76a7981f3839f1c2844a35355a603229a017e274d4988f526f6f6d49d1ec47ac63767efb0bb6fa1a622861a7ceba42227ecdf06b6d99cab216f2f95d1522e3aaa3e318810487667f88c58328eb3e7f16a44cbd34aa89dd2e229c6035f84399c590cee43adb7d438cceb258a30d833760160c53b4c47fcc70a919f82a1ed61d6de29e534d817cff878a0f5f3c44f57419e35a5cb8b97f01e3dd6bea052dd38cfbec715d92a7465b5733fddb1c5f7f26b1932109508ccfc04845b26f2e6621541c720b48457191ddcd7a8111e849d49cf9f163320604ca620c7c0a81cbba13dd8ef8eaa057f0c5e91fd3d579c62cc65b32c95ee07d4c208c7c50a605d80366a99f884567680902033a10d44c758d8a960b29ed5b45df60d28ee3dcab2c7cb84fd275d28ab06e3bd370583c3a7e5a9289f88d360f815980e14dd45c60a8f1819860fda26db57244aacd1446088150fe26be89914dd8e23d99f4868894c10e57c98bc0988c33740ee6abbd698bf6179ce89e93a7ae8a822cedab3eae541f8a996f86f1871229e7d3b4871f87aef1f07ca9ccff19c12f4aea201cb631c7468a606afe9e2282a8eadc121f5c97390cee95333d0101136e6649aee83b062be54f925fd11d7068dee6e4b720ea4029857b877d038a190a1433f5416a2e91bec4a53d2cee11aa03b12b1eee224830047107aa89615cbbc3562bd92bedffd51fb7a9d27b5cec9f5bc2b9f438b62fd632367f737ee9b686bfb1d80b659de7abda4d522faf452af3184cf06fe8ec01db9e5a330f9481a23c72d004f48dbb7a3878c248a966a8aabd5286218a8a01cd245b3e4aac86afe88222f493d8ebe525ae2e97971927dd9afc842cf95cfc2e2e6a61b022b4f9df31673d3f941c912aaac41f1395779fb814b30a9e07022b6d3e795ae2c64a4fba39f9f1f20cdadb0015cf2b1afd04e5082e60535ed56c58c9ae05a9d59b4969adbb276f8af5eedf67cb564330de9d1ff372cc77221a375560d41af7a60d25b3c5a4a17767cddb7373ed1f9d18ec72c0d3fdabe17161f6a897bbbca51c2b22b325ebde352ecb8d2080a3ef72cffe1790a8f68880ebddc0ca13524b0a84f85a26924c962fc75ca2b7f0ec52223201cdb15cee0b5ab13e65d2c813126433c337e3b43a0114b4af855e210cbee4ddc6fd480700708b3bf9bc100f850dae1747fc0a58db96640636d8a7c27c487dc489cd2de8716268083f76c3f9b8a5c7b398e1453549348cc02e6a5b9766d60c11b20c7c3fc8ac26cfcb93b926a5e40e96187a00b95415c252d99bbd4ca9edbcebc63dff2b5137a36a9b02f536a28ffaf4be99bcd7536a0c1d739e78106c02f4464e33e5b6107e63347fc5666772b24ee72df927fb19b710f260561e8bbad056524a3391612a5aa51a9180e820de14ff2986e539bff1989d142ff77af73e933e6e174a085754ad86217348be1032bcefe436017482f8350a9d0cebed695d270168db6559a59db03127b7c934b163454c9be8b7a7c3ac5927552aaf95ef54bd8e7eec8d87b75875fc8a5fc70449b653249103149ad8b0cc96cc16cb5d510d870720afc09ccc496445260c4561970a2b5682056efd145b9ec7482c108cdc6500c4d08139c349ac7ed3535f3e90cdf3b6bbeadbca3c47b9b6bf9a69c847b9bc60b01e5760c91840e34ef45aa6f4e3acdf6f0d7cb31f72185caf5cdd9f07aa7d85407084d0ef90d98657fbe7680262c8cd26f5dd6a382945a31cc8a4ccc644ee914f5452bd0896a664aee12960e93b7259a7a237f1a07d4dba6a0ccd77e40e00ac14afaebc3a8e11cab1c06809728dd134577a95e4ea8229fc61140787fb3e04c8093fad4e78192571b2324eebd851fa5d74e506a70702c17eeaf219cab9da3096c53b48d76b0a29eb4a92d0b5b82aa178e5a656c2fddf9b0ea2f672cc844634e5584e3833574a393cd4e3b9ea5c7b9bc9b492d3b30c5c8ab856580fc12401056ba908a91318686ab10328cf6dd09c4c45e09fb0a0a6717257e66aa5caab2a0cbfe80c3d95ef98628666c5a680afc8a820df9ccf8972213f91c5e4f9fbbafd6464b2bccb5ad2a1483c547b4e5f136c1558585f04deaeb8291efa55ca7f3e571c1b1084fd9bd2ec5e07567b4cca97945ac9c3cebf134dad9b9b51472fb8ce8b27e4961e5e2601186a69c94eaa17bf8bcd5850ae540204a7b19e0ce8f76abe416f1dd10129f904a87a27560f80eb421f288f1898c7615e4e8d33e485b7c23faa5068ebcbf10ed96033abc5324bf841d0dc11207e1a492e956b7c2b6894abbc7691e7f18f57ab2c97ce3a169a547c761aeb675c0c4d776894d72de41283d1150a562b003b8be7d090214c9cfc3e3bd29a94bbea7a956772b72fb14d9ff0ececca75bb1e39dae78e398cd5d600935d0af964ccc90e03250db765b02bd1c6ded8ab07170187509285908be0f8105836039285580041ab5fb0a76431d49855a2bac0727fee161a69628dc8454d451958f8d0cc2daa59a3b5c02cdfe76b560cbde206e67ac5b52c2bdd5672073838555fe4a6a949ca7083b5e4f7bbd2970a345520af0b0b1f66e44e1e8d772c31492e2b38e68d83c8276bc5f6e7b694304b2ac0db3e5aeb5a0e08c6c7998c5a333d67c228936969f3d5aead5ba57610b570577a0d0a8736b832f3b99e51fc35c2a28065b99239e6c88f474027f852310151927ef360672dd5d9b67656a36f28475206216f3fb47823bd85d2123d913ff9c841b0625cadf7b38929409fe033db0c834ebc4934997c0769eeef4fcc5ccef95378bea5900fd0b3470cf9a01ac6ee6fdfd9430605c428aaab87b35bead55549bdcaca1a162e49e06abdabe95dd1bd07cc13dc221c7a8da3b8dc291d189ebe8cb21fa88b2ac37e393d4c11c8b5c3620a9016bc4912e5112097077d9d68a87b2b0e2d4109b2d39f164c6622ddd26d3cfa8f5c751df1093ab8928e41c2d27118dfd4fdd01ad72a1d9b62d572a1a8cc392daa408b3b0578e41f939a5915c036d858b391727b3e583b791809d7b71dbe59326edfc9056ff9ff8cde78bf66c6e56e77ce041c1e9ceb89318c60d4a44d50c4b69c66a21d166ef845d8ea946f24712aa3a18adf3d8a05006fa58f54f7ebd29d3a68f2de2836e05ace548ec37b25f922db828b165a2d0c8d55c6fc3c5a8e64d5b01a2d7ded5863cfc6758baec34268cc35de25fd3cff565239a086a2c3e22933e0cd58adc743dff272ae6622688e9dfd8df08fdfe2d95c5f9d0af325768acc5bef1968f3003e9ccc248acdf2ee425324ee3deb94a7717bf8c20c49cd7c8bfa321242592a10cb1cec8b3c073ee3dad79072a922f9088e79c0bec7c73047d6805e407b53748a9a1ac4901a5c0769062c135e61b0a19fbd06f6fe4c74f897cb632e5bc41c19683a1e8b048be1dcc9b0b5d6bc802fd2231fd62bc1ef4587950a859531c640fa4dcaa0b403ebf0a9503b2e7794d96b9c36a444440f213ace496da9f80eac019ae82bff604f6bb8e1d07f23ab00e3d0a67c21975d2ca14489a005612918c0735d3510819646431730f938aa27c2d34725fd04d292e6f4ed5e5f3402c4e683c8abab6a5748dec21eb67052dd738f281a3e94820f3a96b263af79a6a5d68d83364b200b0207e0980fe24e79104aa8ae1d6237a0c2cc47933735e13bafd29ba91360bd2a54d61a8aaac4c2fc32984680a14e3b82637385f57e60a1d8c3c8a6f21144fdba3efe2dd2a59e5201ee72e5718c3374c46f22e87bfe5482c6c430fa590c5b74de7fc36988381ebbe648ca89b0cc2175a86fa4d87d47535fdcd789811ca57cb235bbd294d2d00a969cb0786fed5d6d7d683fec305981b0a6eaf1a15a1e39ac4aa08e39889cca457261d97bfc50f787dd3358e547336333a5dedd7cdbe8d20f9ce5348a0f9802531434c43eb3d752f594b3dd9477cab9f159ddcbf251bb2681b346f877a2e64a4b368eaeb128873cfba4f2f4db20cef6e8c54c390257017cce4bd0e06bd0bb33d6f6e5ccbfb85859dd27fe3781cf33a5471602557a925de8390680dc6bf1853071344589e780fa399f16dc03ee2b35da2694fb9bff208c356e4b361abbff20137de6286303fb70856bdbaffe6c6de5b5029da96dff1bd9c8a57a2ebf73b6e1aca12339fc879302c3a816a9cc32424c617d17379ac46cade7f2fffdc2b5d649e709e51e621ed9b6e60ad45cd677bb4055decf37c81f66f2147023b9c1fdf58166dc871fd6129a542df68da0cbb30d63361d5a45a29c1abe776f546518c7767b7993292d2d779427babb2cef34ee6b0844ab7a4eba673d8ac2a3f04036232066aaea30f6566eff0ecbd2f0ec702181ce0df30189a8070d15c0a721918aa2b247ea3efc68f8fc55b517be9132cd7b3b7042653be1c0b6bd391ab4d441e974d2f5558e0de8deb6e2364631b7a519c6a2a04784fd8b4237693ebc03fa13a2f18353e8ae120c35bcaccc7a50d9deecc7feb31c4ba442ab04bb331aa7b8b33ce86bb56ae1cfe3681da8169e7ea6e511d4b9949cc78d489fc1d916db457074a978e6a866180c560c4f8cb93cdd0ad5d6559cf57642c324e14cff89a633dcaeb182b37c720e11885377fb285656da00399a93a27f78c936968195b372cde2b195a055bdea55c54b15f014a0fdd2394761a8f55ba18a11dde9113efe2c84c9f9b411e3d12d870388a3fd766078e8554034d2e1a14453022f9067a10e5a02a528d40b70878487791421ead2b399032af5bae80bf4d6a1e135e1b8ea29282aefbd41d5bb9eca7ee13c27fd3f7d60d4e587a69220a313e06be47ce6c9be98faa38ad524cd5237ad17eaa23d6f10d4e39ebedcd436a3b10b6e78733cf88d92cea3dc6417287f19abc02a56c9a04b441f8ce3e6495edabc38c2efbaf05648fead75e3f6d9f843ff59770288e83ee08e19ba7ff22a5c9e7459501fd733ca08510e146c7e2d0f349def7fcc0006aec217333c36bdd36b7fa79d8f318708faee9db54d7f31bdb8cfdcee543ab4bcf9b9dbf99ba7b44acf75ea6ca570e1fee1730df2256efb431afcb8ef29b281ccfc01a5d8060b70cf31653d72a4676d62d9fb50c785d7fbaf7c3e0274f2290bef6b0002e753142b6d0d95f62b6cde817a6e0e5543324d458f38b1f293f756aa641ca7c649e019291194d0393333b977026948043d9481250530f32b576c216b1de21692b3a2372b60f1aab65e372704ff754ef630237bd01ad97621853e5e03707a4626732ec83b2c155acec705e4ef218d4284d7ad9930ae28d6caf7d5f6dabc943d5d1487bf0ba22a9b587b19cc3bf08d9d4f06c04904124782c7453b4a69d9d5acd1bb8083cec8110749406978c0c80753d2de3fe9db66c6b5eb5988e46e9e89f60a66e20ca1de494fbe99a6890f8c89292b8e828360554ecce56d932ae423bd08db2df4d6405c983435269dba369b0e4f470844081a8a06944fac7e0fd73eedf42214de949113be48ed6959d6b8751b71ae86b0c0d11e729a21cc7a10135930cc53c557d06b21a6df87586d99b1c4ab00c2a50f8393779eb073c4c3a820c20ba33752303de301929ff9571f72af002daabe515f51102212c37cb07edf1b260a025d557486582de21fc2c1e52dab5923779dbbda2b24f1402d2eb83cbc2d8ca511c23abf2f6d343d4cf5d95b4b99c7cf729e03e901ef02f443acfca5a76a",
			"020003dd93de03cad93b25d09a19dc92488150a1acf3ba6b6f6fa944ef224234d963ddca0ca649cf6239b4607464575db02eac9bed5427e0adba5821987f06699c5aa512fc9881173921fc8ef65a4f3fcf636006388cef8d0f4ae3d9a408498eb83f29",
		},
	}

	mockedUnspents := make([]explorer.Utxo, 0, len(mocks))
	for _, m := range mocks {
		script, _ := hex.DecodeString(m.script)
		nonce, _ := hex.DecodeString(m.nonce)
		rangeproof, _ := hex.DecodeString(m.rangeproof)
		surjectionproof, _ := hex.DecodeString(m.surjectionproof)

		mockedUnspents = append(mockedUnspents, explorer.NewConfidentialWitnessUtxo(
			m.hash,
			m.index,
			m.valuecommitment,
			m.assetcommitment,
			script,
			nonce,
			rangeproof,
			surjectionproof,
		))
	}
	return mockedUnspents
}

type output struct {
	asset  string
	value  float64
	script string
}

type outputList []output

func (o outputList) TxOutputs() []*transaction.TxOutput {
	outputs := make([]*transaction.TxOutput, 0, len(o))
	for _, out := range o {
		asset, _ := bufferutil.AssetHashToBytes(out.asset)
		value, _ := bufferutil.ValueToBytes(uint64(out.value * math.Pow10(8)))
		script, _ := hex.DecodeString(out.script)
		output := transaction.NewTxOutput(asset, value, script)
		outputs = append(outputs, output)
	}
	return outputs
}
