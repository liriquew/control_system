module github.com/liriquew/control_system/groups_service

go 1.24.1

require (
	github.com/fatih/color v1.18.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.1
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/liriquew/control_system/services_protos v0.0.0
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.4
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

replace github.com/liriquew/control_system/services_protos => ../service_protos
