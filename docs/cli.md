# CLI

Run driftabot locally to check for API drift before adding it to CI.

## Install

See [Installation](/install) for all install options.

## Run on your repository

From the root of your project, run:

```sh
driftabot compare
```

<details>
<summary>Show steps</summary>

```
driftabot compare
       │
       ▼
┌────────────────────────────────────────────┐
│  1. Auto-detect                            │
│     Framework: Express · NestJS · Gin …    │
│     API types: OpenAPI · GraphQL · gRPC    │
└────────────────────────────────────────────┘
       │
       ▼
┌────────────────────────────────────────────┐
│  2. Select schema type                     │
│     GraphQL found?  → compare GraphQL      │
│     gRPC found?     → compare gRPC         │
│     Otherwise       → compare OpenAPI      │
└────────────────────────────────────────────┘
       │
       ▼
┌────────────────────────────────────────────┐
│  3. Scaffold (Node.js only, if needed)     │
│     NestJS  → @nestjs/swagger              │
│     Express → swagger-autogen or tsoa      │
│     Go      → swag init (auto, no prompt)  │
└────────────────────────────────────────────┘
       │
       ▼
┌────────────────────────────────────────────┐
│  4. Generate schemas                       │
│     head  ←  current branch               │
│     base  ←  origin/main (git worktree)   │
└────────────────────────────────────────────┘
       │
       ▼
┌────────────────────────────────────────────┐
│  5. Diff & print results                   │
│     breaking · non-breaking · info         │
└────────────────────────────────────────────┘
```

</details>

This is a good way to verify it works with your project before integrating further.

> If `driftabot compare` fails to auto-detect or generate schemas for your project, you can [generate them manually](/generating-specs) and pass the files directly with `driftabot openapi --base ... --head ...`.

### Check for breaking changes only

```sh
driftabot compare --fail-on-breaking
```

Exits with code `1` if any breaking changes are found — same behavior as in CI.

### Markdown output

```sh
driftabot compare --format markdown
```

Renders the same table that gets posted as a PR comment in CI.
