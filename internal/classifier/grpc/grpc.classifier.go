package grpc

import "github.com/pgomes13/api-drift-engine/pkg/schema"

// Severity returns the severity for gRPC change types.
// The second return value is false if the change type is not a gRPC type.
func Severity(c schema.Change) (schema.Severity, bool) {
	switch c.Type {

	// Service removed — breaking; all its RPCs become unreachable
	case schema.ChangeTypeGRPCServiceRemoved:
		return schema.SeverityBreaking, true

	// Service added — non-breaking
	case schema.ChangeTypeGRPCServiceAdded:
		return schema.SeverityNonBreaking, true

	// RPC removed — breaking; callers can no longer invoke it
	case schema.ChangeTypeGRPCRPCRemoved:
		return schema.SeverityBreaking, true

	// RPC added — non-breaking
	case schema.ChangeTypeGRPCRPCAdded:
		return schema.SeverityNonBreaking, true

	// RPC request type changed — breaking; callers send the wrong message type
	case schema.ChangeTypeGRPCRPCRequestTypeChanged:
		return schema.SeverityBreaking, true

	// RPC response type changed — breaking; callers parse the wrong message type
	case schema.ChangeTypeGRPCRPCResponseTypeChanged:
		return schema.SeverityBreaking, true

	// Streaming mode changed — always breaking; wire protocol changes
	case schema.ChangeTypeGRPCRPCStreamingChanged:
		return schema.SeverityBreaking, true

	// Message removed — breaking; any RPC referencing it becomes invalid
	case schema.ChangeTypeGRPCMessageRemoved:
		return schema.SeverityBreaking, true

	// Message added — non-breaking
	case schema.ChangeTypeGRPCMessageAdded:
		return schema.SeverityNonBreaking, true

	// Field removed — breaking; senders may still populate it, receivers won't see it
	case schema.ChangeTypeGRPCFieldRemoved:
		return schema.SeverityBreaking, true

	// Field added — non-breaking in proto3 (new fields are optional by default)
	case schema.ChangeTypeGRPCFieldAdded:
		return schema.SeverityNonBreaking, true

	// Field type changed — breaking; wire encoding changes
	case schema.ChangeTypeGRPCFieldTypeChanged:
		return schema.SeverityBreaking, true

	// Field number changed — breaking; field numbers are the wire identity in protobuf
	case schema.ChangeTypeGRPCFieldNumberChanged:
		return schema.SeverityBreaking, true

	// Field label changed (singular ↔ repeated) — breaking; wire encoding and semantics change
	case schema.ChangeTypeGRPCFieldLabelChanged:
		return schema.SeverityBreaking, true

	default:
		return "", false
	}
}
