package compare

import (
	"fmt"

	"github.com/DriftaBot/driftabot-engine/internal/classifier"
	differgraphql "github.com/DriftaBot/driftabot-engine/internal/differ/graphql"
	parsergraphql "github.com/DriftaBot/driftabot-engine/internal/parser/graphql"
	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

// GraphQL parses basePath and headPath as GraphQL SDL schemas, diffs them,
// and returns the classified result.
func GraphQL(basePath, headPath string) (schema.DiffResult, error) {
	base, err := parsergraphql.Parse(basePath)
	if err != nil {
		return schema.DiffResult{}, fmt.Errorf("parsing base: %w", err)
	}
	head, err := parsergraphql.Parse(headPath)
	if err != nil {
		return schema.DiffResult{}, fmt.Errorf("parsing head: %w", err)
	}
	return classifier.Classify(basePath, headPath, differgraphql.Diff(base, head)), nil
}
