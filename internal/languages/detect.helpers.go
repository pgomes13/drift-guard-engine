package languages

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// isNestJSProject returns true when the project at dir has a package.json
// that declares @nestjs/swagger as a dependency.
func isNestJSProject(dir string) bool {
	return hasPackageJSONDep(dir, "@nestjs/swagger")
}

// isNodeJSProject returns true when the project at dir has a package.json but
// does NOT include @nestjs/swagger (plain Node.js / TypeScript, not NestJS).
func isNodeJSProject(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
		return false
	}
	return !hasPackageJSONDep(dir, "@nestjs/swagger")
}

// hasPackageJSONDep reports whether package.json in dir lists depName in
// dependencies or devDependencies.
func hasPackageJSONDep(dir, depName string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Dependencies    map[string]json.RawMessage `json:"dependencies"`
		DevDependencies map[string]json.RawMessage `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return strings.Contains(string(data), `"`+depName+`"`)
	}
	_, inDeps := pkg.Dependencies[depName]
	_, inDev := pkg.DevDependencies[depName]
	return inDeps || inDev
}
