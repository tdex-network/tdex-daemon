module securestore

go 1.16

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20200520232829-54ba9589114f

require (
	github.com/btcsuite/btcwallet v0.12.0
	github.com/btcsuite/btcwallet/walletdb v1.3.5
	github.com/lightningnetwork/lnd/kvdb v1.0.0
	github.com/stretchr/testify v1.7.0
)
