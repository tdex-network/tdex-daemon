package permissions_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex-daemon/v1"
	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/permissions"
)

func TestRestrictedMethods(t *testing.T) {
	allMethods := make([]string, 0)
	for _, m := range daemonv1.Operator_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/Operator/%s", m.MethodName))
	}
	for _, m := range daemonv1.Wallet_ServiceDesc.Methods {
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
	for _, m := range daemonv1.WalletUnlocker_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/WalletUnlocker/%s", m.MethodName))
	}
	tradeMethods := tdexv1.File_tdex_v1_trade_proto.Services().ByName("Trade").Methods()
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
