package golang

import (
	"fmt"
	"os"
	"path/filepath"
)

// GoGraphQL finds the GraphQL SDL schema for the Go project and copies it to
// outputDir/schema.graphql.
func GoGraphQL(projectDir, outputDir string) error {
	src := FindGraphQLSchema(projectDir)
	if src == "" {
		return fmt.Errorf(
			"no GraphQL schema found in %s\n\n"+
				"Ensure your schema is committed at one of:\n"+
				"  schema.graphql, schema.gql, src/schema.graphql, graphql/schema.graphql\n\n"+
				"Or use --cmd to provide a generation command:\n"+
				`  drift-agent compare graphql --cmd "go run ./tools/gqlgen" --output schema.graphql`,
			projectDir,
		)
	}
	return copySchema(src, filepath.Join(outputDir, "schema.graphql"))
}

// FindGraphQLSchema returns the absolute path of the first GraphQL schema file
// found in dir, or an empty string if none is found.
func FindGraphQLSchema(dir string) string {
	for _, rel := range []string{
		"schema.graphql",
		"schema.gql",
		"src/schema.graphql",
		"src/schema.gql",
		"graphql/schema.graphql",
		"graphql/schema.gql",
		"api/schema.graphql",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			return filepath.Join(dir, rel)
		}
	}
	return ""
}

func copySchema(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read schema %s: %w", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
