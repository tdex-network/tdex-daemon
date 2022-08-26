package tdexdconnect_test

import (
	"bytes"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/pkg/tdexdconnect"
)

var (
	certificate = []byte(`-----BEGIN CERTIFICATE-----
MIICqDCCAk6gAwIBAgIRAM4trYu7swKO7E9OI5q350YwCgYIKoZIzj0EAwIwQjEN
MAsGA1UEChMEdGRleDExMC8GA1UEAxMoTUJQZGlQaXJhbGJlcnRvLmhvbWVuZXQu
dGVsZWNvbWl0YWxpYS5pdDAeFw0yMTEwMDYxNTMzMDRaFw0yMjEwMDcxNTMzMDRa
MEIxDTALBgNVBAoTBHRkZXgxMTAvBgNVBAMTKE1CUGRpUGlyYWxiZXJ0by5ob21l
bmV0LnRlbGVjb21pdGFsaWEuaXQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAR1
WPEU+MKQVHNAdK0bG89FGsU6Sw6Izqe1taOfS0hBQfP+W2YBccLwooIKmPdjtZAo
V4kZcguksW6gAdzgtQrmo4IBIzCCAR8wDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB
/wQFMAMBAf8wHQYDVR0OBBYEFI5lIKaTMgXPe0v8k9oRakHlPHtoMIHcBgNVHREE
gdQwgdGCKE1CUGRpUGlyYWxiZXJ0by5ob21lbmV0LnRlbGVjb21pdGFsaWEuaXSC
CWxvY2FsaG9zdIIEdW5peIIKdW5peHBhY2tldIcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAAYcQ/oAAAAAAAAAAAAAAAAAAAYcQ/oAAAAAAAAAQCEC/uSHHLIcEwKgB14cQ
/oAAAAAAAACcVyn//lDMuocQ/oAAAAAAAADp7kZyxh+R8IcQ/oAAAAAAAABBBISb
VTjXoYcQ/oAAAAAAAACu3kj//gARIjAKBggqhkjOPQQDAgNIADBFAiEAv+StshVA
d+iSAz/2oGC0e076aiVvWHgKSehisPugrngCIGQ3tjiqzC1+oxNNMFvr7OD4CAkb
Wwq8JtrendvmccXB
-----END CERTIFICATE-----
`)
	macaroon, _ = base64.RawURLEncoding.DecodeString("AgEFdGRleGQChQEDChCqCBfWunS_Tu_-IsVdx6XxEgEwGhUKBm1hcmtldBIEcmVhZBIFd3JpdGUaFwoIb3BlcmF0b3ISBHJlYWQSBXdyaXRlGg4KBXByaWNlEgV3cml0ZRoVCgZ3YWxsZXQSBHJlYWQSBXdyaXRlGhYKB3dlYmhvb2sSBHJlYWQSBXdyaXRlAAAGINX1smi9KiBJiORAbg2aY5wNJ0dk45Uz4Iy3gIyGHgUC")
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		addr        string
		proto       string
		cert        []byte
		mac         []byte
		expectedURL string
	}{
		{
			addr:        "localhost:9000",
			proto:       "https",
			cert:        []byte(certificate),
			mac:         macaroon,
			expectedURL: "tdexdconnect://localhost:9000?cert=MIICqDCCAk6gAwIBAgIRAM4trYu7swKO7E9OI5q350YwCgYIKoZIzj0EAwIwQjENMAsGA1UEChMEdGRleDExMC8GA1UEAxMoTUJQZGlQaXJhbGJlcnRvLmhvbWVuZXQudGVsZWNvbWl0YWxpYS5pdDAeFw0yMTEwMDYxNTMzMDRaFw0yMjEwMDcxNTMzMDRaMEIxDTALBgNVBAoTBHRkZXgxMTAvBgNVBAMTKE1CUGRpUGlyYWxiZXJ0by5ob21lbmV0LnRlbGVjb21pdGFsaWEuaXQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAR1WPEU-MKQVHNAdK0bG89FGsU6Sw6Izqe1taOfS0hBQfP-W2YBccLwooIKmPdjtZAoV4kZcguksW6gAdzgtQrmo4IBIzCCAR8wDgYDVR0PAQH_BAQDAgKkMA8GA1UdEwEB_wQFMAMBAf8wHQYDVR0OBBYEFI5lIKaTMgXPe0v8k9oRakHlPHtoMIHcBgNVHREEgdQwgdGCKE1CUGRpUGlyYWxiZXJ0by5ob21lbmV0LnRlbGVjb21pdGFsaWEuaXSCCWxvY2FsaG9zdIIEdW5peIIKdW5peHBhY2tldIcEfwAAAYcQAAAAAAAAAAAAAAAAAAAAAYcQ_oAAAAAAAAAAAAAAAAAAAYcQ_oAAAAAAAAAQCEC_uSHHLIcEwKgB14cQ_oAAAAAAAACcVyn__lDMuocQ_oAAAAAAAADp7kZyxh-R8IcQ_oAAAAAAAABBBISbVTjXoYcQ_oAAAAAAAACu3kj__gARIjAKBggqhkjOPQQDAgNIADBFAiEAv-StshVAd-iSAz_2oGC0e076aiVvWHgKSehisPugrngCIGQ3tjiqzC1-oxNNMFvr7OD4CAkbWwq8JtrendvmccXB&macaroon=AgEFdGRleGQChQEDChCqCBfWunS_Tu_-IsVdx6XxEgEwGhUKBm1hcmtldBIEcmVhZBIFd3JpdGUaFwoIb3BlcmF0b3ISBHJlYWQSBXdyaXRlGg4KBXByaWNlEgV3cml0ZRoVCgZ3YWxsZXQSBHJlYWQSBXdyaXRlGhYKB3dlYmhvb2sSBHJlYWQSBXdyaXRlAAAGINX1smi9KiBJiORAbg2aY5wNJ0dk45Uz4Iy3gIyGHgUC&proto=https",
		},
		{
			addr:        "localhost:9000",
			proto:       "http",
			cert:        nil,
			mac:         nil,
			expectedURL: "tdexdconnect://localhost:9000?proto=http",
		},
	}

	for _, tt := range tests {
		url, err := tdexdconnect.EncodeToString(tt.addr, tt.proto, tt.cert, tt.mac)
		require.NoError(t, err)
		require.Equal(t, tt.expectedURL, url)

		rpcAddr, proto, certBytes, macBytes, err := tdexdconnect.Decode(url)
		require.NoError(t, err)
		require.Equal(t, tt.addr, rpcAddr)
		if tt.mac == nil {
			require.Len(t, macBytes, 0)
		} else {
			require.Equal(t, tt.mac, macBytes)
		}
		if tt.cert == nil {
			require.Len(t, certBytes, 0)
		} else {
			buf := &bytes.Buffer{}
			pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
			require.Equal(t, string(certificate), buf.String())
		}
		require.Equal(t, tt.proto, proto)
	}
}
