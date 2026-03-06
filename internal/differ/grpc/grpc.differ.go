// Package grpc computes the structural diff between two gRPC schemas,
// producing a flat list of Change values for downstream classification.
package grpc

import (
	"fmt"

	"drift-guard-engine/internal/differ/grpc/helpers"
	"drift-guard-engine/pkg/schema"
)

// Diff computes all changes between two normalized gRPC schemas.
func Diff(base, head *schema.GRPCSchema) []schema.Change {
	var changes []schema.Change

	baseServices := helpers.IndexServices(base)
	headServices := helpers.IndexServices(head)

	// Removed services
	for name, bs := range baseServices {
		hs, exists := headServices[name]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCServiceRemoved,
				Location:    name,
				Description: fmt.Sprintf("Service '%s' was removed", name),
				Before:      name,
			})
			continue
		}
		changes = append(changes, helpers.DiffRPCs(name, bs.RPCs, hs.RPCs)...)
	}

	// Added services
	for name := range headServices {
		if _, exists := baseServices[name]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCServiceAdded,
				Location:    name,
				Description: fmt.Sprintf("Service '%s' was added", name),
				After:       name,
			})
		}
	}

	// Messages
	baseMessages := helpers.IndexMessages(base)
	headMessages := helpers.IndexMessages(head)

	for name, bm := range baseMessages {
		hm, exists := headMessages[name]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCMessageRemoved,
				Location:    name,
				Description: fmt.Sprintf("Message '%s' was removed", name),
				Before:      name,
			})
			continue
		}
		changes = append(changes, helpers.DiffFields(name, bm.Fields, hm.Fields)...)
	}

	for name := range headMessages {
		if _, exists := baseMessages[name]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeGRPCMessageAdded,
				Location:    name,
				Description: fmt.Sprintf("Message '%s' was added", name),
				After:       name,
			})
		}
	}

	return changes
}
