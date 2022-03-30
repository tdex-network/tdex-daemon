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
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv1.Operator_ServiceDesc.ServiceName, m.MethodName))
	}
	for _, m := range daemonv1.Wallet_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv1.Wallet_ServiceDesc.ServiceName, m.MethodName))
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
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv1.WalletUnlocker_ServiceDesc.ServiceName, m.MethodName))
	}

	for _, v := range tdexv1.Trade_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", tdexv1.Trade_ServiceDesc.ServiceName, v.MethodName))
	}

	whitelist := permissions.Whitelist()
	for _, m := range allMethods {
		_, ok := whitelist[m]
		require.True(t, ok, fmt.Sprintf("missing %s in whitelist", m))
	}
}

func TestValidatePermissions(t *testing.T) {
	if err := permissions.Validate(); err != nil {
		t.Fatal(err)
	}
}
