# CI Integration

drift-guard is designed to run in CI as a schema diff gate on pull requests.

## GitHub Action (recommended)

Add API drift detection to any pull request in one line:

```yaml
- uses: pgomes13/drift-guard-engine@v1
```

When API drift is detected, the action automatically posts a PR comment with the full diff report. It auto-detects your framework and runs comparisons for all API types found — REST (OpenAPI), GraphQL, and gRPC. Supported Node.js frameworks: **Express**, **NestJS**. More language and framework support coming soon.

Full workflow example:

```yaml
name: API Drift Check

on:
  pull_request:

permissions:
  contents: read
  pull-requests: write
  issues: write       # required for posting PR comments

jobs:
  drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: pgomes13/drift-guard-engine@v1
```

> **Note:** Both `pull-requests: write` and `issues: write` are required. GitHub's PR comment API uses the Issues REST endpoint, which needs the `issues: write` permission.

### NestJS apps with a database

The action automatically detects database dependencies in `package.json` and starts the appropriate container before spec generation:

| Detected dependency | Container started |
|---|---|
| `typeorm`, `pg`, `@prisma/client`, `sequelize` | `postgres:16-alpine` on port 5432 |
| `mongoose`, `mongodb` | `mongo:7` on port 27017 |

Common connection env vars (`DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME`, `PGHOST`, etc.) are exported automatically. No `services` block needed — the minimal workflow just works:

```yaml
- uses: pgomes13/drift-guard-engine@v1
```

**Custom database config:** If your app uses different env var names or a non-standard connection string, pass them via `env`:

```yaml
- uses: pgomes13/drift-guard-engine@v1
  env:
    DATABASE_URL: postgresql://myuser:mypassword@localhost:5432/mydb
```

When any DB connection env var is already set (`DATABASE_HOST`, `DATABASE_URL`, or `PGHOST`), the action skips auto-starting a container and uses the values you provided instead.

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
| `--format markdown` | Renders a Markdown table — used by the action for PR comments |
| `--format github` | Renders inline PR annotations via workflow commands |
| `--fail-on-breaking` | Exits with code `1` to block merges on breaking changes |
| `--format json` | Use if you need to parse output in a subsequent step |
