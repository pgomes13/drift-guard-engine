package express

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NodeGraphQL finds the GraphQL SDL schema for the Express/Node project and
// copies it to outputDir/schema.graphql.
func NodeGraphQL(projectDir, outputDir string) error {
	src := FindGraphQLSchema(projectDir)
	if src == "" {
		return fmt.Errorf(
			"no GraphQL schema found in %s\n\n"+
				"Ensure your schema is committed at one of:\n"+
				"  schema.graphql, schema.gql, src/schema.graphql, graphql/schema.graphql\n\n"+
				"Or generate it first and commit the result.",
			projectDir,
		)
	}
	return copyFile(src, filepath.Join(outputDir, "schema.graphql"))
}

// HasGraphQLDep reports whether the project at dir has a GraphQL-related dependency.
func HasGraphQLDep(dir string) bool {
	for _, dep := range []string{
		"graphql", "apollo-server-express", "apollo-server",
		"type-graphql", "@apollo/server", "graphql-yoga",
	} {
		if nodeDepExists(dir, dep) {
			return true
		}
	}
	return false
}

// FindGraphQLSchema returns the absolute path of the first GraphQL schema file
// found in dir, or empty string if none is found.
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

func nodeDepExists(dir, depName string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), `"`+depName+`"`)
}
