# CI Integration

Use drift-guard directly in CI to fail a PR when breaking API changes are detected.

## GitHub Actions

```yaml
name: API Drift Check

on:
  pull_request:

jobs:
  drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install drift-guard
        run: |
          curl -sSL https://github.com/pgomes13/api-drift-engine/releases/latest/download/drift-guard_linux_amd64.tar.gz | tar -xz
          sudo mv drift-guard /usr/local/bin/

      - name: Check for breaking changes
        run: drift-guard compare --fail-on-breaking
```

`--fail-on-breaking` exits with code `1` if any breaking changes are found, which fails the CI step.

## Explicit schema files

If auto-detection doesn't work for your project, generate schemas manually and pass them directly:

```yaml
- name: Check for breaking changes
  run: |
    drift-guard openapi \
      --base /tmp/base.yaml \
      --head /tmp/head.yaml \
      --fail-on-breaking
```

See [Generating Specs](/generating-specs) for how to produce the schema files.

## JSON output for downstream steps

```yaml
- name: Diff schemas
  id: diff
  run: |
    drift-guard compare --format json > drift-diff.json
    BREAKING=$(python3 -c "import json; d=json.load(open('drift-diff.json')); print(d.get('summary',{}).get('breaking',0))")
    echo "breaking=$BREAKING" >> $GITHUB_OUTPUT

- name: Annotate PR
  if: steps.diff.outputs.breaking != '0'
  run: echo "::warning::${{ steps.diff.outputs.breaking }} breaking API changes detected"
```

## Other CI systems

drift-guard is a single static binary — install it the same way on any CI runner.

```sh
# GitLab CI / CircleCI / Bitbucket Pipelines
curl -sSL https://github.com/pgomes13/api-drift-engine/releases/latest/download/drift-guard_linux_amd64.tar.gz | tar -xz
./drift-guard compare --fail-on-breaking
```

> For automated consumer notification and issue tracking, use the [API Drift Agent](https://github.com/marketplace/actions/api-drift-agent) — a GitHub Action that builds on the engine.
