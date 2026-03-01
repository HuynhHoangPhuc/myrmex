module github.com/HuynhHoangPhuc/myrmex/services/module-student

go 1.26

require (
	github.com/HuynhHoangPhuc/myrmex/gen/go v0.0.0
	github.com/HuynhHoangPhuc/myrmex/pkg v0.0.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.8.0
	github.com/nats-io/nats.go v1.49.0
	github.com/spf13/viper v1.21.0
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
)

require github.com/jung-kurt/gofpdf v1.16.2 // indirect

replace (
	github.com/HuynhHoangPhuc/myrmex/gen/go => ../../gen/go
	github.com/HuynhHoangPhuc/myrmex/pkg => ../../pkg
)
