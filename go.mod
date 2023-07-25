module github.com/bloXroute-Labs/bloxroute-sdk-go

go 1.20

require (
	github.com/bloXroute-Labs/gateway/v2 v2.126.17
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/ethereum/go-ethereum v1.11.3
	github.com/fasthttp/websocket v1.5.3
	github.com/sourcegraph/jsonrpc2 v0.2.0
	github.com/stretchr/testify v1.8.4
	github.com/valyala/fastjson v1.6.4
	google.golang.org/grpc v1.56.2
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/evalphobia/logrus_fluent v0.5.4 // indirect
	github.com/fluent/fluent-logger-golang v1.9.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.1 // indirect
	github.com/holiman/uint256 v1.2.3 // indirect
	github.com/klauspost/compress v1.16.6 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/orandin/lumberjackrus v1.0.1 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.0 // indirect
	github.com/prysmaticlabs/fastssz v0.0.0-20221107182844-78142813af44 // indirect
	github.com/prysmaticlabs/go-bitfield v0.0.0-20210809151128-385d8c5e3fb7 // indirect
	github.com/prysmaticlabs/gohashtree v0.0.3-alpha // indirect
	github.com/prysmaticlabs/prysm/v4 v4.0.7 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181016170032-d91630c85102 // indirect
	github.com/savsgio/gotils v0.0.0-20230208104028-c358bd845dee // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/thomaso-mirodin/intmath v0.0.0-20160323211736-5dc6d854e46e // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.48.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/genproto v0.0.0-20230724170836-66ad5b6ff146 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230724170836-66ad5b6ff146 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230724170836-66ad5b6ff146 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	// Used in prysm which also has replace for this but for some reason go list which is used in some IDEs does not recognize semver before replace
	github.com/MariusVanDerWijden/tx-fuzz => github.com/marcopolo/tx-fuzz v0.0.0-20220927011827-b5c461bc7cae
	github.com/ethereum/go-ethereum => github.com/bloXroute-Labs/go-ethereum v1.11.5-upscale.1
)
