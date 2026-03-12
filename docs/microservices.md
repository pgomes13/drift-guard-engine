# Microservices Impact — CI Broadcast

When a provider service changes its API, consumer services need to know which of their files and lines reference the broken endpoints or types.

drift-guard implements this with a **CI broadcast** pattern: the provider generates a diff JSON and notifies consumers via `repository_dispatch`. Each consumer scans itself independently.

```
Provider PR opened
       │
       ▼
┌──────────────────────────────────────────┐
│  drift-guard detects breaking changes    │
│  → uploads diff.json artifact            │
│  → sends repository_dispatch to          │
│    service-b, service-c, service-d       │
└──────────────────────────────────────────┘
       │           │           │
       ▼           ▼           ▼
  service-b    service-c    service-d
  scans ./src  scans ./src  scans ./src
  posts hits   posts hits   no hits
  to own PR    to own PR
```

## Provider setup

In the provider service, add `upload-diff`, `notify-consumers`, and `notify-token` to the drift-guard action:

```yaml
# .github/workflows/api-drift.yml (provider service)
name: API Drift Check

on:
  pull_request:

permissions:
  contents: write
  pull-requests: write
  issues: write

jobs:
  drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: pgomes13/drift-guard-engine@v3
        with:
          upload-diff: "true"
          notify-consumers: "org/service-b,org/service-c"
          notify-token: ${{ secrets.CONSUMER_DISPATCH_TOKEN }}
```

`CONSUMER_DISPATCH_TOKEN` must be a GitHub PAT (or fine-grained token) with `contents: write` on each consumer repo.

## Consumer setup

Each consumer service adds a workflow that fires on `repository_dispatch`:

```yaml
# .github/workflows/impact-check.yml (consumer service)
name: Impact Check

on:
  repository_dispatch:
    types: [drift-guard-impact]

permissions:
  pull-requests: write
  issues: write

jobs:
  impact:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: pgomes13/drift-guard-engine/action/impact-check@v3
        with:
          diff-json: ${{ toJson(github.event.client_payload.diff) }}
          scan-dir: "./src"
          format: "markdown"
          fail-on-hits: "false"
```

### How the consumer notifies the team

The action handles three situations automatically:

| Consumer state | What happens |
|----------------|--------------|
| Has open PR(s) | Impact report posted as a PR comment on every open PR |
| No open PRs | A GitHub Issue is opened (or updated) with the full impact report and a `drift-guard` label |
| No hits found | Nothing posted — job exits cleanly |

This means the team is always notified, whether or not they have work in progress.

> **Note:** `issues: write` is required for both PR comments (GitHub's PR comment API uses the Issues endpoint) and for opening issues.

## Artifact-only mode (no dispatch)

If you prefer polling over push, consumers can download the diff artifact directly instead of using `repository_dispatch`:

```yaml
# Consumer — polls provider artifact instead of dispatch
- uses: pgomes13/drift-guard-engine/action/impact-check@v3
  with:
    provider-repo: "org/user-service"
    scan-dir: "./src"
    token: ${{ secrets.PROVIDER_READ_TOKEN }}
```

`PROVIDER_READ_TOKEN` needs `actions:read` on the provider repo.

## Impact check inputs

| Input | Description | Default |
|-------|-------------|---------|
| `provider-repo` | Provider repo to download artifact from | — |
| `diff-json` | Inline JSON diff (from `repository_dispatch` payload) | — |
| `scan-dir` | Directory to scan for source references | `.` |
| `format` | Output format: `text`, `json`, `markdown`, `github` | `github` |
| `fail-on-hits` | Exit 1 if any references found | `false` |
| `token` | Token with `actions:read` on provider repo | `GITHUB_TOKEN` |

## Provider action inputs

| Input | Description | Default |
|-------|-------------|---------|
| `upload-diff` | Upload diff JSON as `drift-guard-diff` artifact | `false` |
| `notify-consumers` | Comma-separated list of repos to notify | — |
| `notify-token` | Token with `contents:write` on consumer repos | — |
