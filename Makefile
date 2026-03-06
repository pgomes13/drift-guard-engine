BIN := drift-guard
CMD := ./cmd/diffengine

.PHONY: build test vet lint clean run-openapi run-graphql run-grpc

build:
	go build -o $(BIN) $(CMD)

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	staticcheck ./...

clean:
	rm -f $(BIN)

## Quick smoke runs against the bundled fixtures
run-openapi: build
	./$(BIN) openapi --base internal/testdata/base.yaml --head internal/testdata/head.yaml

run-graphql: build
	./$(BIN) graphql --base internal/testdata/base.graphql --head internal/testdata/head.graphql

run-grpc: build
	./$(BIN) grpc --base internal/testdata/base.proto --head internal/testdata/head.proto
