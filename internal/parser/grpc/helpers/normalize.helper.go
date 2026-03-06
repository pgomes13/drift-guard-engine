package helpers

import (
	"drift-guard-diff-engine/pkg/schema"

	"github.com/emicklei/proto"
)

// Normalize converts collected proto elements into a normalized GRPCSchema.
func Normalize(v *Visitor) *schema.GRPCSchema {
	s := &schema.GRPCSchema{}

	for _, svc := range v.Services {
		service := schema.GRPCService{Name: svc.Name}
		for _, el := range svc.Elements {
			rpc, ok := el.(*proto.RPC)
			if !ok {
				continue
			}
			service.RPCs = append(service.RPCs, schema.GRPCRPC{
				Name:            rpc.Name,
				RequestType:     rpc.RequestType,
				ResponseType:    rpc.ReturnsType,
				ClientStreaming: rpc.StreamsRequest,
				ServerStreaming: rpc.StreamsReturns,
			})
		}
		s.Services = append(s.Services, service)
	}

	for _, msg := range v.Messages {
		message := schema.GRPCMessage{Name: msg.Name}
		for _, el := range msg.Elements {
			field, ok := el.(*proto.NormalField)
			if !ok {
				continue
			}
			message.Fields = append(message.Fields, schema.GRPCField{
				Name:     field.Name,
				Type:     field.Type,
				Number:   field.Sequence,
				Repeated: field.Repeated,
			})
		}
		s.Messages = append(s.Messages, message)
	}

	return s
}
