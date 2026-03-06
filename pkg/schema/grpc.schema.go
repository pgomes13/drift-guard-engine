package schema

// GRPCSchema is the normalized representation of a protobuf/gRPC schema.
type GRPCSchema struct {
	Services []GRPCService
	Messages []GRPCMessage
}

// GRPCService represents a gRPC service definition.
type GRPCService struct {
	Name string
	RPCs []GRPCRPC
}

// GRPCRPC represents a single RPC method within a service.
type GRPCRPC struct {
	Name            string
	RequestType     string
	ResponseType    string
	ClientStreaming  bool
	ServerStreaming  bool
}

// GRPCMessage represents a protobuf message definition.
type GRPCMessage struct {
	Name   string
	Fields []GRPCField
}

// GRPCField represents a single field within a protobuf message.
type GRPCField struct {
	Name     string
	Type     string
	Number   int  // field number
	Repeated bool
	Optional bool
}

// gRPC change types
const (
	ChangeTypeGRPCServiceRemoved         ChangeType = "grpc_service_removed"
	ChangeTypeGRPCServiceAdded           ChangeType = "grpc_service_added"
	ChangeTypeGRPCRPCRemoved             ChangeType = "grpc_rpc_removed"
	ChangeTypeGRPCRPCAdded               ChangeType = "grpc_rpc_added"
	ChangeTypeGRPCRPCRequestTypeChanged  ChangeType = "grpc_rpc_request_type_changed"
	ChangeTypeGRPCRPCResponseTypeChanged ChangeType = "grpc_rpc_response_type_changed"
	ChangeTypeGRPCRPCStreamingChanged    ChangeType = "grpc_rpc_streaming_changed"
	ChangeTypeGRPCMessageRemoved         ChangeType = "grpc_message_removed"
	ChangeTypeGRPCMessageAdded           ChangeType = "grpc_message_added"
	ChangeTypeGRPCFieldRemoved           ChangeType = "grpc_field_removed"
	ChangeTypeGRPCFieldAdded             ChangeType = "grpc_field_added"
	ChangeTypeGRPCFieldTypeChanged       ChangeType = "grpc_field_type_changed"
	ChangeTypeGRPCFieldNumberChanged     ChangeType = "grpc_field_number_changed"
	ChangeTypeGRPCFieldLabelChanged      ChangeType = "grpc_field_label_changed"
)
