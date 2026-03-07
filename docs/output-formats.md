# Output Formats

## `text` (default)

```
Schema Diff: base.yaml → head.yaml
Total: 4  Breaking: 2  Non-Breaking: 1  Info: 1

SEVERITY        TYPE                    PATH            METHOD  LOCATION        DESCRIPTION
----------------------------------------------------------------------------------------------------
[BREAKING]      endpoint_removed        /users/{id}     DELETE                  Endpoint '/users/{id}' method DELETE was removed
[BREAKING]      param_type_changed      /users/{id}     GET     path.id         Param 'id' type changed from 'string' to 'integer'
[non-breaking]  endpoint_added          /posts                                  Endpoint '/posts' was added
[info]          field_added             /users          POST    request.role    Field 'role' was added
```

## `json`

```json
{
  "base_file": "base.yaml",
  "head_file": "head.yaml",
  "changes": [
    {
      "type": "endpoint_removed",
      "severity": "breaking",
      "path": "/users/{id}",
      "method": "DELETE",
      "location": "",
      "description": "Endpoint '/users/{id}' method DELETE was removed"
    }
  ],
  "summary": {
    "total": 4,
    "breaking": 2,
    "non_breaking": 1,
    "info": 1
  }
}
```

## `github`

Emits GitHub Actions [workflow commands](https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions) that render as inline annotations on the PR diff:

```
::error title=Breaking Change::Endpoint '/users/{id}' method DELETE was removed
::warning title=Non-Breaking Change::Endpoint '/posts' was added
::error title=API Contract Violation::2 breaking change(s) detected between base.yaml and head.yaml
```

Use `--format github` in CI to get inline PR annotations automatically.

## `markdown`

Renders a GitHub-flavored Markdown table — ideal for posting as a PR comment:

```
**Total: 4** | Breaking: 2 | Non-Breaking: 1 | Info: 1

| Severity | Type | Path | Method | Location | Description |
|----------|------|------|--------|----------|-------------|
| [BREAKING] | endpoint_removed | /users/{id} | DELETE |  | Endpoint '/users/{id}' method DELETE was removed |
| [BREAKING] | param_type_changed | /users/{id} | GET | path.id | Param 'id' type changed from 'string' to 'integer' |
| [non-breaking] | endpoint_added | /posts |  |  | Endpoint '/posts' was added |
| [info] | field_added | /users | POST | request.role | Field 'role' was added |
```

This is the format used by the [GitHub Action](./ci.md) when posting automatic PR comments.
