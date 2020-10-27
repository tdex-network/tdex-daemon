package test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbwallet "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
	"google.golang.org/grpc"
	"gopkg.in/macaroon.v2"
	"io/ioutil"
	"testing"
)

func TestHavePermissionMacaroon(t *testing.T) {

	macBytes, err := ioutil.ReadFile("/Users/sekulicd/tdex-testing/admin.macaroon")
	if err != nil {
		t.Fatal(err)
	}
	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macBytes); err != nil {
		t.Fatal(err)
	}

	cred := macaroons.NewMacaroonCredential(mac)
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", "localhost", 9000),
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(cred),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := pbwallet.NewWalletClient(conn)

	_, err = client.InitWallet(
		context.Background(),
		&pbwallet.InitWalletRequest{
			WalletPassword: []byte{72, 101, 108, 108, 111},
			SeedMnemonic: []string{
				"arm",
				"able",
				"about",
				"above",
				"absent",
				"absorb",
				"abstract",
				"thought",
				"abuse",
				"access",
				"accident",
				"client",
			},
		},
	)
	assert.NoError(t, err)
}

func TestNoPermissionMacaroon(t *testing.T) {

	macBytes, err := ioutil.ReadFile("/Users/sekulicd/tdex-testing/price.macaroon")
	if err != nil {
		t.Fatal(err)
	}
	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macBytes); err != nil {
		t.Fatal(err)
	}

	cred := macaroons.NewMacaroonCredential(mac)
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", "localhost", 9000),
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(cred),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := pboperator.NewOperatorClient(conn)

	_, err = client.OpenMarket(
		context.Background(),
		&pboperator.OpenMarketRequest{
			Market: nil,
		},
	)
	assert.Error(t, err)
}

func TestPublicNonOperatorPath(t *testing.T) {

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", "localhost", 9000),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := pbwallet.NewWalletClient(conn)

	_, err = client.UnlockWallet(
		context.Background(),
		&pbwallet.UnlockWalletRequest{
			WalletPassword: []byte{72, 101, 108, 108, 111},
		},
	)

	assert.NoError(t, err)
}
