package compare

import (
	"fmt"

	"github.com/pgomes13/api-drift-engine/internal/classifier"
	differgrpc "github.com/pgomes13/api-drift-engine/internal/differ/grpc"
	parsergrpc "github.com/pgomes13/api-drift-engine/internal/parser/grpc"
	"github.com/pgomes13/api-drift-engine/pkg/schema"
)

// GRPC parses basePath and headPath as Protobuf schemas, diffs them,
// and returns the classified result.
func GRPC(basePath, headPath string) (schema.DiffResult, error) {
	base, err := parsergrpc.Parse(basePath)
	if err != nil {
		return schema.DiffResult{}, fmt.Errorf("parsing base: %w", err)
	}
	head, err := parsergrpc.Parse(headPath)
	if err != nil {
		return schema.DiffResult{}, fmt.Errorf("parsing head: %w", err)
	}
	return classifier.Classify(basePath, headPath, differgrpc.Diff(base, head)), nil
}
