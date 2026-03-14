# Go SDK

driftabot can be used as a Go library. The `pkg/compare` and `pkg/impact` packages are publicly importable by external modules.

## Installation

```sh
go get github.com/DriftaBot/driftabot-engine@latest
```

## Comparing schemas

### OpenAPI

```go
import "github.com/DriftaBot/driftabot-engine/pkg/compare"

result, err := compare.OpenAPI("old.yaml", "new.yaml")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Breaking: %d, Non-breaking: %d\n",
    result.Summary.Breaking, result.Summary.NonBreaking)

for _, c := range result.Changes {
    fmt.Printf("[%s] %s\n", c.Severity, c.Description)
}
```

### GraphQL

```go
result, err := compare.GraphQL("old.graphql", "new.graphql")
```

### gRPC / Protobuf

```go
result, err := compare.GRPC("old.proto", "new.proto")
```

## Impact analysis

Scan source code for references to each breaking change in a diff result:

```go
import (
    "github.com/DriftaBot/driftabot-engine/pkg/compare"
    "github.com/DriftaBot/driftabot-engine/pkg/impact"
)

result, _ := compare.OpenAPI("old.yaml", "new.yaml")

hits, err := impact.Scan("./services", result.Changes)
if err != nil {
    log.Fatal(err)
}

for _, h := range hits {
    fmt.Printf("%s:%d  [%s] %s\n", h.File, h.LineNum, h.ChangeType, h.ChangePath)
}
```

Render a report:

```go
import (
    "os"
    "github.com/DriftaBot/driftabot-engine/pkg/impact"
)

impact.Report(os.Stdout, hits, "text")     // text table
impact.Report(os.Stdout, hits, "markdown") // markdown (collapsible sections)
impact.Report(os.Stdout, hits, "json")     // JSON array
```

## Types

```go
// pkg/schema
type Severity string // "breaking" | "non-breaking" | "info"

type Change struct {
    Type        string
    Severity    Severity
    Path        string
    Method      string
    Location    string
    Description string
    Before      string
    After       string
}

type Summary struct {
    Total      int
    Breaking   int
    NonBreaking int
    Info       int
}

type DiffResult struct {
    BaseFile string
    HeadFile string
    Changes  []Change
    Summary  Summary
}

// pkg/impact
type Hit struct {
    File       string
    LineNum    int
    Line       string
    ChangeType string
    ChangePath string
}
```
