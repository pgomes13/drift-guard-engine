// Package graphql computes the structural diff between two GraphQL schemas,
// producing a flat list of Change values for downstream classification.
package graphql

import (
	"fmt"

	"github.com/DriftaBot/driftabot-engine/internal/differ/graphql/helpers"
	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

// Diff computes all changes between two normalized GraphQL schemas.
func Diff(base, head *schema.GQLSchema) []schema.Change {
	var changes []schema.Change

	baseTypes := helpers.IndexGQLTypes(base)
	headTypes := helpers.IndexGQLTypes(head)

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
		changes = append(changes, helpers.DiffGQLType(bt, ht)...)
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
