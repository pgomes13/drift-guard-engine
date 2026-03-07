package languages

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/pgomes13/drift-guard-engine/internal/generate"
)

// GenerateNode delegates to the generate package.
var GenerateNode = generate.Node

// isNodeProject returns true when the project has a package.json but does NOT
// include @nestjs/swagger (i.e. it is a plain JS/TS project, not NestJS).
func isNodeProject(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Dependencies    map[string]json.RawMessage `json:"dependencies"`
		DevDependencies map[string]json.RawMessage `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return !strings.Contains(string(data), `"@nestjs/swagger"`)
	}
	_, inDeps := pkg.Dependencies["@nestjs/swagger"]
	_, inDev := pkg.DevDependencies["@nestjs/swagger"]
	return !inDeps && !inDev
}
