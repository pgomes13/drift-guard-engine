package schema

// GQLTypeKind mirrors GraphQL's named type categories.
type GQLTypeKind string

const (
	GQLTypeKindObject    GQLTypeKind = "OBJECT"
	GQLTypeKindInterface GQLTypeKind = "INTERFACE"
	GQLTypeKindUnion     GQLTypeKind = "UNION"
	GQLTypeKindEnum      GQLTypeKind = "ENUM"
	GQLTypeKindInput     GQLTypeKind = "INPUT"
	GQLTypeKindScalar    GQLTypeKind = "SCALAR"
)

// GQLArgument is an argument on a field or directive.
type GQLArgument struct {
	Name         string
	Type         string // e.g. "String!", "[ID!]!"
	DefaultValue string // empty if none
	Description  string
}

// GQLField is a field on an Object, Interface, or Input type.
type GQLField struct {
	Name        string
	Type        string // e.g. "String!", "[User!]!"
	Arguments   []GQLArgument
	Deprecated  bool
	Description string
}

// GQLType is the normalized representation of any named GraphQL type.
type GQLType struct {
	Name        string
	Kind        GQLTypeKind
	Description string
	// Object / Interface
	Fields     []GQLField
	Interfaces []string // implemented interfaces (Object only)
	// Union
	Members []string
	// Enum
	Values []string
}

// GQLSchema is the full normalized GraphQL schema.
type GQLSchema struct {
	Types []GQLType
}

// GraphQL change types
const (
	ChangeTypeGQLTypeRemoved           ChangeType = "gql_type_removed"
	ChangeTypeGQLTypeAdded             ChangeType = "gql_type_added"
	ChangeTypeGQLTypeKindChanged       ChangeType = "gql_type_kind_changed"
	ChangeTypeGQLFieldRemoved          ChangeType = "gql_field_removed"
	ChangeTypeGQLFieldAdded            ChangeType = "gql_field_added"
	ChangeTypeGQLFieldTypeChanged      ChangeType = "gql_field_type_changed"
	ChangeTypeGQLFieldDeprecated       ChangeType = "gql_field_deprecated"
	ChangeTypeGQLArgRemoved            ChangeType = "gql_arg_removed"
	ChangeTypeGQLArgAdded              ChangeType = "gql_arg_added"
	ChangeTypeGQLArgTypeChanged        ChangeType = "gql_arg_type_changed"
	ChangeTypeGQLArgDefaultChanged     ChangeType = "gql_arg_default_changed"
	ChangeTypeGQLEnumValueRemoved      ChangeType = "gql_enum_value_removed"
	ChangeTypeGQLEnumValueAdded        ChangeType = "gql_enum_value_added"
	ChangeTypeGQLUnionMemberRemoved    ChangeType = "gql_union_member_removed"
	ChangeTypeGQLUnionMemberAdded      ChangeType = "gql_union_member_added"
	ChangeTypeGQLInterfaceRemoved      ChangeType = "gql_interface_removed"
	ChangeTypeGQLInterfaceAdded        ChangeType = "gql_interface_added"
	ChangeTypeGQLInputFieldRemoved     ChangeType = "gql_input_field_removed"
	ChangeTypeGQLInputFieldAdded       ChangeType = "gql_input_field_added"
	ChangeTypeGQLInputFieldTypeChanged ChangeType = "gql_input_field_type_changed"
)
