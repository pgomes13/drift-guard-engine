# drift-guard-diff-engine

A schema diff engine that detects and classifies breaking vs. non-breaking API contract changes across **OpenAPI**, **GraphQL**, and **gRPC** schemas.

## Features

- Parses OpenAPI 3.x (YAML/JSON), GraphQL SDL, and Protobuf (`.proto`) schemas
- Produces a flat, structured list of changes with severity classification
- Three output formats: `text`, `json`, `github` (GitHub Actions annotations)
- `--fail-on-breaking` exit code for CI gates
- Fully tested pipeline: parser → differ → classifier → reporter

## Install

```sh
go build -o drift-guard ./cmd/drift-guard
```

Or via Make:

```sh
make build
```

## Usage

```sh
drift-guard <command> --base <file> --head <file> [--format <format>] [--fail-on-breaking]
```

| Command | Description |
|---|---|
| `openapi` | Diff two OpenAPI 3.x schemas (YAML or JSON) |
| `graphql` | Diff two GraphQL SDL schemas |
| `grpc` | Diff two Protobuf schemas (`.proto`) |

| Flag | Description | Default |
|---|---|---|
| `--base` | Path to the base (before) schema file | required |
| `--head` | Path to the head (after) schema file | required |
| `-f, --format` | Output format: `text`, `json`, `github` | `text` |
| `--fail-on-breaking` | Exit with code `1` if breaking changes are detected | `false` |

### Examples

```sh
# OpenAPI — text output
drift-guard openapi --base api/base.yaml --head api/head.yaml

# GraphQL — JSON output
drift-guard graphql --base schema/base.graphql --head schema/head.graphql --format json

# gRPC — fail CI on breaking changes
drift-guard grpc --base proto/base.proto --head proto/head.proto --fail-on-breaking

# GitHub Actions annotations
drift-guard openapi --base base.yaml --head head.yaml --format github
```

## Output formats

### `text`

```
Schema Diff: base.yaml → head.yaml
Total: 4  Breaking: 2  Non-Breaking: 1  Info: 1

SEVERITY        TYPE                    PATH            METHOD  LOCATION        DESCRIPTION
----------------------------------------------------------------------------------------------------
[BREAKING]      endpoint_removed        /users/{id}     DELETE                  Endpoint '/users/{id}' method DELETE was removed
[BREAKING]      param_type_changed      /users/{id}     GET     path.id         Param 'id' type changed from 'string' to 'integer'
[non-breaking]  endpoint_added          /posts                                  Endpoint '/posts' was added
[info]          field_added             /users          POST    request.role    Field 'role' was added
```

### `json`

```json
{
  "base_file": "base.yaml",
  "head_file": "head.yaml",
  "changes": [
    {
      "type": "endpoint_removed",
      "severity": "breaking",
      "path": "/users/{id}",
      "method": "DELETE",
      "location": "",
      "description": "Endpoint '/users/{id}' method DELETE was removed"
    }
  ],
  "summary": {
    "total": 4,
    "breaking": 2,
    "non_breaking": 1,
    "info": 1
  }
}
```

### `github`

Emits GitHub Actions [workflow commands](https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions) that render as inline annotations on the PR diff:

```
::error title=Breaking Change::Endpoint '/users/{id}' method DELETE was removed
::warning title=Non-Breaking Change::Endpoint '/posts' was added
::error title=API Contract Violation::2 breaking change(s) detected between base.yaml and head.yaml
```

## Severity rules

### OpenAPI

| Change | Severity |
|---|---|
| Endpoint / method removed | breaking |
| Endpoint / method added | non-breaking |
| Parameter removed | breaking |
| Parameter added | non-breaking |
| Parameter type changed | breaking |
| Parameter required: optional → required | breaking |
| Parameter required: required → optional | non-breaking |
| Request body removed | breaking |
| Response code removed | breaking |
| Field removed | breaking |
| Field added | non-breaking |
| Field type changed | breaking |
| Field required: optional → required | breaking |

### GraphQL

| Change | Severity |
|---|---|
| Type removed | breaking |
| Type added | non-breaking |
| Type kind changed (e.g. Object → Interface) | breaking |
| Output field removed | breaking |
| Output field added | non-breaking |
| Output field deprecated | info |
| Output field type: non-null → nullable (`T!` → `T`) | breaking |
| Output field type: nullable → non-null (`T` → `T!`) | non-breaking |
| Argument removed | breaking |
| Argument added (required) | breaking |
| Argument added (optional) | non-breaking |
| Enum value removed | breaking |
| Enum value added | non-breaking |
| Union member removed | breaking |
| Union member added | non-breaking |
| Input field removed | breaking |
| Input field added (required) | breaking |
| Input field added (optional) | non-breaking |
| Interface removed from type | breaking |
| Interface added to type | non-breaking |

### gRPC

| Change | Severity |
|---|---|
| Service removed | breaking |
| Service added | non-breaking |
| RPC removed | breaking |
| RPC added | non-breaking |
| RPC request type changed | breaking |
| RPC response type changed | breaking |
| RPC streaming mode changed | breaking |
| Message removed | breaking |
| Message added | non-breaking |
| Field removed | breaking |
| Field added | non-breaking |
| Field type changed | breaking |
| Field number changed | breaking |
| Field label changed (singular ↔ repeated) | breaking |

## Development

```sh
make build       # compile binary
make test        # run all tests
make vet         # run go vet
make lint        # run go vet + staticcheck
make clean       # remove binary

make run-openapi  # build and diff bundled OpenAPI fixtures
make run-graphql  # build and diff bundled GraphQL fixtures
```

## Architecture

```
cmd/drift-guard/          # CLI entry point
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

## CI

Two GitHub Actions workflows are included:

- **`ci.yml`** — runs `make build`, `make test`, `make vet` on every push and pull request
- **`pull_request.yml`** — auto-generates a PR description from commit messages as bullet points
