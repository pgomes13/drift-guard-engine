package languages

import (
	"fmt"
	"os"
	"path/filepath"
)

// GeneratorFunc generates an OpenAPI schema for a project directory, writing
// the output into outputDir.
type GeneratorFunc func(projectDir, outputDir string) error

// ProjectInfo holds the human-readable project type name alongside the
// generator function resolved for that project.
type ProjectInfo struct {
	TypeName string
	Generate GeneratorFunc
}

// DetectProjectInfo is like DetectGenerator but also returns a display name
// for the detected project type.
func DetectProjectInfo(dir string) (ProjectInfo, error) {
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return ProjectInfo{"Go", GenerateGo}, nil
	}
	if isNestJSProject(dir) {
		return ProjectInfo{"NestJS", GenerateNest}, nil
	}
	if isExpressProject(dir) {
		return ProjectInfo{"Express", GenerateNode}, nil
	}
	if isNodeJSProject(dir) {
		return ProjectInfo{"Node.js", GenerateNode}, nil
	}
	if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
		return ProjectInfo{}, fmt.Errorf(
			"detected Python project\n\n" +
				"Auto-generation is not supported for Python. Use --cmd with your OpenAPI generation script:\n\n" +
				"  FastAPI example:\n" +
				`    drift-guard compare openapi --cmd "python scripts/generate_schema.py" --output openapi.json`,
		)
	}
	return ProjectInfo{}, fmt.Errorf(
		"could not detect project type in %s\n\n"+
			"Use --cmd to provide a generation command:\n"+
			`  drift-guard compare openapi --cmd "<your-generator>" --output <schema-file>`,
		dir,
	)
}

// GraphQLProjectInfo holds the human-readable project type name alongside the
// GraphQL generator function resolved for that project.
type GraphQLProjectInfo struct {
	TypeName    string
	GenerateGQL GeneratorFunc
}

// DetectGraphQLInfo returns GraphQL project info if the project uses GraphQL,
// or nil if no GraphQL API is detected.
func DetectGraphQLInfo(dir string) *GraphQLProjectInfo {
	if isNestJSProject(dir) && (nestHasGraphQLDep(dir) || hasGraphQLSchema(dir)) {
		return &GraphQLProjectInfo{"NestJS", GenerateNestGraphQL}
	}
	if (isExpressProject(dir) || isNodeJSProject(dir)) && (nodeHasGraphQLDeps(dir) || hasGraphQLSchema(dir)) {
		return &GraphQLProjectInfo{"Express", GenerateNodeGraphQL}
	}
	return nil
}

// DetectGraphQLGenerator returns the GraphQL generator for the project at dir.
func DetectGraphQLGenerator(dir string) (GeneratorFunc, error) {
	if isNestJSProject(dir) {
		return GenerateNestGraphQL, nil
	}
	if isExpressProject(dir) || isNodeJSProject(dir) {
		return GenerateNodeGraphQL, nil
	}
	return nil, fmt.Errorf("could not detect project type for GraphQL generation in %s", dir)
}

// DetectGenerator inspects dir and returns the appropriate schema generation
// function, or an error with actionable instructions when auto-generation is
// not supported.
func DetectGenerator(dir string) (GeneratorFunc, error) {
	// Go project
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return GenerateGo, nil
	}

	// NestJS project
	if isNestJSProject(dir) {
		return GenerateNest, nil
	}

	// Express project (package.json with express, not NestJS)
	if isExpressProject(dir) {
		return GenerateNode, nil
	}

	// Generic Node.js project
	if isNodeJSProject(dir) {
		return GenerateNode, nil
	}

	// Python project
	if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
		return nil, fmt.Errorf(
			"detected Python project\n\n" +
				"Auto-generation is not supported for Python. Use --cmd with your OpenAPI generation script:\n\n" +
				"  FastAPI example:\n" +
				`    drift-guard compare openapi --cmd "python scripts/generate_schema.py" --output openapi.json`,
		)
	}

	return nil, fmt.Errorf(
		"could not detect project type in %s\n\n"+
			"Use --cmd to provide a generation command:\n"+
			`  drift-guard compare openapi --cmd "<your-generator>" --output <schema-file>`,
		dir,
	)
}

