package helpers

import (
	"strings"

	"drift-guard-diff-engine/pkg/schema"

	"github.com/vektah/gqlparser/v2/ast"
)

// NormalizeFields converts an SDL FieldList into normalized GQLField values.
func NormalizeFields(fields ast.FieldList) []schema.GQLField {
	result := make([]schema.GQLField, 0, len(fields))
	for _, f := range fields {
		gf := schema.GQLField{
			Name:        f.Name,
			Type:        f.Type.String(),
			Description: f.Description,
		}
		if f.Directives.ForName("deprecated") != nil {
			gf.Deprecated = true
		}
		gf.Arguments = NormalizeArgs(f.Arguments)
		result = append(result, gf)
	}
	return result
}

// NormalizeArgs converts an SDL ArgumentDefinitionList into normalized GQLArgument values.
func NormalizeArgs(args ast.ArgumentDefinitionList) []schema.GQLArgument {
	result := make([]schema.GQLArgument, 0, len(args))
	for _, a := range args {
		ga := schema.GQLArgument{
			Name:        a.Name,
			Type:        a.Type.String(),
			Description: a.Description,
		}
		if a.DefaultValue != nil {
			ga.DefaultValue = strings.TrimSpace(a.DefaultValue.String())
		}
		result = append(result, ga)
	}
	return result
}
