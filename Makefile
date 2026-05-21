.PHONY: build test test-integration test-all lint proto-gen proto-install proto-lint proto-format clean localnet-build localnet-start localnet-stop localnet-reset localnet-logs help

BUILD_DIR := build
CHAIN_BINARY := wevibe-chain

protoVer=0.18.1
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=docker run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-gen:
	@echo "Generating protobuf files..."
	@$(protoImage) sh ./scripts/protocgen.sh
	@go mod tidy

proto-lint:
	@$(protoImage) buf lint proto/ --error-format=json

proto-format:
	@$(protoImage) find ./proto -name "*.proto" -exec clang-format -i {} \;

build:
	@echo "Building $(CHAIN_BINARY)..."
	@mkdir -p $(BUILD_DIR)
	cd $(BUILD_DIR) && go mod init $(CHAIN_BINARY) 2>/dev/null || true
	go build -o $(BUILD_DIR)/$(CHAIN_BINARY) ./...

test:
	go test -v ./...

test-integration:
	go test ./tests/integration/... -v -count=1 -timeout 120s 2>&1

test-all:
	go test ./... -timeout 120s 2>&1

test-verbose:
	go test -v -count=1 ./...

lint:
	golangci-lint run ./...

proto-install:
	@echo "Installing proto tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/cosmos/gogoproto/protoc-gen-gogo@latest

clean:
	rm -rf $(BUILD_DIR)
	find . -name "*.pb.go" -delete

deps:
	go mod download
	go mod tidy

localnet-build:
	docker compose build

localnet-start:
	docker compose up --build -d
	@echo "Waiting for chain to start..."
	@sleep 10
	@echo "Chain status:"
	@curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height' 2>/dev/null || echo "Not ready yet — check: docker compose logs"

localnet-stop:
	docker compose down

localnet-reset:
	docker compose down -v
	@echo "Chain data wiped."

localnet-logs:
	docker compose logs -f --tail 50

.PHONY: build test test-integration test-all lint proto-gen proto-install proto-lint proto-format clean localnet-build localnet-start localnet-stop localnet-reset localnet-logs help

x/org/module:
	@echo "Building org module..."
	go build ./x/org/...

x/reputation/module:
	@echo "Building reputation module..."
	go build ./x/reputation/...

all-modules: x/org/module x/reputation/module

help:
	@echo "Available targets:"
	@echo "  build           - Build the chain binary"
	@echo "  test            - Run all tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-all        - Run all tests with timeout"
	@echo "  test-verbose    - Run all tests with verbose output"
	@echo "  lint            - Run linter"
	@echo "  proto-gen       - Generate protobuf code"
	@echo "  proto-install   - Install proto generation tools"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Download and tidy dependencies"
	@echo "  localnet-start  - Build and launch the local validator stack"
	@echo "  localnet-stop   - Stop and remove the local validator stack"
