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

If your NestJS app connects to a database on startup (e.g. via TypeORM or Mongoose), the action needs a live database service to generate the OpenAPI spec. Add a `services` block to your job:

```yaml
name: API Drift Check

on:
  pull_request:

permissions:
  contents: read
  pull-requests: write
  issues: write

jobs:
  drift:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: mydb
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: pgomes13/drift-guard-engine@v1
        env:
          DATABASE_HOST: localhost
          DATABASE_PORT: 5432
          DATABASE_USER: postgres
          DATABASE_PASSWORD: postgres
          DATABASE_NAME: mydb
```

Pass the database connection values as `env` variables matching what your app reads from the environment (e.g. `DATABASE_HOST`, `DB_URL`).

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
