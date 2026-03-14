# MCP Server

drift-guard ships an [MCP (Model Context Protocol)](https://modelcontextprotocol.io) server that exposes schema diffing as tools for AI assistants such as Claude Desktop.

## Running the server

```sh
# Build
make build-mcp

# Or run directly
go run ./cmd/mcp-server
```

The server communicates over stdio and is registered in your MCP host's configuration file.

### Claude Desktop configuration

Add drift-guard to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):

```json
{
  "mcpServers": {
    "drift-guard": {
      "command": "/path/to/drift-guard-mcp"
    }
  }
}
```

## Available tools

### `diff_openapi`

Compare two OpenAPI 3.x schema files (YAML or JSON).

| Parameter   | Required | Description                              |
| ----------- | -------- | ---------------------------------------- |
| `base_file` | Yes      | Path to the base (old) schema file       |
| `head_file` | Yes      | Path to the head (new) schema file       |
| `format`    | No       | Output format: `text` (default), `json`, `markdown`, `github` |

### `diff_graphql`

Compare two GraphQL SDL schema files (`.graphql` or `.gql`).

| Parameter   | Required | Description                              |
| ----------- | -------- | ---------------------------------------- |
| `base_file` | Yes      | Path to the base (old) schema file       |
| `head_file` | Yes      | Path to the head (new) schema file       |
| `format`    | No       | Output format: `text` (default), `json`, `markdown`, `github` |

### `diff_grpc`

Compare two Protobuf `.proto` files.

| Parameter   | Required | Description                              |
| ----------- | -------- | ---------------------------------------- |
| `base_file` | Yes      | Path to the base (old) `.proto` file     |
| `head_file` | Yes      | Path to the head (new) `.proto` file     |
| `format`    | No       | Output format: `text` (default), `json`, `markdown`, `github` |

### `detect_project`

Detect the project type, framework, and available schema types for a directory.

| Parameter | Required | Description                              |
| --------- | -------- | ---------------------------------------- |
| `dir`     | Yes      | Absolute path to the project directory   |

Returns the detected framework (e.g. `NestJS`, `Gin`) and whether OpenAPI, GraphQL, and gRPC schemas are present.
