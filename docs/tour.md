# Take a Tour

This page walks you through the full Drift Agent setup — from zero to your first automated drift alert.

---

## How it works

**api-drift-engine** is the core diff engine. It detects breaking API changes between two schema versions.

**API Drift Agent** sits on top of the engine. When a provider PR introduces breaking changes, the agent automatically finds every consumer repo in your org that references those endpoints and opens a GitHub Issue in each one.

```
Provider repo PR opened
       │
       ▼
api-drift-engine  ←  auto-detects & diffs API schema
       │  breaking changes found
       ▼
API Drift Agent     ←  searches org for affected consumers
       │
       ▼
GitHub Issues       ←  opened in each consumer repo
```

---

## Step 1 — Add the agent to your provider repo

Commit this workflow to your **provider** repo's **main branch** at `.github/workflows/drift.yml`:

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

      - uses: DriftAgent/api-drift-engine@v1
        with:
          org-read-token: ${{ secrets.ORG_READ_TOKEN }}
          # anthropic-api-key: ${{ secrets.ANTHROPIC_API_KEY }}  # optional: AI risk analysis
```

Also add the following secrets under **Settings → Secrets and variables → Actions**:

| Secret | When needed | What it does |
|---|---|---|
| `ORG_READ_TOKEN` | Private orgs | PAT with `repo` + `read:org` scopes — lets the agent search your org for consumer repos |
| `ANTHROPIC_API_KEY` | Optional | Enables Claude-powered risk summaries in opened issues |

> For public orgs the default `GITHUB_TOKEN` is sufficient — you can omit `org-read-token`.

---

## Step 2 — Make API changes on a branch

Create a branch and make a breaking change to your API — for example, remove an endpoint, rename a field, or change a parameter type. Commit and push the branch.

---

## Step 3 — Open a pull request

Open a PR from your branch into main. This triggers the agent workflow.

The agent will:
1. Auto-detect your API schema and diff it against main
2. Classify every change as breaking, non-breaking, or info
3. If breaking changes are found — search your org for repos that reference the affected endpoints
4. Open (or update) a GitHub Issue in each affected consumer repo

---

## Step 4 — Check the results

**If issues were created in consumer repos** — everything is working. The agent found consumers impacted by your breaking change.

**If no issues were created and no errors** — the agent ran but found no consumers referencing the changed endpoints. This is expected if no consumer repos use those paths yet.

**If the action failed or behaved unexpectedly** — see [Troubleshooting](/api-drift-engine#troubleshooting) for common causes and fixes.

---

## What's next

- [API Drift Agent](/api-drift-engine) — full reference for inputs and the Python CLI
- [MCP (AI)](/mcp) — use the engine as tools inside Claude Desktop
- [Supported frameworks](/supported) — what languages and frameworks the engine supports
