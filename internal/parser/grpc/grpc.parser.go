// Package grpc reads and parses a .proto file into a normalized GRPCSchema.
package grpc

import (
	"fmt"
	"os"

	"github.com/DriftAgent/api-drift-engine/internal/parser/grpc/helpers"
	"github.com/DriftAgent/api-drift-engine/pkg/schema"

	"github.com/emicklei/proto"
)

// Parse reads a .proto file and returns a normalized GRPCSchema.
func Parse(path string) (*schema.GRPCSchema, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}
	defer f.Close()

	parser := proto.NewParser(f)
	doc, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parsing proto %s: %w", path, err)
	}

	v := &helpers.Visitor{}
	proto.Walk(doc, proto.WithService(func(s *proto.Service) { v.Visit(s) }),
		proto.WithMessage(func(m *proto.Message) { v.Visit(m) }))

	return helpers.Normalize(v), nil
}
