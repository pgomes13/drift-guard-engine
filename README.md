# driftabot-engine

API type safety across **OpenAPI**, **GraphQL**, and **gRPC** — catch breaking changes before they reach production.

**[Full documentation →](https://driftabot.github.io/driftabot-engine)**

## Quick install

```sh
# Homebrew
brew tap DriftaBot/tap
brew install drift-agent

# npm
npm install @driftabot/engine
```

## Quick start

```sh
# Auto-generate and compare specs between branches
drift-agent compare

# GitHub Action — one line
- uses: DriftaBot/driftabot-engine@v1
```

## npm / Node.js API

```ts
import { compareOpenAPI, impact } from "@driftabot/engine";

const result = compareOpenAPI("old.yaml", "new.yaml");
const hits = impact(result, "./src");
```
