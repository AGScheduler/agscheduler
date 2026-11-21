module github.com/agscheduler/agscheduler/examples

go 1.27.4

require (
	github.com/agscheduler/agscheduler v0.20.0
	github.com/agscheduler/agscheduler/backends/gorm v0.20.0
	github.com/agscheduler/agscheduler/backends/memory v0.20.0
	github.com/agscheduler/agscheduler/backends/mongodb v0.20.0
	github.com/agscheduler/agscheduler/queues/kafka v0.20.0
	github.com/agscheduler/agscheduler/queues/memory v0.20.0
	github.com/agscheduler/agscheduler/queues/mqtt v0.20.0
	github.com/agscheduler/agscheduler/queues/nsq v0.20.0
	github.com/agscheduler/agscheduler/queues/rabbitmq v0.20.0
	github.com/agscheduler/agscheduler/queues/redis v0.20.0
	github.com/agscheduler/agscheduler/services/cluster v0.20.0
	github.com/agscheduler/agscheduler/services/grpc v0.20.0
	github.com/agscheduler/agscheduler/services/http v0.20.0
	github.com/agscheduler/agscheduler/services/proto v0.20.0
	github.com/agscheduler/agscheduler/stores/elasticsearch v0.20.0
	github.com/agscheduler/agscheduler/stores/etcd v0.20.0
	github.com/agscheduler/agscheduler/stores/gorm v0.20.0
	github.com/agscheduler/agscheduler/stores/memory v0.20.0
	github.com/agscheduler/agscheduler/stores/mongodb v0.20.0
	github.com/agscheduler/agscheduler/stores/redis v0.20.0
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/elastic/go-elasticsearch/v8 v8.19.7
	github.com/nsqio/go-nsq v1.1.0
	github.com/rabbitmq/amqp091-go v1.17.0
	github.com/redis/go-redis/v9 v9.23.0
	github.com/twmb/franz-go v1.22.3
	github.com/twmb/franz-go/pkg/kadm v1.20.0
	go.etcd.io/etcd/client/v3 v3.7.4
	go.mongodb.org/mongo-driver v1.17.10
	google.golang.org/grpc v1.85.0
	google.golang.org/protobuf v1.36.12
	gorm.io/driver/mysql v1.6.0
	gorm.io/gorm v1.31.2
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/bytedance/gopkg v0.1.4 // indirect
	github.com/bytedance/sonic v1.15.5 // indirect
	github.com/bytedance/sonic/loader v0.5.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.7 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.11.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/gin-contrib/cors v1.7.11 // indirect
	github.com/gin-contrib/sse v1.2.1 // indirect
	github.com/gin-gonic/gin v1.12.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.2 // indirect
	github.com/go-playground/universal-translator v0.18.2 // indirect
	github.com/go-playground/validator/v10 v10.30.6 // indirect
	github.com/go-sql-driver/mysql v1.10.2 // indirect
	github.com/goccy/go-json v0.12.0 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.34.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.20.5 // indirect
	github.com/klauspost/cpuid/v2 v2.4.2 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.14.0 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.33 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.65.0 // indirect
	github.com/sony/sonyflake v1.3.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.15.1 // indirect
	github.com/ugorji/go/codec v1.3.2 // indirect
	github.com/xdg-go/scram v1.4.0 // indirect
	github.com/xdg-go/stringprep v1.0.5 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.etcd.io/etcd/api/v3 v3.7.4 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.7.4 // indirect
	go.mongodb.org/mongo-driver/v2 v2.10.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.2 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	golang.org/x/arch v0.33.0 // indirect
	golang.org/x/crypto v0.59.0 // indirect
	golang.org/x/net v0.63.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261118221215-c8830f628b3f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261118221215-c8830f628b3f // indirect
)
