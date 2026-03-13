# AI Agent

`drift-guard-agent` is a LangGraph-powered agent that runs in your **provider repo**. When a PR introduces breaking API changes, it automatically finds every consumer repo in your org that references those broken endpoints and opens a GitHub Issue in each one — no configuration or installation required on the consumer side.

## How it works

```
PR opened in provider repo
        │
        ▼
drift-guard detects breaking changes
        │
        ├── no breaking changes → done
        │
        ▼
Search org repos for code referencing the broken endpoint paths
(search terms derived from the diff — no service URL config needed)
        │
        ├── no matches → done
        │
        ▼
Shallow-clone each matched consumer repo using your token
        │
        ▼
Run drift-guard impact locally against each checkout
(no drift-guard needed in the consumer repo)
        │
        ├── no hits → skip
        │
        ▼
[optional] LLM explains impact per consumer (requires ANTHROPIC_API_KEY)
        │
        ▼
Open / update GitHub Issue in each affected consumer repo
```

## Installation

```sh
pip install drift-guard-agent
```

Requires Python ≥ 3.11.

## GitHub Actions setup (provider repo only)

Add this step after `pgomes13/drift-guard-engine@v3` in your workflow. **Nothing needs to be installed or configured in consumer repos.**

```yaml
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
        id: drift

      - name: Consumer impact scan
        if: steps.drift.outputs.drift-status == 'drift_detected'
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          ORG_READ_TOKEN: ${{ secrets.ORG_READ_TOKEN }}
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}  # optional
        run: |
          pip install drift-guard-agent
          drift-guard-agent \
            --diff ${{ steps.drift.outputs.diff-json }} \
            --org ${{ github.repository_owner }} \
            --token $ORG_READ_TOKEN \
            --github-token $GITHUB_TOKEN \
            --pr ${{ github.event.pull_request.number }}
```

**`ORG_READ_TOKEN`** must be a GitHub PAT (or fine-grained token) with `repo:read` and `read:org` scopes. This is the only secret needed — it lets the agent discover and clone consumer repos including private ones.

## What consumer repos receive

```markdown
## ⚠️ Breaking API changes from org/provider-service (PR #42)

Your repository references API endpoints that have been removed or changed.

### Breaking changes

- `GET /users/{id}` — endpoint removed
- `POST /users` — endpoint removed

### Affected files in this repo

| File | Line | Referenced path |
| ---- | ---- | --------------- |
| `src/api/users.ts` | 14 | `GET /users/{id}` |
| `src/hooks/useUsers.ts` | 31 | `POST /users` |

**Action required:** Update these references before the provider PR is merged.
```

Issues are opened with a `drift-guard` label. If an open issue already exists, it is updated rather than duplicating.

## CLI reference

```sh
# Read diff from file
drift-guard-agent --diff diff.json --org my-org

# Read diff from stdin (pipe from drift-guard)
drift-guard openapi --base old.yaml --head new.yaml --format json \
  | drift-guard-agent --org my-org

# Dry run: print issues without posting
drift-guard-agent --diff diff.json --org my-org --dry-run
```

| Flag | Env var | Description |
|---|---|---|
| `--diff` | — | Path to diff JSON, or `-` for stdin (default: `-`) |
| `--org` | `GITHUB_REPOSITORY_OWNER` | GitHub org to search for consumers |
| `--token` | `ORG_READ_TOKEN` | PAT with `repo:read` + `read:org` |
| `--github-token` | `GITHUB_TOKEN` | Token for posting Issues |
| `--pr` | `PR_NUMBER` | PR number to link in consumer Issues |
| `--provider-repo` | `GITHUB_REPOSITORY` | Provider repo full name (excluded from consumer search) |
| `--model` | `DRIFT_GUARD_MODEL` | Anthropic model for LLM explanations (default: `claude-opus-4-6`) |
| `--dry-run` | — | Print issues without posting to GitHub |

## LLM explanations (optional)

If `ANTHROPIC_API_KEY` is set, the agent uses `claude-opus-4-6` to add a plain-English explanation to each Issue entry:

> `src/api/users.ts:14` calls `GET /users/{id}` which has been removed. At runtime this will return 404, causing the user profile lookup to fail silently.

If the key is not set, the agent still runs and opens Issues — only the explanation paragraph is omitted.
