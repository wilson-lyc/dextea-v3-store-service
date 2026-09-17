module github.com/wilson-lyc/dextea-store-service

go 1.27.0

require (
	github.com/go-sql-driver/mysql v1.10.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/wilson-lyc/dextea-v3-proto v0.1.0
	golang.org/x/crypto v0.48.0
	google.golang.org/grpc v1.67.3
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/wilson-lyc/dextea-v3-proto => ../dextea-proto

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
