# api-drift-engine

API type safety across **OpenAPI**, **GraphQL**, and **gRPC** — catch breaking changes before they reach production.

**[Full documentation →](https://driftagent.github.io/api-drift-engine)**

## Quick install

```sh
# Homebrew
brew tap DriftAgent/tap
brew install drift-guard

# npm
npm install @drift-agent/api-drift-engine
```

## Quick start

```sh
# Auto-generate and compare specs between branches
drift-guard compare

# GitHub Action — one line
- uses: DriftAgent/api-drift-engine@v1
```

## npm / Node.js API

```ts
import { compareOpenAPI, impact } from "@drift-agent/api-drift-engine";

const result = compareOpenAPI("old.yaml", "new.yaml");
const hits = impact(result, "./src");
```
