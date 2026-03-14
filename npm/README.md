# @driftabot/engine

API schema diff engine — detect breaking changes in **OpenAPI**, **GraphQL**, and **gRPC** schemas.

Thin npm wrapper around the [`driftabot`](https://github.com/DriftaBot/driftabot-engine) Go binary. On install, the correct pre-built binary for your platform is downloaded automatically.

## Installation

```sh
npm install @driftabot/engine
```

Requires Node.js ≥ 16. The binary is downloaded for your platform (macOS arm64/amd64, Linux arm64/amd64, Windows amd64) during `npm install`.

## CLI

After installing, the `driftabot` binary is available as an npm bin:

```sh
npx driftabot --help
```

### OpenAPI

```sh
driftabot openapi --base old.yaml --head new.yaml
driftabot openapi --base old.yaml --head new.yaml --format json
driftabot openapi --base old.yaml --head new.yaml --fail-on-breaking
```

### GraphQL

```sh
driftabot graphql --base old.graphql --head new.graphql
driftabot graphql --base old.graphql --head new.graphql --format markdown
```

### gRPC / Protobuf

```sh
driftabot grpc --base old.proto --head new.proto
driftabot grpc --base old.proto --head new.proto --format json
```

### Impact analysis

Scan source code for references to each breaking change:

```sh
# From a saved diff JSON
driftabot openapi --base old.yaml --head new.yaml --format json > diff.json
driftabot impact --diff diff.json --scan ./src

# Pipe mode
driftabot openapi --base old.yaml --head new.yaml --format json \
  | driftabot impact --scan ./src

# Output as markdown
driftabot impact --diff diff.json --scan ./src --format markdown

# GitHub Actions annotations (::error / ::warning workflow commands)
driftabot impact --diff diff.json --scan ./src --format github
```

## Node.js / TypeScript API

```ts
import { compareOpenAPI, compareGraphQL, compareGRPC, impact } from "@driftabot/engine";

// Diff two OpenAPI schemas
const result = compareOpenAPI("old.yaml", "new.yaml");
console.log(result.summary);
// { total: 3, breaking: 1, non_breaking: 2, info: 0 }

// Diff two GraphQL schemas
const gqlResult = compareGraphQL("old.graphql", "new.graphql");

// Diff two Protobuf schemas
const grpcResult = compareGRPC("old.proto", "new.proto");

// Scan source for references to breaking changes
const hits = impact(result, "./src");
// Returns Hit[] — file paths and line numbers that reference each breaking change

// Text, markdown, or GitHub Actions annotations
const report = impact(result, "./src", { format: "markdown" });
const ghAnnotations = impact(result, "./src", { format: "github" });
```

### CommonJS

```js
const { compareOpenAPI, impact } = require("@driftabot/engine");
```

## TypeScript types

```ts
type Severity = "breaking" | "non-breaking" | "info";

interface Change {
  type: string;       // e.g. "endpoint_removed", "field_type_changed"
  severity: Severity;
  path: string;       // e.g. "/users/{id}"
  method: string;     // e.g. "DELETE"
  location: string;
  description: string;
  before?: string;
  after?: string;
}

interface Summary {
  total: number;
  breaking: number;
  non_breaking: number;
  info: number;
}

interface DiffResult {
  base_file: string;
  head_file: string;
  changes: Change[];
  summary: Summary;
}

interface Hit {
  file: string;
  line_num: number;
  line: string;
  change_type: string;  // e.g. "endpoint_removed"
  change_path: string;  // e.g. "DELETE /users/{id}"
}

interface ImpactOptions {
  format?: "text" | "json" | "markdown" | "github";
}
```

## License

MIT
