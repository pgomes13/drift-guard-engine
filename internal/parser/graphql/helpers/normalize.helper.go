// Package helpers provides internal normalization utilities for the GraphQL parser.
package helpers

import (
	"github.com/DriftaBot/driftabot-engine/pkg/schema"

	"github.com/vektah/gqlparser/v2/ast"
)

// BuiltinTypes are intrinsic GraphQL types that should be excluded from diffs.
var BuiltinTypes = map[string]bool{
	"String": true, "Boolean": true, "Int": true, "Float": true, "ID": true,
	"__Schema": true, "__Type": true, "__TypeKind": true, "__Field": true,
	"__InputValue": true, "__EnumValue": true, "__Directive": true,
	"__DirectiveLocation": true,
}

// Normalize converts a parsed SDL document into a normalized GQLSchema.
func Normalize(doc *ast.SchemaDocument) *schema.GQLSchema {
	s := &schema.GQLSchema{}

	for _, def := range doc.Definitions {
		if BuiltinTypes[def.Name] {
			continue
		}

		t := schema.GQLType{
			Name:        def.Name,
			Description: def.Description,
		}

		switch def.Kind {
		case ast.Object:
			t.Kind = schema.GQLTypeKindObject
			t.Fields = NormalizeFields(def.Fields)
			for _, iface := range def.Interfaces {
				t.Interfaces = append(t.Interfaces, iface)
			}

		case ast.Interface:
			t.Kind = schema.GQLTypeKindInterface
			t.Fields = NormalizeFields(def.Fields)

		case ast.Union:
			t.Kind = schema.GQLTypeKindUnion
			for _, m := range def.Types {
				t.Members = append(t.Members, m)
			}

		case ast.Enum:
			t.Kind = schema.GQLTypeKindEnum
			for _, v := range def.EnumValues {
				t.Values = append(t.Values, v.Name)
			}

		case ast.InputObject:
			t.Kind = schema.GQLTypeKindInput
			t.Fields = NormalizeFields(def.Fields)

		case ast.Scalar:
			t.Kind = schema.GQLTypeKindScalar

		default:
			continue
		}

		s.Types = append(s.Types, t)
	}

	return s
}
