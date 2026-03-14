# Development

## Make commands

```sh
make build       # compile CLI binary
make build-mcp   # compile MCP server binary
make test        # run all tests
make vet         # run go vet
make lint        # run go vet + staticcheck
make clean       # remove binary

make run-openapi  # build and diff bundled OpenAPI fixtures
make run-graphql  # build and diff bundled GraphQL fixtures
make run-grpc     # build and diff bundled gRPC fixtures
```

## Architecture

```
cmd/drift-agent/          # CLI entry point (drift-agent binary)
cmd/server/               # gRPC server entry point
cmd/mcp-server/           # MCP server entry point (AI/LLM integration)
api/drift-agent/v1/       # Protobuf service definition & generated Go code
internal/
  parser/
    openapi/             # OpenAPI YAML/JSON → schema.Schema
    graphql/             # GraphQL SDL → schema.GQLSchema
    grpc/                # Protobuf .proto → schema.GRPCSchema
  differ/
    openapi/             # Diffs two schema.Schema values
    graphql/             # Diffs two schema.GQLSchema values
    grpc/                # Diffs two schema.GRPCSchema values
  classifier/
    openapi/             # Assigns severity to OpenAPI changes
    graphql/             # Assigns severity to GraphQL changes
    grpc/                # Assigns severity to gRPC changes
  reporter/              # Renders DiffResult as text / JSON / GitHub annotations
pkg/schema/
  types.schema.go        # Shared types: Change, DiffResult, Severity
  openapi.schema.go      # OpenAPI types and change type constants
  graphql.schema.go      # GraphQL types and change type constants
  grpc.schema.go         # gRPC types and change type constants
```

## Releasing a new version

```sh
make release          # bump patch
make release minor    # bump minor
make release major    # bump major

make release gha      # re-sync floating major tag only (no version bump)
```

Each `make release` call bumps the version from the latest local git tag, pushes the semver tag, then force-updates the floating major tag (e.g. `v2`). Pushing the semver tag triggers the `release.yml` workflow which cross-compiles binaries via [GoReleaser](https://goreleaser.com), publishes a GitHub Release, updates the Homebrew formula in [`DriftaBot/homebrew-tap`](https://github.com/DriftaBot/homebrew-tap), and keeps the floating tag in sync for GitHub Action users.
