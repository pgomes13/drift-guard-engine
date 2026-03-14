package helpers

import (
	"fmt"
	"strings"

	"github.com/DriftAgent/api-drift-engine/pkg/schema"
)

// DiffStringSet computes added/removed entries between two string slices and
// emits Change records using the supplied change types and format functions.
func DiffStringSet(
	typeName string,
	base, head []string,
	removedType, addedType schema.ChangeType,
	locFn func(typeName, val string) string,
	removedDescFn func(typeName, val string) string,
	addedDescFn func(typeName, val string) string,
) []schema.Change {
	var changes []schema.Change
	baseSet := ToSet(base)
	headSet := ToSet(head)

	for v := range baseSet {
		if !headSet[v] {
			changes = append(changes, schema.Change{
				Type:        removedType,
				Location:    locFn(typeName, v),
				Description: removedDescFn(typeName, v),
				Before:      v,
			})
		}
	}
	for v := range headSet {
		if !baseSet[v] {
			changes = append(changes, schema.Change{
				Type:        addedType,
				Location:    locFn(typeName, v),
				Description: addedDescFn(typeName, v),
				After:       v,
			})
		}
	}
	return changes
}

func DiffGQLEnumValues(typeName string, base, head []string) []schema.Change {
	return DiffStringSet(typeName, base, head,
		schema.ChangeTypeGQLEnumValueRemoved,
		schema.ChangeTypeGQLEnumValueAdded,
		func(t, v string) string { return fmt.Sprintf("%s.%s", t, v) },
		func(t, v string) string { return fmt.Sprintf("Enum value '%s' was removed from '%s'", v, t) },
		func(t, v string) string { return fmt.Sprintf("Enum value '%s' was added to '%s'", v, t) },
	)
}

func DiffGQLUnionMembers(typeName string, base, head []string) []schema.Change {
	return DiffStringSet(typeName, base, head,
		schema.ChangeTypeGQLUnionMemberRemoved,
		schema.ChangeTypeGQLUnionMemberAdded,
		func(t, v string) string { return fmt.Sprintf("%s | %s", t, v) },
		func(t, v string) string { return fmt.Sprintf("Union member '%s' was removed from '%s'", v, t) },
		func(t, v string) string { return fmt.Sprintf("Union member '%s' was added to '%s'", v, t) },
	)
}

func DiffGQLInterfaces(typeName string, base, head []string) []schema.Change {
	return DiffStringSet(typeName, base, head,
		schema.ChangeTypeGQLInterfaceRemoved,
		schema.ChangeTypeGQLInterfaceAdded,
		func(t, v string) string { return fmt.Sprintf("%s implements %s", t, v) },
		func(t, v string) string { return fmt.Sprintf("Type '%s' no longer implements interface '%s'", t, v) },
		func(t, v string) string { return fmt.Sprintf("Type '%s' now implements interface '%s'", t, v) },
	)
}

func ToSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[strings.TrimSpace(s)] = true
	}
	return m
}
