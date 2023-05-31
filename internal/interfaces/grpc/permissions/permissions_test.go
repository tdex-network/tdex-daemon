package permissions_test

import (
	"fmt"
	"testing"

	grpchealth "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/permissions"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
)

func TestRestrictedMethods(t *testing.T) {
	allMethods := make([]string, 0)
	for _, m := range daemonv2.OperatorService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv2.OperatorService_ServiceDesc.ServiceName, m.MethodName))
	}
	for _, m := range daemonv2.WebhookService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv2.WebhookService_ServiceDesc.ServiceName, m.MethodName))
	}
	for _, m := range daemonv2.FeederService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv2.FeederService_ServiceDesc.ServiceName, m.MethodName))
	}

	allPermissions := permissions.AllPermissionsByMethod()
	for _, method := range allMethods {
		_, ok := allPermissions[method]
		require.True(t, ok, fmt.Sprintf("missing permission for %s", method))
	}
}

func TestWhitelistedMethods(t *testing.T) {
	allMethods := make([]string, 0)
	for _, m := range daemonv2.WalletService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", daemonv2.WalletService_ServiceDesc.ServiceName, m.MethodName))
	}

	for _, v := range tdexv2.TradeService_ServiceDesc.Methods {
		allMethods = append(allMethods, fmt.Sprintf("/%s/%s", tdexv2.TradeService_ServiceDesc.ServiceName, v.MethodName))
	}

	allMethods = append(allMethods, fmt.Sprintf("/%s/%s", grpchealth.Health_ServiceDesc.ServiceName, "Check"))

	whitelist := permissions.Whitelist()
	for _, m := range allMethods {
		_, ok := whitelist[m]
		require.True(t, ok, fmt.Sprintf("missing %s in whitelist", m))
	}
}
