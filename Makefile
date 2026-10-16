GO ?= go
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null)
ifeq ($(GOLANGCI_LINT),)
GOLANGCI_LINT := $(shell $(GO) env GOPATH)/bin/golangci-lint
endif
PROTO_DIR=proto
GEN_DIR=gen/go

PROTO_FILES=$(shell find $(PROTO_DIR) -name "*.proto")

.PHONY: proto generate test run clean vet lint vuln fmt-check check

proto: generate

generate:
	@command -v protoc >/dev/null 2>&1 || (echo "protoc not found. install protobuf compiler first." && exit 1)
	protoc \
		--proto_path=$(PROTO_DIR) \
		--proto_path=$(GOPATH)/pkg/mod \
		--go_out=$(GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_DIR) --grpc-gateway_opt=paths=source_relative \
		$(PROTO_FILES)

test:
	$(GO) test -race ./...

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
