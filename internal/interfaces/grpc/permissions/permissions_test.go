package permissions_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbwallet "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/wallet"
	pbunlocker "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/walletunlocker"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/permissions"
	pbtrade "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

func TestRestrictedMethods(t *testing.T) {
	allMethods := make([]string, 0)
	for _, m := range pboperator.Operator_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/Operator/%s", m.MethodName))
	}
	for _, m := range pbwallet.Wallet_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/Wallet/%s", m.MethodName))
	}

	allPermissions := permissions.AllPermissionsByMethod()
	for _, method := range allMethods {
		_, ok := allPermissions[method]
		require.True(t, ok, fmt.Sprintf("missing permission for %s", method))
	}
}

func TestWhitelistedMethods(t *testing.T) {
	allMethods := make([]string, 0)
	for _, m := range pbunlocker.WalletUnlocker_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/WalletUnlocker/%s", m.MethodName))
	}
	tradeMethods := pbtrade.File_trade_proto.Services().ByName("Trade").Methods()
	for i := 0; i < tradeMethods.Len(); i++ {
		m := tradeMethods.Get(i)
		allMethods = append(allMethods, fmt.Sprintf("/Trade/%s", m.Name()))
	}

	whitelist := permissions.Whitelist()
	for _, m := range allMethods {
		_, ok := whitelist[m]
		require.True(t, ok, fmt.Sprintf("missing %s in whitelist", m))
	}
}
