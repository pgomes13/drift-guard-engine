# Playground

Try drift-agent interactively — no install required.

**[Open Playground →](https://drift-agent-theta.vercel.app/)**

## Schema diff

Paste or edit two schema versions side by side and click **Compare** to see a categorised list of breaking, non-breaking, and info-level changes.

Supports all three schema types:

| Tab | Format |
|-----|--------|
| OpenAPI | YAML (OpenAPI 3.x) |
| GraphQL | SDL |
| gRPC / Protobuf | `.proto` (proto3) |

Each tab loads a built-in sample diff so you can see results immediately without writing any schema.

## Impact analysis

When the diff contains breaking changes, an **Impact Analysis** tab appears next to the diff results.

1. Click **Impact Analysis**
2. Paste a code file from your service into the editor (set the filename so the scanner recognises the language)
3. Click **Scan for References**

The playground scans the pasted code and shows every line that references each breaking change, grouped by change:

```
🔴 DELETE /users/{id} (endpoint_removed) — 2 hit(s)
  service.go : 12   client.Delete("/users/" + id)
  service.go : 34   r.DELETE("/users/:id", handler)
```

This is the same scanner used by `drift-agent impact` — the playground lets you try it without a local install.
