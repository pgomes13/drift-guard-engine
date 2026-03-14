package helpers

import (
	"fmt"

	"github.com/DriftAgent/api-drift-engine/pkg/schema"
)

func IndexGQLArgs(args []schema.GQLArgument) map[string]schema.GQLArgument {
	m := make(map[string]schema.GQLArgument, len(args))
	for _, a := range args {
		m[a.Name] = a
	}
	return m
}

func DiffGQLArgs(typeName, fieldName string, base, head []schema.GQLArgument) []schema.Change {
	var changes []schema.Change
	loc := fmt.Sprintf("%s.%s", typeName, fieldName)

	baseArgs := IndexGQLArgs(base)
	headArgs := IndexGQLArgs(head)

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
