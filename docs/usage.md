# Usage

## Diff two schema files

```sh
drift-agent <command> --base <file> --head <file> [--format <format>] [--fail-on-breaking]
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
drift-agent openapi --base api/base.yaml --head api/head.yaml

# GraphQL — JSON output
drift-agent graphql --base schema/base.graphql --head schema/head.graphql --format json

# gRPC — fail CI on breaking changes
drift-agent grpc --base proto/base.proto --head proto/head.proto --fail-on-breaking
```

## Impact analysis

After detecting breaking changes, use `drift-agent impact` to scan source code and find every file and line that references each breaking change.

```sh
drift-agent <schema-command> --base <file> --head <file> --format json \
  | drift-agent impact --scan <dir>
```

Or from a saved diff file:

```sh
drift-agent openapi --base base.yaml --head head.yaml --format json > diff.json
drift-agent impact --diff diff.json --scan ./src
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
drift-agent openapi --base old.yaml --head new.yaml --format json \
  | drift-agent impact --scan ./services

# Markdown report — collapsible sections per breaking change
drift-agent impact --diff diff.json --scan ./src --format markdown

# JSON output (machine-readable)
drift-agent impact --diff diff.json --scan ./src --format json
```

See [Output Formats](/output-formats) for format details and examples.


