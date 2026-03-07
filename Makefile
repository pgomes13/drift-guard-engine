BIN              := drift-guard
CMD              := ./cmd/drift-guard
HOMEBREW_TAP     := pgomes13/homebrew-tap
FORMULA          := drift-guard

.PHONY: build test vet lint clean run-openapi run-graphql run-grpc release major minor patch gha homebrew commit

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

## Release targets
##
## Usage:
##   make release          # bump patch → tag → push → update floating major tag
##   make release minor    # bump minor → tag → push → update floating major tag
##   make release major    # bump major → tag → push → update floating major tag
##   make release gha      # force-update floating major tag only (no version bump)
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
ifneq (,$(filter gha,$(MAKECMDGOALS)))
	@set -e; \
	LATEST=$$(git describe --tags --abbrev=0 --match "v*.*.*" 2>/dev/null); \
	if [ -z "$$LATEST" ]; then echo "error: no version tag found"; exit 1; fi; \
	FLOAT=$$(echo "$$LATEST" | grep -oE '^v[0-9]+'); \
	echo "Updating floating tag $$FLOAT → $$LATEST"; \
	git tag -f "$$FLOAT"; \
	git push origin "$$FLOAT" --force
else
	@set -e; \
	CURRENT=$$(git describe --tags --abbrev=0 --match "v*.*.*" 2>/dev/null | sed 's/^v//'); \
	if [ -z "$$CURRENT" ]; then \
		echo "error: no semver tag found in repo (expected v<major>.<minor>.<patch>)"; exit 1; \
	fi; \
	echo "Current: v$$CURRENT"; \
	MAJOR=$$(echo "$$CURRENT" | cut -d. -f1); \
	MINOR=$$(echo "$$CURRENT" | cut -d. -f2); \
	PATCH=$$(echo "$$CURRENT" | cut -d. -f3); \
	case "$(_bump)" in \
		major) NEXT="v$$((MAJOR + 1)).0.0" ;; \
		minor) NEXT="v$$MAJOR.$$((MINOR + 1)).0" ;; \
		patch) NEXT="v$$MAJOR.$$MINOR.$$((PATCH + 1))" ;; \
	esac; \
	echo "Next:    $$NEXT  ($(_bump) bump)"; \
	git tag "$$NEXT"; \
	git push origin "$$NEXT"; \
	FLOAT=$$(echo "$$NEXT" | grep -oE '^v[0-9]+'); \
	echo "Updating floating tag $$FLOAT → $$NEXT"; \
	git tag -f "$$FLOAT"; \
	git push origin "$$FLOAT" --force
endif

gha:
	@true
