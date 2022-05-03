package permissions_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/permissions"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"
	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	tdexold "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

func TestRestrictedMethods(t *testing.T) {
	allMethods := make([]string, 0)
	for _, m := range daemonv1.OperatorService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv1.OperatorService_ServiceDesc.ServiceName, m.MethodName))
	}
	for _, m := range daemonv1.WalletService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv1.WalletService_ServiceDesc.ServiceName, m.MethodName))
	}

	allPermissions := permissions.AllPermissionsByMethod()
	for _, method := range allMethods {
		_, ok := allPermissions[method]
		require.True(t, ok, fmt.Sprintf("missing permission for %s", method))
	}
}

func TestWhitelistedMethods(t *testing.T) {
	allMethods := make([]string, 0)
	for _, m := range daemonv1.WalletUnlockerService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv1.WalletUnlockerService_ServiceDesc.ServiceName, m.MethodName))
	}

	for _, v := range tdexv1.TradeService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", tdexv1.TradeService_ServiceDesc.ServiceName, v.MethodName))
	}
	tradeMethods := tdexold.File_trade_proto.Services().ByName("Trade").Methods()
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

func TestValidatePermissions(t *testing.T) {
	if err := permissions.Validate(); err != nil {
		t.Fatal(err)
	}
}
