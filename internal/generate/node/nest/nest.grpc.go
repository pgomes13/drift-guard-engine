package nest

import (
	"fmt"
	"path/filepath"

	"github.com/DriftAgent/api-drift-engine/internal/generate/node/express"
)

// NestGRPC finds the primary .proto file for the NestJS project and
// copies it to outputDir/schema.proto.
func NestGRPC(projectDir, outputDir string) error {
	src := express.FindProtoFile(projectDir)
	if src == "" {
		return fmt.Errorf(
			"no .proto file found in %s\n\n"+
				"Ensure your proto file is in one of:\n"+
				"  proto/, protos/, src/proto/, or the project root.",
			projectDir,
		)
	}
	return nestCopyFile(src, filepath.Join(outputDir, "schema.proto"))
}
