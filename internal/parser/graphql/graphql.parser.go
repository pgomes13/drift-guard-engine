package graphql

import (
	"fmt"
	"os"

	"github.com/DriftaBot/driftabot-engine/internal/parser/graphql/helpers"
	"github.com/DriftaBot/driftabot-engine/pkg/schema"

	"github.com/vektah/gqlparser/v2/ast"
	gqlparser "github.com/vektah/gqlparser/v2/parser"
)

// Parse reads a .graphql / .gql SDL file and returns a normalized GQLSchema.
func Parse(path string) (*schema.GQLSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	src := &ast.Source{Name: path, Input: string(data)}
	doc, parseErr := gqlparser.ParseSchema(src)
	if parseErr != nil {
		return nil, fmt.Errorf("parsing GraphQL SDL %s: %w", path, parseErr)
	}

	return helpers.Normalize(doc), nil
}
