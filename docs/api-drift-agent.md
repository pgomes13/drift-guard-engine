# API Drift Agent

[![View on GitHub Marketplace](https://img.shields.io/badge/GitHub%20Marketplace-api--drift--agent-blue?logo=github)](https://github.com/marketplace/actions/api-drift-agent)

> **Recommended integration.** The API Drift Agent is the recommended way to solve API drift at scale. Rather than wiring up the engine manually, the agent handles discovery, analysis, and consumer notification automatically.

`api-drift-agent` is a LangGraph-powered agentic workflow that detects breaking API changes in provider PRs and automatically opens GitHub Issues in affected consumer repos — no changes required in consumer repos, no explicit consumer list to maintain.

## How it works

```
Provider PR opened
       │
       ▼
┌─────────────────────────────────────┐
│  Download drift-guard-engine binary │
│  Auto-detect schema type & compare  │
│  (OpenAPI, GraphQL, or gRPC/proto)  │
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

## Prerequisites

- For private orgs, create a GitHub Personal Access Token (PAT) with `repo` and `read:org` scopes, then add it as a repository secret named `ORG_READ_TOKEN` (**Settings → Secrets and variables → Actions → New repository secret**).
- Optionally, add an `ANTHROPIC_API_KEY` secret to enable Claude-powered risk analysis in the issues the agent opens.

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
          # anthropic-api-key: ${{ secrets.ANTHROPIC_API_KEY }}  # optional: enables AI risk analysis
```

## Inputs

| Input | Required | Description |
|---|---|---|
| `base-schema` | No | Path to schema file (auto-detected if omitted). Supports OpenAPI (`.yaml`/`.yml`/`.json`), GraphQL (`.graphql`/`.gql`), and Protobuf (`.proto`). |
| `head-schema` | No | Path on PR branch (defaults to `base-schema`) |
| `generate-schema-cmd` | No | Shell command to generate the schema before diffing (e.g. `npm run build && node scripts/gen-swagger.js`). Useful for code-first frameworks that don't commit a schema file. |
| `org-read-token` | No | PAT with `repo:read` + `read:org` for private repos |
| `anthropic-api-key` | No | Enables Claude risk analysis in opened issues |

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `curl: (22) 404` when downloading drift-guard binary | Release assets are missing or named differently than expected | Check that the latest release on `pgomes13/drift-guard-engine` has GoReleaser artifacts attached — re-run the release if assets are missing |
| Action fails: "No API schema found" | Schema file not at a standard path, or generated at runtime and not committed | Set the `base-schema` input explicitly, or use `generate-schema-cmd` to generate it before diffing |
| Action fails: "drift-guard-engine failed to diff schemas" | Schema file is invalid or malformed | Validate locally: `drift-guard openapi --base ... --head ...` (or `graphql`/`grpc`) |
| No issues created, no errors | Missing `issues: write` permission | Add `issues: write` under `permissions:` in your workflow — the action will now emit a warning if this is missing |
| No consumers found (private org) | `GITHUB_TOKEN` can't search private repos | Set `org-read-token` to a PAT with `repo:read` + `read:org` |
| No consumers found (public org) | Breaking change path is too generic (e.g. `/v1`) | The agent searches for the first stable path segment — very short or version-only segments may not yield useful results |
| Issues created but no AI explanations | `ANTHROPIC_API_KEY` not set | Set the secret in your repo — the agent runs without it but skips the Claude risk analysis |

## Python CLI

Use this if you want to run the agent locally or integrate it into a non-GitHub CI system. You'll need a diff JSON file produced by `drift-guard-engine` first.

```sh
pip install drift-guard-agent

drift-guard-agent \
  --diff diff.json \
  --org my-org \
  --token $ORG_READ_TOKEN \
  --github-token $GITHUB_TOKEN \
  --pr 42
```

