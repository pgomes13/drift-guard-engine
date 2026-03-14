package compare

import (
	"fmt"

	"github.com/DriftaBot/driftabot-engine/internal/classifier"
	differopenapi "github.com/DriftaBot/driftabot-engine/internal/differ/openapi"
	parseropenapi "github.com/DriftaBot/driftabot-engine/internal/parser/openapi"
	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

// OpenAPI parses basePath and headPath as OpenAPI 3.x documents, diffs them,
// and returns the classified result.
func OpenAPI(basePath, headPath string) (schema.DiffResult, error) {
	base, err := parseropenapi.Parse(basePath)
	if err != nil {
		return schema.DiffResult{}, fmt.Errorf("parsing base: %w", err)
	}
	head, err := parseropenapi.Parse(headPath)
	if err != nil {
		return schema.DiffResult{}, fmt.Errorf("parsing head: %w", err)
	}
	return classifier.Classify(basePath, headPath, differopenapi.Diff(base, head)), nil
}
