// Package graphql computes the structural diff between two GraphQL schemas,
// producing a flat list of Change values for downstream classification.
package graphql

import (
	"fmt"
	"strings"

	"drift-guard-diff-engine/pkg/schema"
)

// Diff computes all changes between two normalized GraphQL schemas.
func Diff(base, head *schema.GQLSchema) []schema.Change {
	var changes []schema.Change

	baseTypes := indexGQLTypes(base)
	headTypes := indexGQLTypes(head)

	// Removed types
	for name, bt := range baseTypes {
		ht, exists := headTypes[name]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLTypeRemoved,
				Location:    name,
				Description: fmt.Sprintf("Type '%s' (%s) was removed", name, bt.Kind),
				Before:      string(bt.Kind),
			})
			continue
		}
		changes = append(changes, diffGQLType(bt, ht)...)
	}

	// Added types
	for name, ht := range headTypes {
		if _, exists := baseTypes[name]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLTypeAdded,
				Location:    name,
				Description: fmt.Sprintf("Type '%s' (%s) was added", name, ht.Kind),
				After:       string(ht.Kind),
			})
		}
	}

	return changes
}

func indexGQLTypes(s *schema.GQLSchema) map[string]schema.GQLType {
	m := make(map[string]schema.GQLType, len(s.Types))
	for _, t := range s.Types {
		m[t.Name] = t
	}
	return m
}

func diffGQLType(base, head schema.GQLType) []schema.Change {
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
		changes = append(changes, diffGQLFields(base.Name, base.Fields, head.Fields)...)
		changes = append(changes, diffGQLInterfaces(base.Name, base.Interfaces, head.Interfaces)...)

	case schema.GQLTypeKindInput:
		changes = append(changes, diffGQLInputFields(base.Name, base.Fields, head.Fields)...)

	case schema.GQLTypeKindEnum:
		changes = append(changes, diffGQLEnumValues(base.Name, base.Values, head.Values)...)

	case schema.GQLTypeKindUnion:
		changes = append(changes, diffGQLUnionMembers(base.Name, base.Members, head.Members)...)
	}

	return changes
}

// --------------------------------------------------------------------------
// Object / Interface fields
// --------------------------------------------------------------------------

func indexGQLFields(fields []schema.GQLField) map[string]schema.GQLField {
	m := make(map[string]schema.GQLField, len(fields))
	for _, f := range fields {
		m[f.Name] = f
	}
	return m
}

func diffGQLFields(typeName string, base, head []schema.GQLField) []schema.Change {
	var changes []schema.Change

	baseFields := indexGQLFields(base)
	headFields := indexGQLFields(head)

	for name, bf := range baseFields {
		hf, exists := headFields[name]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLFieldRemoved,
				Location:    fmt.Sprintf("%s.%s", typeName, name),
				Description: fmt.Sprintf("Field '%s.%s' was removed", typeName, name),
				Before:      bf.Type,
			})
			continue
		}

		if bf.Type != hf.Type {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLFieldTypeChanged,
				Location:    fmt.Sprintf("%s.%s", typeName, name),
				Description: fmt.Sprintf("Field '%s.%s' type changed from '%s' to '%s'", typeName, name, bf.Type, hf.Type),
				Before:      bf.Type,
				After:       hf.Type,
			})
		}

		if !bf.Deprecated && hf.Deprecated {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLFieldDeprecated,
				Location:    fmt.Sprintf("%s.%s", typeName, name),
				Description: fmt.Sprintf("Field '%s.%s' was deprecated", typeName, name),
			})
		}

		changes = append(changes, diffGQLArgs(typeName, name, bf.Arguments, hf.Arguments)...)
	}

	for name, hf := range headFields {
		if _, exists := baseFields[name]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLFieldAdded,
				Location:    fmt.Sprintf("%s.%s", typeName, name),
				Description: fmt.Sprintf("Field '%s.%s' was added with type '%s'", typeName, name, hf.Type),
				After:       hf.Type,
			})
		}
	}

	return changes
}

// --------------------------------------------------------------------------
// Field arguments
// --------------------------------------------------------------------------

func indexGQLArgs(args []schema.GQLArgument) map[string]schema.GQLArgument {
	m := make(map[string]schema.GQLArgument, len(args))
	for _, a := range args {
		m[a.Name] = a
	}
	return m
}

func diffGQLArgs(typeName, fieldName string, base, head []schema.GQLArgument) []schema.Change {
	var changes []schema.Change
	loc := fmt.Sprintf("%s.%s", typeName, fieldName)

	baseArgs := indexGQLArgs(base)
	headArgs := indexGQLArgs(head)

	for name, ba := range baseArgs {
		ha, exists := headArgs[name]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLArgRemoved,
				Location:    fmt.Sprintf("%s(arg:%s)", loc, name),
				Description: fmt.Sprintf("Argument '%s' was removed from field '%s'", name, loc),
				Before:      ba.Type,
			})
			continue
		}

		if ba.Type != ha.Type {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLArgTypeChanged,
				Location:    fmt.Sprintf("%s(arg:%s)", loc, name),
				Description: fmt.Sprintf("Argument '%s' on '%s' type changed from '%s' to '%s'", name, loc, ba.Type, ha.Type),
				Before:      ba.Type,
				After:       ha.Type,
			})
		}

		if ba.DefaultValue != ha.DefaultValue {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLArgDefaultChanged,
				Location:    fmt.Sprintf("%s(arg:%s)", loc, name),
				Description: fmt.Sprintf("Argument '%s' on '%s' default changed from '%s' to '%s'", name, loc, ba.DefaultValue, ha.DefaultValue),
				Before:      ba.DefaultValue,
				After:       ha.DefaultValue,
			})
		}
	}

	for name, ha := range headArgs {
		if _, exists := baseArgs[name]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLArgAdded,
				Location:    fmt.Sprintf("%s(arg:%s)", loc, name),
				Description: fmt.Sprintf("Argument '%s' was added to field '%s' with type '%s'", name, loc, ha.Type),
				After:       ha.Type,
			})
		}
	}

	return changes
}

// --------------------------------------------------------------------------
// Input fields
// --------------------------------------------------------------------------

func diffGQLInputFields(typeName string, base, head []schema.GQLField) []schema.Change {
	var changes []schema.Change

	baseFields := indexGQLFields(base)
	headFields := indexGQLFields(head)

	for name, bf := range baseFields {
		hf, exists := headFields[name]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLInputFieldRemoved,
				Location:    fmt.Sprintf("%s.%s", typeName, name),
				Description: fmt.Sprintf("Input field '%s.%s' was removed", typeName, name),
				Before:      bf.Type,
			})
			continue
		}
		if bf.Type != hf.Type {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLInputFieldTypeChanged,
				Location:    fmt.Sprintf("%s.%s", typeName, name),
				Description: fmt.Sprintf("Input field '%s.%s' type changed from '%s' to '%s'", typeName, name, bf.Type, hf.Type),
				Before:      bf.Type,
				After:       hf.Type,
			})
		}
	}

	for name, hf := range headFields {
		if _, exists := baseFields[name]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGQLInputFieldAdded,
				Location:    fmt.Sprintf("%s.%s", typeName, name),
				Description: fmt.Sprintf("Input field '%s.%s' was added with type '%s'", typeName, name, hf.Type),
				After:       hf.Type,
			})
		}
	}

	return changes
}

// --------------------------------------------------------------------------
// Enum values / Union members / Interfaces  (shared set-diff logic)
// --------------------------------------------------------------------------

// diffStringSet computes added/removed entries between two string slices and
// emits Change records using the supplied change types and format functions.
func diffStringSet(
	typeName string,
	base, head []string,
	removedType, addedType schema.ChangeType,
	locFn func(typeName, val string) string,
	removedDescFn func(typeName, val string) string,
	addedDescFn func(typeName, val string) string,
) []schema.Change {
	var changes []schema.Change
	baseSet := toSet(base)
	headSet := toSet(head)

	for v := range baseSet {
		if !headSet[v] {
			changes = append(changes, schema.Change{
				Type:        removedType,
				Location:    locFn(typeName, v),
				Description: removedDescFn(typeName, v),
				Before:      v,
			})
		}
	}
	for v := range headSet {
		if !baseSet[v] {
			changes = append(changes, schema.Change{
				Type:        addedType,
				Location:    locFn(typeName, v),
				Description: addedDescFn(typeName, v),
				After:       v,
			})
		}
	}
	return changes
}

func diffGQLEnumValues(typeName string, base, head []string) []schema.Change {
	return diffStringSet(typeName, base, head,
		schema.ChangeTypeGQLEnumValueRemoved,
		schema.ChangeTypeGQLEnumValueAdded,
		func(t, v string) string { return fmt.Sprintf("%s.%s", t, v) },
		func(t, v string) string { return fmt.Sprintf("Enum value '%s' was removed from '%s'", v, t) },
		func(t, v string) string { return fmt.Sprintf("Enum value '%s' was added to '%s'", v, t) },
	)
}

func diffGQLUnionMembers(typeName string, base, head []string) []schema.Change {
	return diffStringSet(typeName, base, head,
		schema.ChangeTypeGQLUnionMemberRemoved,
		schema.ChangeTypeGQLUnionMemberAdded,
		func(t, v string) string { return fmt.Sprintf("%s | %s", t, v) },
		func(t, v string) string { return fmt.Sprintf("Union member '%s' was removed from '%s'", v, t) },
		func(t, v string) string { return fmt.Sprintf("Union member '%s' was added to '%s'", v, t) },
	)
}

func diffGQLInterfaces(typeName string, base, head []string) []schema.Change {
	return diffStringSet(typeName, base, head,
		schema.ChangeTypeGQLInterfaceRemoved,
		schema.ChangeTypeGQLInterfaceAdded,
		func(t, v string) string { return fmt.Sprintf("%s implements %s", t, v) },
		func(t, v string) string { return fmt.Sprintf("Type '%s' no longer implements interface '%s'", t, v) },
		func(t, v string) string { return fmt.Sprintf("Type '%s' now implements interface '%s'", t, v) },
	)
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func toSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[strings.TrimSpace(s)] = true
	}
	return m
}
