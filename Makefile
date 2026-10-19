GO ?= go
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null)
ifeq ($(GOLANGCI_LINT),)
GOLANGCI_LINT := $(shell $(GO) env GOPATH)/bin/golangci-lint
endif
PROTO_DIR=proto
GEN_DIR=gen/go
GOOGLEAPIS_PROTO_DIR=$(shell find "$(shell go env GOMODCACHE)/github.com/grpc-ecosystem" -type d -path '*/third_party/googleapis' -print -quit)

PROTO_FILES=$(shell find $(PROTO_DIR) -name "*.proto")

.PHONY: proto generate test integration run clean vet lint vuln fmt-check check

proto: generate

generate:
	@command -v protoc >/dev/null 2>&1 || (echo "protoc not found. install protobuf compiler first." && exit 1)
	protoc \
		--proto_path=$(PROTO_DIR) \
		--proto_path=$(GOOGLEAPIS_PROTO_DIR) \
		--go_out=$(GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_DIR) --grpc-gateway_opt=paths=source_relative \
		$(PROTO_FILES)

test:
	$(GO) test -race ./...

integration:
	@set -eu; \
	compose_file=deploy/local/docker-compose.yml; \
	project=transaction-history-predeploy; \
	docker compose -p $$project -f $$compose_file down --volumes --remove-orphans >/dev/null 2>&1 || true; \
	trap 'docker compose -p $$project -f $$compose_file down --volumes --remove-orphans' EXIT; \
	docker compose -p $$project -f $$compose_file up -d --wait --wait-timeout 240; \
	TH_SQL_PORT=$${TH_SQL_PORT:-14333} TH_KAFKA_PORT=$${TH_KAFKA_PORT:-29092} $(GO) test -tags=integration -count=1 -v ./test/integration

vet:
	$(GO) vet ./...

lint:
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || (echo "golangci-lint not found. install with: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0" && exit 1)
	$(GOLANGCI_LINT) run

vuln:
	govulncheck ./...

fmt-check:
	test -z "$(shell gofmt -l .)"

check: fmt-check vet test lint vuln

run:
	$(GO) run ./cmd/server

clean:
	rm -rf $(GEN_DIR)

start: run
