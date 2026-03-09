BIN              := drift-guard
MCP_BIN          := drift-guard-mcp
API_BIN          := drift-guard-api
CMD              := ./cmd/drift-guard
MCP_CMD          := ./cmd/mcp-server
API_CMD          := ./cmd/playground
HOMEBREW_TAP     := pgomes13/homebrew-tap
FORMULA          := drift-guard

.PHONY: build build-mcp build-api test vet lint clean run-openapi run-graphql run-grpc run-api release major minor patch homebrew commit

build:
	go build -o $(BIN) $(CMD)

build-mcp:
	go build -o $(MCP_BIN) $(MCP_CMD)

build-api:
	go build -o $(API_BIN) $(API_CMD)

run-api: build-api
	./$(API_BIN)

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	staticcheck ./...

clean:
	rm -f $(BIN) $(MCP_BIN) $(API_BIN)

## Quick smoke runs against the bundled fixtures
run-openapi: build
	./$(BIN) openapi --base internal/testdata/base.yaml --head internal/testdata/head.yaml

run-graphql: build
	./$(BIN) graphql --base internal/testdata/base.graphql --head internal/testdata/head.graphql

run-grpc: build
	./$(BIN) grpc --base internal/testdata/base.proto --head internal/testdata/head.proto

## Commit: stage all changes, commit with a message, and push to the current branch.
##
## Usage:
##   make commit   # prompts for commit message
##
commit:
	@read -p "Commit message: " msg; \
	git add .; \
	git commit -m "$$msg"; \
	git push origin $$(git rev-parse --abbrev-ref HEAD)

## Release targets
##
## Usage:
##   make release          # bump patch → tag → push
##   make release minor    # bump minor → tag → push
##   make release major    # bump major → tag → push
##
## Pushing the semver tag triggers the release.yml workflow (goreleaser + Homebrew update).
ifneq (,$(filter major,$(MAKECMDGOALS)))
  _bump := major
else ifneq (,$(filter minor,$(MAKECMDGOALS)))
  _bump := minor
else
  _bump := patch
endif

major minor patch homebrew:
	@true

release:
	@set -e; \
	_CURRENT_TAG=$$(git tag --list 'v*.*.*' --points-at HEAD --sort=-version:refname | head -1); \
	if [ -n "$$_CURRENT_TAG" ]; then \
		echo "error: HEAD is already tagged as $$_CURRENT_TAG — commit new changes before releasing."; \
		exit 1; \
	fi; \
	_TAG=$$(git log --decorate=short --pretty=format:"%D" | \
	  while IFS= read -r _line; do \
	    _t=$$(printf '%s' "$$_line" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | sort -t. -k1,1 -k2,2n -k3,3n | tail -1); \
	    if [ -n "$$_t" ]; then printf '%s\n' "$$_t"; break; fi; \
	  done); \
	if [ -z "$$_TAG" ]; then \
		echo "error: no semver tag found in repo (expected v<major>.<minor>.<patch>)"; exit 1; \
	fi; \
	CURRENT=$$(echo "$$_TAG" | sed 's/^v//'); \
	echo "Current: v$$CURRENT"; \
	MAJOR=$$(echo "$$CURRENT" | cut -d. -f1); \
	MINOR=$$(echo "$$CURRENT" | cut -d. -f2); \
	PATCH=$$(echo "$$CURRENT" | cut -d. -f3); \
	case "$(_bump)" in \
		major) NEXT="v$$((MAJOR + 1)).0.0" ;; \
		minor) NEXT="v$$MAJOR.$$((MINOR + 1)).0" ;; \
		patch) NEXT="v$$MAJOR.$$MINOR.$$((PATCH + 1))" ;; \
	esac; \
	echo "Next:    $$NEXT"; \
	git tag -f "$$NEXT"; \
	git push origin "$$NEXT" --force; \
	MAJOR_VER=$$(echo "$$NEXT" | grep -oE '^v[0-9]+'); \
	git tag -f "$$MAJOR_VER"; \
	git push origin "$$MAJOR_VER" --force; \
	echo "Floating tag updated: $$MAJOR_VER -> $$NEXT"
