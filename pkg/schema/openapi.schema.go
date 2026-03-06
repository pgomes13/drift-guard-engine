package schema

// Property represents a field in a JSON Schema object.
type Property struct {
	Name        string
	Type        string
	Format      string
	Required    bool
	Ref         string
	Description string
	Enum        []string
	Items       *Property  // for array types
	Properties  []Property // for object types
}

// Parameter represents an OpenAPI operation parameter.
type Parameter struct {
	Name     string
	In       string // query, path, header, cookie
	Required bool
	Type     string
	Format   string
	Ref      string
}

// RequestBody represents the request body of an operation.
type RequestBody struct {
	Required   bool
	Properties []Property
}

// Response represents a single status-code response.
type Response struct {
	StatusCode string
	Properties []Property
}

// Operation represents a single HTTP method on a path.
type Operation struct {
	Method      string
	OperationID string
	Parameters  []Parameter
	RequestBody *RequestBody
	Responses   []Response
}

// Endpoint represents a path with all its operations.
type Endpoint struct {
	Path       string
	Operations []Operation
}

// Schema is the normalized representation of an OpenAPI spec.
type Schema struct {
	Title     string
	Version   string
	Endpoints []Endpoint
}

// OpenAPI change types
const (
	ChangeTypeEndpointRemoved      ChangeType = "endpoint_removed"
	ChangeTypeEndpointAdded        ChangeType = "endpoint_added"
	ChangeTypeMethodRemoved        ChangeType = "method_removed"
	ChangeTypeMethodAdded          ChangeType = "method_added"
	ChangeTypeParamRemoved         ChangeType = "param_removed"
	ChangeTypeParamAdded           ChangeType = "param_added"
	ChangeTypeParamTypeChanged     ChangeType = "param_type_changed"
	ChangeTypeParamRequiredChanged ChangeType = "param_required_changed"
	ChangeTypeRequestBodyChanged   ChangeType = "request_body_changed"
	ChangeTypeResponseChanged      ChangeType = "response_changed"
	ChangeTypeFieldRemoved         ChangeType = "field_removed"
	ChangeTypeFieldAdded           ChangeType = "field_added"
	ChangeTypeFieldTypeChanged     ChangeType = "field_type_changed"
	ChangeTypeFieldRequiredChanged ChangeType = "field_required_changed"
)
