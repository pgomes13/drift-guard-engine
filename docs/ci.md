# CI Integration

drift-guard is designed to run in CI as a schema diff gate on pull requests.

## GitHub Action (recommended)

Available on the [GitHub Marketplace](https://github.com/marketplace/actions/drift-guard). Add API drift detection to any pull request in one line:

```yaml
- uses: pgomes13/drift-guard-engine@v1
  with:
    node-version: "20" # optional, default: "20"
```

<details>
<summary>Show steps</summary>

```
Pull Request opened
       │
       ▼
┌─────────────────────────────────────┐
│  Detect framework & API types       │
│  Express · NestJS · Gin · Echo …    │
│  OpenAPI · GraphQL · gRPC           │
└─────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Generate schemas                   │
│  head  ←  current branch            │
│  base  ←  origin/main (worktree)    │
└─────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Diff & classify changes            │
│  breaking · non-breaking · info     │
└─────────────────────────────────────┘
       │
       ├──────────────────────────────────────┐
       ▼                                      ▼
┌─────────────────────┐         ┌─────────────────────────┐
│  Post PR comment    │         │  Update GitHub Pages    │
│  Markdown diff      │         │  drift log (per PR)     │
│  table with badges  │         │  with trend chart       │
└─────────────────────┘         └─────────────────────────┘
```

</details>

See [Supported](/supported) for all supported languages and frameworks.

Full workflow example:

```yaml
name: API Drift Check

on:
  pull_request:

permissions:
  contents: write # required for updating the drift log on GitHub Pages
  pull-requests: write
  issues: write # required for posting PR comments

jobs:
  drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: pgomes13/drift-guard-engine@v1
```

> **Note:** `contents: write` is required to update the drift log on your GitHub Pages branch. `issues: write` is required because GitHub's PR comment API uses the Issues REST endpoint.

## Live example

See drift-guard in action on a real pull request: [pgomes13/nest-coffee#8](https://github.com/pgomes13/nest-coffee/pull/8)

## Action inputs

| Input               | Required | Default            | Description                                                                                          |
| ------------------- | -------- | ------------------ | ---------------------------------------------------------------------------------------------------- |
| `node-version`      | No       | `"20"`             | Node.js version used to generate schemas                                                             |
| `upload-diff`       | No       | `"false"`          | Upload the JSON diff as a `drift-guard-diff` artifact — required for [microservice consumer checks](/microservices) |
| `notify-consumers`  | No       | `""`               | Comma-separated list of repos to notify via `repository_dispatch` (e.g. `org/service-b,org/service-c`) |
| `notify-token`      | No       | `""`               | GitHub token with `repo` write access to consumer repos — required when `notify-consumers` is set   |
| `service-url`       | No       | `""`               | Drift Guard service URL (paid tier) — see [Connected mode](#connected-mode-paid-tier) below          |
| `project-token`     | No       | `github.token`     | Token used to authenticate with the Drift Guard service (paid tier)                                  |

## Action outputs

| Output          | Description                                                             |
| --------------- | ----------------------------------------------------------------------- |
| `drift-status`  | `no_drift`, `drift_detected`, `auto_accepted`, or `error`               |
| `check-run-id`  | GitHub Check Run ID created by the service (connected mode only)        |
| `report-url`    | URL to the drift report in the portal (connected mode only)             |

## Standalone mode (default)

By default the action runs entirely within your repository using the drift-guard CLI binary. It generates schemas from your source code, diffs them, posts a Markdown PR comment, and optionally updates a drift log on GitHub Pages — no external service required.

## Connected mode (paid tier)

When `service-url` is provided the action delegates analysis to the Drift Guard service. The service manages GitHub Check Runs, baseline storage, and portal-based approve/reject review.

```yaml
- uses: pgomes13/drift-guard-engine@v1
  with:
    service-url: https://your-drift-guard-service.example.com
    project-token: ${{ secrets.DRIFT_GUARD_TOKEN }}
```

When `service-url` is set and the GitHub App installation ID is available, the action sends a `POST /drift/{owner}/{repo}/pulls/{pr}/analyze` request to the service and waits for the Check Run result. If the service is unreachable it falls back to standalone mode automatically.

## Key flags

| Flag                 | Purpose                                                       |
| -------------------- | ------------------------------------------------------------- |
| `--format markdown`  | Renders a Markdown table — used by the action for PR comments |
| `--format github`    | Renders inline PR annotations via workflow commands           |
| `--fail-on-breaking` | Exits with code `1` to block merges on breaking changes       |
| `--format json`      | Use if you need to parse output in a subsequent step          |
