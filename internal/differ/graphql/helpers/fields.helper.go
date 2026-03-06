package helpers

import (
	"fmt"

	"drift-guard-diff-engine/pkg/schema"
)

func IndexGQLFields(fields []schema.GQLField) map[string]schema.GQLField {
	m := make(map[string]schema.GQLField, len(fields))
	for _, f := range fields {
		m[f.Name] = f
	}
	return m
}

func DiffGQLFields(typeName string, base, head []schema.GQLField) []schema.Change {
	var changes []schema.Change

	baseFields := IndexGQLFields(base)
	headFields := IndexGQLFields(head)

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

		changes = append(changes, DiffGQLArgs(typeName, name, bf.Arguments, hf.Arguments)...)
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

func DiffGQLInputFields(typeName string, base, head []schema.GQLField) []schema.Change {
	var changes []schema.Change

	baseFields := IndexGQLFields(base)
	headFields := IndexGQLFields(head)

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
