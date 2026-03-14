# @pgomes13/drift-guard

API schema diff engine — detect breaking changes in **OpenAPI**, **GraphQL**, and **gRPC** schemas.

Thin npm wrapper around the [`drift-guard`](https://github.com/pgomes13/api-drift-engine) Go binary. On install, the correct pre-built binary for your platform is downloaded automatically.

## Installation

```sh
npm install @pgomes13/drift-guard
```

Requires Node.js ≥ 16. The binary is downloaded for your platform (macOS arm64/amd64, Linux arm64/amd64, Windows amd64) during `npm install`.

## CLI

After installing, the `drift-guard` binary is available as an npm bin:

```sh
npx drift-guard --help
```

### OpenAPI

```sh
drift-guard openapi --base old.yaml --head new.yaml
drift-guard openapi --base old.yaml --head new.yaml --format json
drift-guard openapi --base old.yaml --head new.yaml --fail-on-breaking
```

### GraphQL

```sh
drift-guard graphql --base old.graphql --head new.graphql
drift-guard graphql --base old.graphql --head new.graphql --format markdown
```

### gRPC / Protobuf

```sh
drift-guard grpc --base old.proto --head new.proto
drift-guard grpc --base old.proto --head new.proto --format json
```

### Impact analysis

Scan source code for references to each breaking change:

```sh
# From a saved diff JSON
drift-guard openapi --base old.yaml --head new.yaml --format json > diff.json
drift-guard impact --diff diff.json --scan ./src

# Pipe mode
drift-guard openapi --base old.yaml --head new.yaml --format json \
  | drift-guard impact --scan ./src

# Output as markdown
drift-guard impact --diff diff.json --scan ./src --format markdown
```

## Node.js / TypeScript API

```ts
import { compareOpenAPI, compareGraphQL, compareGRPC, impact } from "@pgomes13/drift-guard";

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

// Text or markdown report
const report = impact(result, "./src", { format: "markdown" });
```

### CommonJS

```js
const { compareOpenAPI, impact } = require("@pgomes13/drift-guard");
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
```

## License

MIT
