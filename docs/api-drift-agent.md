# API Drift Agent

[![View on GitHub Marketplace](https://img.shields.io/badge/GitHub%20Marketplace-api--drift--agent-blue?logo=github)](https://github.com/marketplace/actions/api-drift-agent)

A LangGraph-powered agent that detects breaking API changes in provider PRs and automatically opens GitHub Issues in affected consumer repos — **zero config on the consumer side**.

No changes are required in consumer repos — `api-drift-agent` scans your entire GitHub org automatically.

## How it works

```
Provider PR opened
       │
       ▼
┌─────────────────────────────────────┐
│  Download drift-guard-engine binary │
│  Compare base ↔ head OpenAPI schema │
└─────────────────────────────────────┘
       │ breaking changes found
       ▼
┌─────────────────────────────────────┐
│  Search org for repos that          │
│  reference affected endpoints       │
└─────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Clone each consumer repo           │
│  Scan for affected files            │
│  Open (or update) a GitHub Issue    │
└─────────────────────────────────────┘
```

## Usage

Add to your **provider** repo's workflow:

```yaml
name: API Drift Check

on:
  pull_request:

permissions:
  contents: read
  issues: write

jobs:
  drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: DriftAgent/api-drift-agent@v1
        with:
          org-read-token: ${{ secrets.ORG_READ_TOKEN }}
```

## Inputs

| Input | Required | Description |
|---|---|---|
| `base-schema` | No | Path to OpenAPI schema (auto-detected if omitted) |
| `head-schema` | No | Path on PR branch (defaults to `base-schema`) |
| `org-read-token` | No | PAT with `repo:read` + `read:org` for private repos |

## Python CLI

The agent is also available as a standalone CLI:

```sh
pip install drift-guard-agent

drift-guard-agent \
  --diff diff.json \
  --org my-org \
  --token $ORG_READ_TOKEN \
  --github-token $GITHUB_TOKEN \
  --pr 42
```

