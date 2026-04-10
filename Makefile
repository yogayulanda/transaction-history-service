GO ?= go
PROTO_DIR=proto
GEN_DIR=gen/go

PROTO_FILES=$(shell find $(PROTO_DIR) -name "*.proto")

.PHONY: proto generate test run clean

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
	$(GO) test ./...

run:
	$(GO) run ./cmd/server

clean:
	rm -rf $(GEN_DIR)

start: run
