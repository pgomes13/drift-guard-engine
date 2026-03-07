BIN              := drift-guard
CMD              := ./cmd/drift-guard
HOMEBREW_TAP     := pgomes13/homebrew-tap
FORMULA          := drift-guard

.PHONY: build test vet lint clean run-openapi run-graphql run-grpc release major minor patch commit

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

## Release: bump major, minor, or patch version based on the current homebrew
## tap formula, then tag and push.
##
## Usage:
##   make release          # default: bump patch
##   make release minor
##   make release major
##   make release patch
##
## Requires: gh CLI (https://cli.github.com) authenticated with repo access.
ifneq (,$(filter major,$(MAKECMDGOALS)))
  _bump := major
else ifneq (,$(filter minor,$(MAKECMDGOALS)))
  _bump := minor
else
  _bump := patch
endif

major minor patch:
	@true

release:
	@command -v gh >/dev/null 2>&1 || { echo "error: gh CLI not found — install from https://cli.github.com"; exit 1; }
	@set -e; \
	echo "Fetching current version from $(HOMEBREW_TAP)..."; \
	RAW=$$(gh api "repos/$(HOMEBREW_TAP)/contents/$(FORMULA).rb" --jq '.content' | base64 -d); \
	CURRENT=$$(echo "$$RAW" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
	if [ -z "$$CURRENT" ]; then \
		echo "error: could not parse version from $(FORMULA).rb in $(HOMEBREW_TAP)"; exit 1; \
	fi; \
	MAJOR=$$(echo "$$CURRENT" | cut -d. -f1); \
	MINOR=$$(echo "$$CURRENT" | cut -d. -f2); \
	PATCH=$$(echo "$$CURRENT" | cut -d. -f3); \
	case "$(_bump)" in \
		major) NEXT="v$$((MAJOR + 1)).0.0" ;; \
		minor) NEXT="v$$MAJOR.$$((MINOR + 1)).0" ;; \
		patch) NEXT="v$$MAJOR.$$MINOR.$$((PATCH + 1))" ;; \
	esac; \
	echo "Current: v$$CURRENT  →  Next: $$NEXT  ($(_bump) bump)"; \
	git tag "$$NEXT"; \
	git push origin "$$NEXT"
