# npm SDK

<a href="https://www.npmjs.com/package/@drift-agent/api-drift-engine" target="_blank">@drift-agent/api-drift-engine</a> is a thin npm wrapper around the drift-agent binary. On install, the correct pre-built binary for your platform is downloaded automatically — no Go toolchain required.

## Installation

```sh
npm install @drift-agent/api-drift-engine
```

Requires Node.js ≥ 16. Supported platforms: macOS arm64/amd64, Linux arm64/amd64, Windows amd64.

## Programmatic API

### OpenAPI

```ts
import { compareOpenAPI } from "@drift-agent/api-drift-engine";

const result = compareOpenAPI("old.yaml", "new.yaml");

console.log(result.summary);
// { total: 3, breaking: 1, non_breaking: 2, info: 0 }

for (const change of result.changes) {
  console.log(`[${change.severity}] ${change.description}`);
}
```

### GraphQL

```ts
import { compareGraphQL } from "@drift-agent/api-drift-engine";

const result = compareGraphQL("old.graphql", "new.graphql");
```

### gRPC / Protobuf

```ts
import { compareGRPC } from "@drift-agent/api-drift-engine";

const result = compareGRPC("old.proto", "new.proto");
```

### Impact analysis

Scan source code for references to each breaking change:

```ts
import { compareOpenAPI, impact } from "@drift-agent/api-drift-engine";

const result = compareOpenAPI("old.yaml", "new.yaml");

// Returns Hit[] — one entry per matching file:line
const hits = impact(result, "./src");

for (const hit of hits) {
  console.log(`${hit.file}:${hit.line_num}  (${hit.change_path})`);
}
```

Text, markdown, or GitHub Actions report:

```ts
const report = impact(result, "./src", { format: "markdown" });
console.log(report);

// GitHub Actions workflow commands for inline PR annotations
const ghReport = impact(result, "./src", { format: "github" });
console.log(ghReport);
```

## CLI via npx

The `drift-agent` binary is available as an npm bin after install:

```sh
npx drift-agent openapi --base old.yaml --head new.yaml
npx drift-agent graphql --base old.graphql --head new.graphql --format json
npx drift-agent impact --diff diff.json --scan ./src
```

## CommonJS

```js
const { compareOpenAPI, impact } = require("@drift-agent/api-drift-engine");
```

## TypeScript types

```ts
type Severity = "breaking" | "non-breaking" | "info";

interface Change {
  type: string;        // e.g. "endpoint_removed", "field_type_changed"
  severity: Severity;
  path: string;        // e.g. "/users/{id}"
  method: string;      // e.g. "DELETE"
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
