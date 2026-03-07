# Usage

## Diff two schema files

```sh
drift-guard <command> --base <file> --head <file> [--format <format>] [--fail-on-breaking]
```

| Command   | Description                                 |
| --------- | ------------------------------------------- |
| `openapi` | Diff two OpenAPI 3.x schemas (YAML or JSON) |
| `graphql` | Diff two GraphQL SDL schemas                |
| `grpc`    | Diff two Protobuf schemas (`.proto`)        |

### Flags

| Flag                 | Description                                         | Default  |
| -------------------- | --------------------------------------------------- | -------- |
| `--base`             | Path to the base (before) schema file               | required |
| `--head`             | Path to the head (after) schema file                | required |
| `-f, --format`       | Output format: `text`, `json`, `github`             | `text`   |
| `--fail-on-breaking` | Exit with code `1` if breaking changes are detected | `false`  |

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

## Compare branches automatically

`drift-guard compare` auto-detects your project type, generates OpenAPI specs for the current branch and the base branch (`origin/main` / `origin/master`), and diffs them.

```sh
drift-guard compare
```

Supported project types: **Go**, **Express**, **NestJS**, **Node.js**.
