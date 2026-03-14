package helpers

import (
	"fmt"

	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

func IndexProperties(props []schema.Property) map[string]schema.Property {
	m := make(map[string]schema.Property, len(props))
	for _, p := range props {
		m[p.Name] = p
	}
	return m
}

func DiffProperties(path, method, location string, base, head []schema.Property) []schema.Change {
	var changes []schema.Change

	baseProps := IndexProperties(base)
	headProps := IndexProperties(head)

	for name, bp := range baseProps {
		hp, exists := headProps[name]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeFieldRemoved,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("%s.%s", location, name),
				Description: fmt.Sprintf("Field '%s' was removed from %s in %s %s", name, location, method, path),
			})
			continue
		}
		if bp.Type != hp.Type {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeFieldTypeChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("%s.%s.type", location, name),
				Description: fmt.Sprintf("Field '%s' type changed from '%s' to '%s' in %s of %s %s", name, bp.Type, hp.Type, location, method, path),
				Before:      bp.Type,
				After:       hp.Type,
			})
		}
		if bp.Required != hp.Required {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeFieldRequiredChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("%s.%s.required", location, name),
				Description: fmt.Sprintf("Field '%s' required changed from %v to %v in %s of %s %s", name, bp.Required, hp.Required, location, method, path),
				Before:      fmt.Sprintf("%v", bp.Required),
				After:       fmt.Sprintf("%v", hp.Required),
			})
		}
	}

	for name := range headProps {
		if _, exists := baseProps[name]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeFieldAdded,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("%s.%s", location, name),
				Description: fmt.Sprintf("Field '%s' was added to %s in %s %s", name, location, method, path),
			})
		}
	}

	return changes
}
