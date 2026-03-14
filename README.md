# DriftaBot

API type safety across **OpenAPI**, **GraphQL**, and **gRPC** — catch breaking changes before they reach production.

**[Full documentation →](https://driftabot.github.io/engine/)**

## Quick install

```sh
# Homebrew
brew tap DriftaBot/tap
brew install driftabot

# npm
npm install @driftabot/engine
```

## Quick start

```sh
# Auto-generate and compare specs between branches
driftabot compare

# GitHub Action — one line
- uses: DriftaBot/engine@v5
```

## npm / Node.js API

```ts
import { compareOpenAPI, impact } from "@driftabot/engine";

const result = compareOpenAPI("old.yaml", "new.yaml");
const hits = impact(result, "./src");
```
