package helpers

import (
	"fmt"

	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

func DiffRequestBody(path, method string, base, head *schema.RequestBody) []schema.Change {
	if base == nil && head == nil {
		return nil
	}
	if base == nil {
		return []schema.Change{{
			Type:        schema.ChangeTypeRequestBodyChanged,
			Path:        path,
			Method:      method,
			Location:    "request.body",
			Description: fmt.Sprintf("Request body was added to %s %s", method, path),
		}}
	}
	if head == nil {
		return []schema.Change{{
			Type:        schema.ChangeTypeRequestBodyChanged,
			Path:        path,
			Method:      method,
			Location:    "request.body",
			Description: fmt.Sprintf("Request body was removed from %s %s", method, path),
		}}
	}
	return DiffProperties(path, method, "request.body", base.Properties, head.Properties)
}
