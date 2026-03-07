# drift-guard

Detect and classify breaking vs. non-breaking API contract changes across **OpenAPI**, **GraphQL**, and **gRPC** schemas.

**[Full documentation →](https://pgomes13.github.io/drift-guard-engine)**

## Quick install

```sh
brew tap pgomes13/tap
brew install drift-guard
```

## Quick start

```sh
# Auto-generate and compare specs between branches
drift-guard compare

# GitHub Action — one line
- uses: pgomes13/drift-guard-engine@v1
```
