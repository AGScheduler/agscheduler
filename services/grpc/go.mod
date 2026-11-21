module github.com/agscheduler/agscheduler/services/grpc

go 1.27.4

require (
	github.com/agscheduler/agscheduler v0.20.0
	github.com/agscheduler/agscheduler/backends/memory v0.20.0
	github.com/agscheduler/agscheduler/queues/memory v0.20.0
	github.com/agscheduler/agscheduler/services/proto v0.20.0
	github.com/agscheduler/agscheduler/stores/memory v0.20.0
	github.com/stretchr/testify v1.12.1
	google.golang.org/grpc v1.85.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75 // indirect
	github.com/sony/sonyflake v1.3.0 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/net v0.63.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261118221215-c8830f628b3f // indirect
)
