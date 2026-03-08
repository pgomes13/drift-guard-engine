package languages

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// goProjectTypeName returns the full display name for a Go project at dir,
// e.g. "Go (Gin)", "Go (Echo)", or plain "Go" when no framework is detected.
func goProjectTypeName(dir string) string {
	if fw := goFrameworkName(dir); fw != "" {
		return "Go (" + fw + ")"
	}
	return "Go"
}

// isGoProject returns true when dir is a Go project. It accepts both modern
// projects (go.mod present) and legacy projects (no go.mod but .go sources or
// a dep/glide manifest are present).
func isGoProject(dir string) bool {
	// Modern: go.mod present
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return true
	}
	// Legacy dep / glide manifests
	for _, name := range []string{"Gopkg.toml", "Gopkg.lock", "glide.yaml", "glide.lock"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	// Any .go files in the root directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
			return true
		}
	}
	return false
}

// goFrameworkName returns the display name of the Go web framework used in
// the project at dir, or an empty string if no recognised framework is found.
// It searches go.mod, go.sum, Gopkg.toml/lock, and the vendor/ directory so
// that both modern (go modules) and legacy (dep/glide) projects are covered.
func goFrameworkName(dir string) string {
	// Ordered list of (import path substring → display name).
	frameworks := []struct{ path, name string }{
		{"github.com/gin-gonic/gin", "Gin"},
		{"github.com/labstack/echo", "Echo"},
		{"github.com/gofiber/fiber", "Fiber"},
		{"github.com/go-chi/chi", "Chi"},
		{"github.com/gorilla/mux", "Gorilla Mux"},
	}

	// Check text-based manifest files.
	for _, fname := range []string{"go.mod", "go.sum", "Gopkg.toml", "Gopkg.lock"} {
		data, err := os.ReadFile(filepath.Join(dir, fname))
		if err != nil {
			continue
		}
		content := string(data)
		for _, fw := range frameworks {
			if strings.Contains(content, fw.path) {
				return fw.name
			}
		}
	}

	// Fallback: vendor directory layout.
	for _, fw := range frameworks {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash("vendor/"+fw.path))); err == nil {
			return fw.name
		}
	}

	return ""
}

// isNestJSProject returns true when the project at dir has a package.json
// that declares any core NestJS package as a dependency.
func isNestJSProject(dir string) bool {
	for _, pkg := range []string{"@nestjs/core", "@nestjs/common", "@nestjs/swagger"} {
		if hasPackageJSONDep(dir, pkg) {
			return true
		}
	}
	return false
}

// isExpressProject returns true when the project at dir has a package.json
// that declares express as a dependency but is NOT a NestJS project.
func isExpressProject(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
		return false
	}
	return hasPackageJSONDep(dir, "express") && !isNestJSProject(dir)
}

// isNodeJSProject returns true when the project at dir has a package.json but
// is not NestJS or Express (generic Node.js / TypeScript project).
func isNodeJSProject(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
		return false
	}
	return !isNestJSProject(dir)
}

// nestHasGraphQLDep reports whether @nestjs/graphql is listed in package.json.
func nestHasGraphQLDep(dir string) bool {
	return hasPackageJSONDep(dir, "@nestjs/graphql")
}

// nodeHasGraphQLDeps reports whether any GraphQL-related package is listed
// in the project's package.json.
func nodeHasGraphQLDeps(dir string) bool {
	for _, dep := range []string{
		"apollo-server-express", "apollo-server",
		"type-graphql", "@apollo/server", "graphql-yoga",
	} {
		if hasPackageJSONDep(dir, dep) {
			return true
		}
	}
	return false
}

// hasProtoFilesInDir reports whether any .proto files exist under dir.
func hasProtoFilesInDir(dir string) bool {
	for _, sub := range []string{"proto", "protos", "src/proto", "src/protos", "grpc", "."} {
		entries, err := os.ReadDir(filepath.Join(dir, filepath.FromSlash(sub)))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && len(e.Name()) > 6 && e.Name()[len(e.Name())-6:] == ".proto" {
				return true
			}
		}
	}
	return false
}

// hasGraphQLSchema reports whether a GraphQL schema file exists in dir.
func hasGraphQLSchema(dir string) bool {
	for _, rel := range []string{
		"schema.graphql", "schema.gql",
		"src/schema.graphql", "src/schema.gql",
		"graphql/schema.graphql", "graphql/schema.gql",
		"api/schema.graphql",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			return true
		}
	}
	return false
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
