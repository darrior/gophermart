# Run `gophermart` with provided args.
run *args:
    go run ./cmd/gophermart/ {{ args }}

# Build `gophermart`.
build:
    go build -o build/gophermart ./cmd/gophermart/

# Run `golangci-lint`. Important: golangci-lint must be in `PATH`.
lint:
    golangci-lint run ./...    

# Run go test ./...
test:
    go test ./...

# Calculate test coverage.
cover path='./...':
    @go test {{ path }} -coverprofile /tmp/cover.out > /dev/null
    @cat /tmp/cover.out | grep -v 'mock_.*\.go' | tee /tmp/cover.out > /dev/null
    go tool cover -func /tmp/cover.out
    @rm /tmp/cover.out

# Calculate test coverage in HTML.
cover-html path='./...':
    @go test {{ path }} -coverprofile /tmp/cover.out > /dev/null
    @cat /tmp/cover.out | grep -v 'mock_.*\.go' | tee /tmp/cover.out > /dev/null
    @go tool cover -html /tmp/cover.out
    @rm /tmp/cover.out

# Create new SQL migration.
[working-directory('migrations')]
new-migration name:
    goose create {{ name }} sql

# [Re]generate mocks for `gophermart`'s interfaces.
generate-mocks: _generate-repository-mock _generate-accrual-mock _generate-service-mock _generate-client-mock

_generate-repository-mock: (_generate-mock "mock_repository.go" "internal/service" "Repository")

_generate-accrual-mock: (_generate-mock "mock_accrual.go" "internal/service" "AccrualSystem")

_generate-service-mock: (_generate-mock "mock_service.go" "internal/handlers" "Service")

_generate-client-mock: (_generate-mock "mock_client.go" "internal/gateways/accrual" "client")

# Common generate-XXX-mock implementation
_generate-mock dest-file package interface:
    go tool mockgen -destination=internal/mocks/{{ dest-file }} -package=mocks github.com/darrior/gophermart/{{ package }} {{ interface }}
