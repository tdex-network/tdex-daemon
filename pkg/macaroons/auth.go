package macaroons

import (
	"context"
	"encoding/hex"

	macaroon "gopkg.in/macaroon.v2"
)

// MacaroonCredential wraps a macaroon to implement the
// credentials.PerRPCCredentials interface.
type MacaroonCredential struct {
	*macaroon.Macaroon
	withTLS bool
}

// RequireTransportSecurity implements the PerRPCCredentials interface.
func (m MacaroonCredential) RequireTransportSecurity() bool {
	return m.withTLS
}

// GetRequestMetadata implements the PerRPCCredentials interface. This method
// is required in order to pass the wrapped macaroon into the gRPC context.
// With this, the macaroon will be available within the request handling scope
// of the ultimate gRPC server implementation.
func (m MacaroonCredential) GetRequestMetadata(
	ctx context.Context, uri ...string,
) (map[string]string, error) {

	macBytes, err := m.MarshalBinary()
	if err != nil {
		return nil, err
	}

	md := make(map[string]string)
	md["macaroon"] = hex.EncodeToString(macBytes)
	return md, nil
}

// NewMacaroonCredential returns a copy of the passed macaroon wrapped in a
// MacaroonCredential struct which implements PerRPCCredentials.
func NewMacaroonCredential(m *macaroon.Macaroon, withTLS bool) MacaroonCredential {
	ms := MacaroonCredential{}
	ms.Macaroon = m.Clone()
	ms.withTLS = withTLS
	return ms
}
