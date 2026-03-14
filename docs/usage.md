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
| `-f, --format`       | Output format: `text`, `json`, `markdown`, `github` | `text`   |
| `--fail-on-breaking` | Exit with code `1` if breaking changes are detected | `false`  |

### Examples

```sh
# OpenAPI — text output
drift-guard openapi --base api/base.yaml --head api/head.yaml

# GraphQL — JSON output
drift-guard graphql --base schema/base.graphql --head schema/head.graphql --format json

# gRPC — fail CI on breaking changes
drift-guard grpc --base proto/base.proto --head proto/head.proto --fail-on-breaking
```

## Impact analysis

After detecting breaking changes, use `drift-guard impact` to scan source code and find every file and line that references each breaking change.

```sh
drift-guard <schema-command> --base <file> --head <file> --format json \
  | drift-guard impact --scan <dir>
```

Or from a saved diff file:

```sh
drift-guard openapi --base base.yaml --head head.yaml --format json > diff.json
drift-guard impact --diff diff.json --scan ./src
```

### Flags

| Flag         | Description                                            | Default |
| ------------ | ------------------------------------------------------ | ------- |
| `--diff`     | Path to a JSON diff file; omit or use `-` to read stdin | stdin  |
| `--scan`     | Directory to scan for source references                | `.`     |
| `-f, --format` | Output format: `text`, `json`, `markdown`, `github` | `text`  |

### Examples

```sh
# Pipe OpenAPI diff directly into impact scan
drift-guard openapi --base old.yaml --head new.yaml --format json \
  | drift-guard impact --scan ./services

# Markdown report — collapsible sections per breaking change
drift-guard impact --diff diff.json --scan ./src --format markdown

# JSON output (machine-readable)
drift-guard impact --diff diff.json --scan ./src --format json
```

### Output formats

| Format | Best for |
|--------|----------|
| `text` | Local terminal |
| `markdown` | PR comment — collapsible sections, summary count |
| `json` | Scripting / custom tooling |
| `github` | GitHub Actions — emits `::error` / `::warning` workflow commands for inline PR annotations |

### Sample output (text)

```
Breaking change: DELETE /users/{id} (endpoint_removed)
  services/user-service/client.go:42     client.Delete("/users/" + id)
  apps/mobile-api/routes.go:17           r.DELETE("/users/:id", handler)
```


