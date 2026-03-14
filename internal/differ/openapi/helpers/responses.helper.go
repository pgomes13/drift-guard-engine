package helpers

import (
	"fmt"

	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

func DiffResponses(path, method string, base, head []schema.Response) []schema.Change {
	var changes []schema.Change

	baseMap := make(map[string]schema.Response, len(base))
	for _, r := range base {
		baseMap[r.StatusCode] = r
	}
	headMap := make(map[string]schema.Response, len(head))
	for _, r := range head {
		headMap[r.StatusCode] = r
	}

	for code, br := range baseMap {
		hr, exists := headMap[code]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeResponseChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("response.%s", code),
				Description: fmt.Sprintf("Response %s was removed from %s %s", code, method, path),
			})
			continue
		}
		changes = append(changes, DiffProperties(path, method, fmt.Sprintf("response.%s", code), br.Properties, hr.Properties)...)
	}

	for code := range headMap {
		if _, exists := baseMap[code]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeResponseChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("response.%s", code),
				Description: fmt.Sprintf("Response %s was added to %s %s", code, method, path),
			})
		}
	}

	return changes
}
