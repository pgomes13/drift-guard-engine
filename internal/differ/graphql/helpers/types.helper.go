// Package helpers provides internal diffing utilities for the GraphQL differ.
package helpers

import (
	"fmt"

	"drift-guard-engine/pkg/schema"
)

func IndexGQLTypes(s *schema.GQLSchema) map[string]schema.GQLType {
	m := make(map[string]schema.GQLType, len(s.Types))
	for _, t := range s.Types {
		m[t.Name] = t
	}
	return m
}

func DiffGQLType(base, head schema.GQLType) []schema.Change {
	var changes []schema.Change

	// Type kind changed (e.g. Object → Interface) — always breaking
	if base.Kind != head.Kind {
		changes = append(changes, schema.Change{
			Type:        schema.ChangeTypeGQLTypeKindChanged,
			Location:    base.Name,
			Description: fmt.Sprintf("Type '%s' kind changed from %s to %s", base.Name, base.Kind, head.Kind),
			Before:      string(base.Kind),
			After:       string(head.Kind),
		})
		return changes // further field diff is meaningless after a kind change
	}

	switch base.Kind {
	case schema.GQLTypeKindObject, schema.GQLTypeKindInterface:
		changes = append(changes, DiffGQLFields(base.Name, base.Fields, head.Fields)...)
		changes = append(changes, DiffGQLInterfaces(base.Name, base.Interfaces, head.Interfaces)...)

	case schema.GQLTypeKindInput:
		changes = append(changes, DiffGQLInputFields(base.Name, base.Fields, head.Fields)...)

	case schema.GQLTypeKindEnum:
		changes = append(changes, DiffGQLEnumValues(base.Name, base.Values, head.Values)...)

	case schema.GQLTypeKindUnion:
		changes = append(changes, DiffGQLUnionMembers(base.Name, base.Members, head.Members)...)
	}

	return changes
}
