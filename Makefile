PROTO_DIR=api
GEN_DIR=gen/go

PROTO_FILES=$(shell find $(PROTO_DIR) -name "*.proto")

.PHONY: generate clean start

generate:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--proto_path=$(GOPATH)/pkg/mod \
		--go_out=$(GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_DIR) --grpc-gateway_opt=paths=source_relative \
		$(PROTO_FILES)

clean:
	rm -rf $(GEN_DIR)

start:
	go run ./cmd/server
