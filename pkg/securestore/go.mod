module github.com/tdex-network/tdex-daemon/pkg/securestore

go 1.14

replace go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20201125193152-8a03d2e9614b

require (
	github.com/btcsuite/btcwallet v0.12.0
	github.com/btcsuite/btcwallet/walletdb v1.3.5
	github.com/lightningnetwork/lnd/kvdb v1.0.0
	github.com/stretchr/testify v1.7.0
)
