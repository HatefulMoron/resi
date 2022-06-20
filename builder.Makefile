PKG_NAME := github.com/hatefulmoron/resi

PROTOS=$(shell find protos -name \*.proto)


.PHONY: protos
protos:
	protoc -I. --go_out=,paths=source_relative:. --go-grpc_out=,paths=source_relative:. ${PROTOS}

.PHONY: lint
lint:
	golangci-lint run --timeout 10m0s ./...
