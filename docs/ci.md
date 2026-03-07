# CI Integration

drift-guard is designed to run in CI as a schema diff gate on pull requests.

## GitHub Action (recommended)

Add API drift detection to any pull request in one line:

```yaml
- uses: pgomes13/drift-guard-engine@v1
```

When API drift is detected, the action automatically posts a PR comment with the full diff report. Supported Node.js frameworks: **Express**, **NestJS**. More language and framework support coming soon.

Full workflow example:

```yaml
name: API Drift Check

on:
  pull_request:

permissions:
  contents: read
  pull-requests: write

jobs:
  drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: pgomes13/drift-guard-engine@v1
```

### Action inputs

| Input | Description | Default |
|---|---|---|
| `node-version` | Node.js version for spec generation | `20` |

## Manual install + diff

For diffing two existing schema files directly:

```yaml
- name: Install drift-guard
  run: |
    brew tap pgomes13/tap
    brew install drift-guard

- name: Check for breaking API changes
  run: |
    drift-guard openapi \
      --base api/base.yaml \
      --head api/head.yaml \
      --format github \
      --fail-on-breaking
```

## Key flags

| Flag | Purpose |
|---|---|
| `--format github` | Renders inline PR annotations via workflow commands |
| `--fail-on-breaking` | Exits with code `1` to block merges on breaking changes |
| `--format json` | Use if you need to parse output in a subsequent step |
