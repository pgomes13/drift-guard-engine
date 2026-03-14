# Take a Tour

This page walks you through the full DriftGuard setup — from zero to your first automated drift alert in about 5 minutes.

---

## How it all fits together

**drift-guard-engine** is the core diff engine. It compares two schema files and classifies every change as breaking, non-breaking, or info.

**API Drift Agent** sits on top of the engine. When a provider PR changes an API, it uses the engine to find breaking changes, then searches your GitHub org for consumer repos that reference those endpoints, and opens GitHub Issues in each one — automatically.

```
Your provider repo
       │  PR opened
       ▼
drift-guard-engine  ←  detects breaking changes
       │
       ▼
API Drift Agent     ←  finds affected consumers in your org
       │
       ▼
GitHub Issues       ←  opened in each consumer repo
```

---

## Step 1 — Make sure your schema exists

The agent needs an OpenAPI schema file committed in your provider repo (e.g. `openapi.yaml` or `docs/openapi.json`).

**Don't have one?** See [Generating Specs](/generating-specs) for tools that export a schema from your framework (NestJS, Express, Gin, etc.).

If you already have a schema file checked in, move to Step 2.

---

## Step 2 — Create the workflow file

In your **provider** repo, create `.github/workflows/drift.yml`:

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
          # anthropic-api-key: ${{ secrets.ANTHROPIC_API_KEY }}  # optional: AI risk analysis
```

---

## Step 3 — Add the secrets

Go to your provider repo's **Settings → Secrets and variables → Actions** and add:

| Secret | Required | What it does |
|---|---|---|
| `ORG_READ_TOKEN` | For private orgs | PAT with `repo` + `read:org` scopes — lets the agent search your org for consumer repos |
| `ANTHROPIC_API_KEY` | Optional | Enables Claude-powered risk summaries in the GitHub Issues the agent opens |

For public orgs the default `GITHUB_TOKEN` is enough — you can omit `org-read-token` entirely.

---

## Step 4 — Open a pull request

Push a branch that changes your OpenAPI schema and open a PR. The agent will:

1. Download the drift-guard-engine binary
2. Diff the schema between `base` (main) and `head` (your branch)
3. If breaking changes are found — search your org for repos that reference the affected endpoints
4. Clone each consumer repo, scan for affected files, and open (or update) a GitHub Issue

If no issues are created and no errors appear, see [Troubleshooting](/api-drift-agent#troubleshooting).

---

## What's next

- [API Drift Agent](/api-drift-agent) — full reference for inputs, troubleshooting, and the Python CLI
- [MCP (AI)](/mcp) — use the engine as tools inside Claude Desktop
- [Supported frameworks](/supported) — what languages and frameworks the engine can parse schemas from
