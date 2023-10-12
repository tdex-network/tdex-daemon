module github.com/tdex-network/tdex-daemon

go 1.16

replace github.com/tdex-network/tdex-daemon/pkg/explorer => ./pkg/explorer

replace github.com/tdex-network/tdex-daemon/pkg/securestore => ./pkg/securestore

replace github.com/tdex-network/tdex-daemon/pkg/macaroons => ./pkg/macaroons

replace github.com/tdex-network/tdex-daemon/old-v0 => ./cmd/migration/v0-v1/v0-domain

require (
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/btcsuite/btcd v0.23.4
	github.com/btcsuite/btcd/btcec/v2 v2.2.0
	github.com/btcsuite/btcd/btcutil v1.1.3
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgraph-io/badger/v3 v3.2103.2
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang/glog v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.15.2
	github.com/improbable-eng/grpc-web v0.13.0
	github.com/klauspost/compress v1.16.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_golang v1.14.0
	github.com/rs/cors v1.7.0 // indirect
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.9.0
	github.com/sony/gobreaker v0.4.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.8.2
	github.com/tdex-network/reflection v1.1.0
	github.com/tdex-network/tdex-daemon/old-v0 v0.0.0-00010101000000-000000000000
	github.com/tdex-network/tdex-daemon/pkg/explorer v0.0.0-20211001103242-a11e4485705a
	github.com/tdex-network/tdex-daemon/pkg/macaroons v0.0.0-20210813140257-70d50a8b72a4
	github.com/tdex-network/tdex-daemon/pkg/securestore v0.0.0-20210813140257-70d50a8b72a4
	github.com/thanhpk/randstr v1.0.4
	github.com/timshannon/badgerhold/v3 v3.0.0 // indirect
	github.com/timshannon/badgerhold/v4 v4.0.2
	github.com/urfave/cli/v2 v2.3.0
	github.com/vulpemventures/go-elements v0.4.5
	go.etcd.io/bbolt v1.3.7 // indirect
	go.uber.org/ratelimit v0.2.0
	golang.org/x/net v0.17.0
	golang.org/x/sync v0.1.0
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/macaroon-bakery.v2 v2.3.0
	gopkg.in/macaroon.v2 v2.1.0
)
