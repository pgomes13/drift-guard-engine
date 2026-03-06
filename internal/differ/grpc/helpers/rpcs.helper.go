package helpers

import (
	"fmt"

	"drift-guard-diff-engine/pkg/schema"
)

func indexRPCs(rpcs []schema.GRPCRPC) map[string]schema.GRPCRPC {
	m := make(map[string]schema.GRPCRPC, len(rpcs))
	for _, r := range rpcs {
		m[r.Name] = r
	}
	return m
}

// DiffRPCs compares two slices of RPCs within a service and returns all changes.
func DiffRPCs(serviceName string, base, head []schema.GRPCRPC) []schema.Change {
	var changes []schema.Change

	baseRPCs := indexRPCs(base)
	headRPCs := indexRPCs(head)

	for name, br := range baseRPCs {
		hr, exists := headRPCs[name]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCRPCRemoved,
				Location:    fmt.Sprintf("%s.%s", serviceName, name),
				Description: fmt.Sprintf("RPC '%s.%s' was removed", serviceName, name),
				Before:      name,
			})
			continue
		}

		if br.RequestType != hr.RequestType {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCRPCRequestTypeChanged,
				Location:    fmt.Sprintf("%s.%s", serviceName, name),
				Description: fmt.Sprintf("RPC '%s.%s' request type changed from '%s' to '%s'", serviceName, name, br.RequestType, hr.RequestType),
				Before:      br.RequestType,
				After:       hr.RequestType,
			})
		}

		if br.ResponseType != hr.ResponseType {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCRPCResponseTypeChanged,
				Location:    fmt.Sprintf("%s.%s", serviceName, name),
				Description: fmt.Sprintf("RPC '%s.%s' response type changed from '%s' to '%s'", serviceName, name, br.ResponseType, hr.ResponseType),
				Before:      br.ResponseType,
				After:       hr.ResponseType,
			})
		}

		if br.ClientStreaming != hr.ClientStreaming || br.ServerStreaming != hr.ServerStreaming {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCRPCStreamingChanged,
				Location:    fmt.Sprintf("%s.%s", serviceName, name),
				Description: fmt.Sprintf("RPC '%s.%s' streaming mode changed", serviceName, name),
				Before:      streamingLabel(br.ClientStreaming, br.ServerStreaming),
				After:       streamingLabel(hr.ClientStreaming, hr.ServerStreaming),
			})
		}
	}

	for name := range headRPCs {
		if _, exists := baseRPCs[name]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCRPCAdded,
				Location:    fmt.Sprintf("%s.%s", serviceName, name),
				Description: fmt.Sprintf("RPC '%s.%s' was added", serviceName, name),
				After:       name,
			})
		}
	}

	return changes
}

func streamingLabel(client, server bool) string {
	switch {
	case client && server:
		return "bidirectional streaming"
	case client:
		return "client streaming"
	case server:
		return "server streaming"
	default:
		return "unary"
	}
}
