// Package helpers provides internal types and normalization utilities for the gRPC parser.
package helpers

import "github.com/emicklei/proto"

// Visitor collects services and messages from a parsed proto document.
type Visitor struct {
	Services []proto.Service
	Messages []proto.Message
}

// Visit dispatches each proto element to the appropriate handler.
func (v *Visitor) Visit(n proto.Visitee) {
	switch el := n.(type) {
	case *proto.Service:
		v.Services = append(v.Services, *el)
	case *proto.Message:
		v.Messages = append(v.Messages, *el)
	}
}
