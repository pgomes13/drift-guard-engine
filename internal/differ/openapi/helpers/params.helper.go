package helpers

import (
	"fmt"

	"drift-guard-diff-engine/pkg/schema"
)

func IndexParams(params []schema.Parameter) map[string]schema.Parameter {
	m := make(map[string]schema.Parameter, len(params))
	for _, p := range params {
		key := p.In + ":" + p.Name
		m[key] = p
	}
	return m
}

func DiffParams(path, method string, base, head []schema.Parameter) []schema.Change {
	var changes []schema.Change

	baseParams := IndexParams(base)
	headParams := IndexParams(head)

	for key, bp := range baseParams {
		hp, exists := headParams[key]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeParamRemoved,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("param.%s.%s", bp.In, bp.Name),
				Description: fmt.Sprintf("Parameter '%s' (in %s) was removed from %s %s", bp.Name, bp.In, method, path),
			})
			continue
		}
		if bp.Type != hp.Type {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeParamTypeChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("param.%s.%s.type", bp.In, bp.Name),
				Description: fmt.Sprintf("Parameter '%s' type changed from '%s' to '%s' in %s %s", bp.Name, bp.Type, hp.Type, method, path),
				Before:      bp.Type,
				After:       hp.Type,
			})
		}
		if bp.Required != hp.Required {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeParamRequiredChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("param.%s.%s.required", bp.In, bp.Name),
				Description: fmt.Sprintf("Parameter '%s' required changed from %v to %v in %s %s", bp.Name, bp.Required, hp.Required, method, path),
				Before:      fmt.Sprintf("%v", bp.Required),
				After:       fmt.Sprintf("%v", hp.Required),
			})
		}
	}

	for key, hp := range headParams {
		if _, exists := baseParams[key]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeParamAdded,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("param.%s.%s", hp.In, hp.Name),
				Description: fmt.Sprintf("Parameter '%s' (in %s) was added to %s %s", hp.Name, hp.In, method, path),
			})
		}
	}

	return changes
}
