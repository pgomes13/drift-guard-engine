# api-drift-engine

API type safety across **OpenAPI**, **GraphQL**, and **gRPC** — catch breaking changes before they reach production.

**[Full documentation →](https://pgomes13.github.io/api-drift-engine)**

## Quick install

```sh
# Homebrew
brew tap pgomes13/tap
brew install drift-guard

# npm
npm install @pgomes13/drift-guard
```

## Quick start

```sh
# Auto-generate and compare specs between branches
drift-guard compare

# GitHub Action — one line
- uses: pgomes13/api-drift-engine@v1
```

## npm / Node.js API

```ts
import { compareOpenAPI, impact } from "@pgomes13/drift-guard";

const result = compareOpenAPI("old.yaml", "new.yaml");
const hits = impact(result, "./src");
```
