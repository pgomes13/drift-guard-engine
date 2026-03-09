# CLI

Run drift-guard locally to check for API drift before adding it to CI.

## Install

See [Installation](/install) for all install options.

## Run on your repository

From the root of your project, run:

```sh
drift-guard compare
```

<details>
<summary>Show steps</summary>

1. Auto-detect your framework and API types (OpenAPI, GraphQL, gRPC)
2. Prompt you for which API type to compare
3. For Express/NestJS projects with no existing swagger script, offer to scaffold `swagger-autogen` or `tsoa`. For Go projects, run `swag init` automatically — no prompt, no scaffolding. This requires [`swag`](https://github.com/swaggo/swag) to be installed and your handlers to have `swag` annotations; if either is missing the command will error.
4. Generate schemas for your current branch (head) and `origin/main` (base) using a git worktree
5. Print the diff

</details>

This is a good way to verify it works with your project before wiring up the GitHub Action.

> If `drift-guard compare` fails to auto-detect or generate schemas for your project, you can [generate them manually](/generating-specs) and pass the files directly with `drift-guard openapi --base ... --head ...`.

### Check for breaking changes only

```sh
drift-guard compare --fail-on-breaking
```

Exits with code `1` if any breaking changes are found — same behavior as in CI.

### Markdown output

```sh
drift-guard compare --format markdown
```

Renders the same table that gets posted as a PR comment in CI.
