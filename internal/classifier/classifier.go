package classifier

import (
	"github.com/pgomes13/api-drift-engine/internal/classifier/graphql"
	"github.com/pgomes13/api-drift-engine/internal/classifier/grpc"
	"github.com/pgomes13/api-drift-engine/internal/classifier/openapi"
	"github.com/pgomes13/api-drift-engine/pkg/schema"
)

// Classify assigns a Severity to each Change and builds a DiffResult.
func Classify(baseFile, headFile string, changes []schema.Change) schema.DiffResult {
	classified := make([]schema.Change, 0, len(changes))
	result := schema.DiffResult{
		BaseFile: baseFile,
		HeadFile: headFile,
	}

	for _, c := range changes {
		c.Severity = severityFor(c)
		classified = append(classified, c)

		switch c.Severity {
		case schema.SeverityBreaking:
			result.Summary.Breaking++
		case schema.SeverityNonBreaking:
			result.Summary.NonBreaking++
		case schema.SeverityInfo:
			result.Summary.Info++
		}
	}

	result.Changes = classified
	result.Summary.Total = len(classified)
	return result
}

// severityFor dispatches to the schema-type-specific severity function.
func severityFor(c schema.Change) schema.Severity {
	if s, ok := openapi.Severity(c); ok {
		return s
	}
	if s, ok := graphql.Severity(c); ok {
		return s
	}
	if s, ok := grpc.Severity(c); ok {
		return s
	}
	return schema.SeverityInfo
}
