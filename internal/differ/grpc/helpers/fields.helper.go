package helpers

import (
	"fmt"

	"github.com/DriftAgent/api-drift-engine/pkg/schema"
)

func indexFields(fields []schema.GRPCField) map[string]schema.GRPCField {
	m := make(map[string]schema.GRPCField, len(fields))
	for _, f := range fields {
		m[f.Name] = f
	}
	return m
}

// DiffFields compares two slices of message fields and returns all changes.
func DiffFields(messageName string, base, head []schema.GRPCField) []schema.Change {
	var changes []schema.Change

	baseFields := indexFields(base)
	headFields := indexFields(head)

	for name, bf := range baseFields {
		hf, exists := headFields[name]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCFieldRemoved,
				Location:    fmt.Sprintf("%s.%s", messageName, name),
				Description: fmt.Sprintf("Field '%s.%s' was removed", messageName, name),
				Before:      bf.Type,
			})
			continue
		}

		if bf.Type != hf.Type {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCFieldTypeChanged,
				Location:    fmt.Sprintf("%s.%s", messageName, name),
				Description: fmt.Sprintf("Field '%s.%s' type changed from '%s' to '%s'", messageName, name, bf.Type, hf.Type),
				Before:      bf.Type,
				After:       hf.Type,
			})
		}

		if bf.Number != hf.Number {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCFieldNumberChanged,
				Location:    fmt.Sprintf("%s.%s", messageName, name),
				Description: fmt.Sprintf("Field '%s.%s' number changed from %d to %d", messageName, name, bf.Number, hf.Number),
				Before:      fmt.Sprintf("%d", bf.Number),
				After:       fmt.Sprintf("%d", hf.Number),
			})
		}

		if bf.Repeated != hf.Repeated {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCFieldLabelChanged,
				Location:    fmt.Sprintf("%s.%s", messageName, name),
				Description: fmt.Sprintf("Field '%s.%s' repeated label changed", messageName, name),
				Before:      labelStr(bf.Repeated),
				After:       labelStr(hf.Repeated),
			})
		}
	}

	for name, hf := range headFields {
		if _, exists := baseFields[name]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCFieldAdded,
				Location:    fmt.Sprintf("%s.%s", messageName, name),
				Description: fmt.Sprintf("Field '%s.%s' was added with type '%s'", messageName, name, hf.Type),
				After:       hf.Type,
			})
		}
	}

	return changes
}

func labelStr(repeated bool) string {
	if repeated {
		return "repeated"
	}
	return "singular"
}
