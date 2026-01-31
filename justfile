# Run `gophermart` with provided args.
run *args:
    go run ./cmd/gophermart/ {{args}}

# Build `gophermart`.
build:
    go build -o build/gophermart ./cmd/gophermart/

# Run `golangci-lint`. Important: golangci-lint must be in `PATH`.
lint:
    golangci-lint run ./...    

# Run go test ./...
test:
    go test ./...

# Create new SQL migration.
[working-directory: 'migrations']
new-migration name:
    goose create {{name}} sql

# [Re]generate mocks for `gophermart`'s interfaces.
generate-mocks: _generate-repository-mock _generate-accrual-mock _generate-service-mock _generate-client-mock

_generate-repository-mock: (_generate-mock "mock_repository.go" "internal/service" "Repository")
    
_generate-accrual-mock: (_generate-mock "mock_accrual.go" "internal/service" "AccrualSystem")

_generate-service-mock: (_generate-mock "mock_service.go" "internal/handlers" "Service")

_generate-client-mock: (_generate-mock "mock_client.go" "internal/gateways/accrual" "client")

# Common generate-XXX-mock implementation
_generate-mock dest-file package interface:
    go tool mockgen -destination=internal/mocks/{{dest-file}} -package=mocks github.com/darrior/gophermart/{{package}} {{interface}}
