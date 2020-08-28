package explorer

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"math"
	"reflect"
	"sort"
	"testing"
	"time"
)

var testKey []byte
var testAddress string

func newTestData() (string, []byte, error) {
	if len(testKey) > 0 && len(testAddress) > 0 {
		return testAddress, testKey, nil
	}

	key, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", nil, err
	}
	blindKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", nil, err
	}
	p2wpkh := payment.FromPublicKey(
		key.PubKey(),
		&network.Regtest,
		blindKey.PubKey(),
	)
	addr, err := p2wpkh.ConfidentialWitnessPubKeyHash()
	if err != nil {
		return "", nil, err
	}

	testAddress = addr
	testKey = blindKey.Serialize()
	return testAddress, testKey, nil
}

func TestGetUnspents(t *testing.T) {
	address, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	_, err = Faucet(address)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)

	utxos, err := GetUnSpents(address)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(utxos))
	assert.Equal(t, true, len(utxos[0].Nonce()) > 1)
	assert.Equal(t, true, len(utxos[0].RangeProof()) > 0)
	assert.Equal(t, true, len(utxos[0].SurjectionProof()) > 0)
}

func TestSelectUtxos(t *testing.T) {
	address, key1, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	_, err = Faucet(address)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)

	utxos, err := GetUnSpents(address)
	if err != nil {
		t.Fatal(err)
	}

	key2, _ := hex.DecodeString(
		"61d35040a47f43433fd902f34bd3a5d988e4327d60b46f6f1540e93891730789",
	)
	blindingKeys := [][]byte{key1, key2}
	targetAmount := uint64(1.5 * math.Pow10(8))

	selectedUtxos, change, err := SelectUnSpents(
		utxos,
		blindingKeys,
		targetAmount,
		network.Regtest.AssetID,
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedChange := uint64(0.5 * math.Pow10(8))
	assert.Equal(t, 2, len(selectedUtxos))
	assert.Equal(t, expectedChange, change)
}

func TestFailingSelectUtxos(t *testing.T) {
	address, _, err := newTestData()
	if err != nil {
		t.Fatal(err)
	}

	utxos, err := GetUnSpents(address)
	if err != nil {
		t.Fatal(err)
	}

	key, _ := hex.DecodeString(
		"61d35040a47f43433fd902f34bd3a5d988e4327d60b46f6f1540e93891730789",
	)
	blindingKeys := [][]byte{key}
	targetAmount := uint64(1.5 * math.Pow10(8))

	_, _, err = SelectUnSpents(
		utxos,
		blindingKeys,
		targetAmount,
		network.Regtest.AssetID,
	)
	assert.Equal(t, true, err != nil)
}

func TestGetBestPairs(t *testing.T) {
	type args struct {
		items  []uint64
		target uint64
	}
	tests := []struct {
		name string
		args args
		want []uint64
	}{
		{
			name: "1",
			args: args{
				items:  []uint64{61, 61, 61, 38, 61, 61, 61, 1, 1, 1, 3},
				target: 6,
			},
			want: []uint64{38},
		},
		{
			name: "2",
			args: args{
				items:  []uint64{61, 61, 61, 61, 61, 61, 1, 1, 1, 3},
				target: 6,
			},
			want: []uint64{3, 1, 1, 1},
		},
		{
			name: "3",
			args: args{
				items:  []uint64{61, 61},
				target: 6,
			},
			want: []uint64{61},
		},
		{
			name: "4",
			args: args{
				items:  []uint64{2, 2},
				target: 6,
			},
			want: []uint64{},
		},
		{
			name: "5",
			args: args{
				items:  []uint64{61, 1, 1, 1, 3, 56},
				target: 6,
			},
			want: []uint64{56},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Slice(tt.args.items, func(i, j int) bool {
				return tt.args.items[i] > tt.args.items[j]
			})
			if got := getBestCombination(tt.args.items, tt.args.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBestPairs() = %v, want %v", got, tt.want)
			}
		})
	}
}
