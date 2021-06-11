module github.com/tdex-network/tdex-daemon/pkg/macaroons

go 1.16

require (
	github.com/btcsuite/btcwallet v0.12.0
	github.com/btcsuite/btcwallet/walletdb v1.3.5
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tdex-network/tdex-daemon/pkg/securestore v0.0.0-20210603132950-1bfb0312f255
	go.etcd.io/bbolt v1.3.6 // indirect
	google.golang.org/grpc v1.38.0
	gopkg.in/macaroon-bakery.v2 v2.3.0
	gopkg.in/macaroon.v2 v2.1.0
)
