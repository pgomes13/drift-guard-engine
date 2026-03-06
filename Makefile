BIN := diffengine
CMD := ./cmd/diffengine

.PHONY: build test vet lint clean run-openapi run-graphql

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
	./$(BIN) --base internal/testdata/base.yaml --head internal/testdata/head.yaml --type openapi --format text

run-graphql: build
	./$(BIN) --base internal/testdata/base.graphql --head internal/testdata/head.graphql --type graphql --format text
