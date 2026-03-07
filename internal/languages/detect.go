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

// DetectGenerator inspects dir and returns the appropriate schema generation
// function, or an error with actionable instructions when auto-generation is
// not supported.
func DetectGenerator(dir string) (GeneratorFunc, error) {
	// Go project
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return GenerateGo, nil
	}

	// NestJS project (Node.js + @nestjs/swagger)
	if isNestJSProject(dir) {
		return GenerateNest, nil
	}

	// Plain Node.js project (package.json without NestJS) — tsoa works here too.
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

