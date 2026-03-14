package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/DriftaBot/driftabot-engine/api/drift-agent/v1"
	"github.com/DriftaBot/driftabot-engine/internal/classifier"
	differgraphql "github.com/DriftaBot/driftabot-engine/internal/differ/graphql"
	differopenapi "github.com/DriftaBot/driftabot-engine/internal/differ/openapi"
	parsergraphql "github.com/DriftaBot/driftabot-engine/internal/parser/graphql"
	parseropenapi "github.com/DriftaBot/driftabot-engine/internal/parser/openapi"
	"github.com/DriftaBot/driftabot-engine/pkg/schema"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedDiffEngineServer
}

func (s *server) Diff(_ context.Context, req *pb.DiffRequest) (*pb.DiffResponse, error) {
	if len(req.BaseContent) == 0 || len(req.HeadContent) == 0 {
		return nil, status.Error(codes.InvalidArgument, "base_content and head_content are required")
	}

	basePath, err := writeTempFile(req.BaseContent, req.BaseName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "write base temp file: %v", err)
	}
	defer os.Remove(basePath)

	headPath, err := writeTempFile(req.HeadContent, req.HeadName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "write head temp file: %v", err)
	}
	defer os.Remove(headPath)

	schemaType := resolveSchemaType(req.Type, req.BaseName)

	result, err := runDiff(schemaType, basePath, headPath)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "diff failed: %v", err)
	}

	return toProto(result), nil
}

func runDiff(schemaType, basePath, headPath string) (schema.DiffResult, error) {
	switch schemaType {
	case "graphql":
		base, err := parsergraphql.Parse(basePath)
		if err != nil {
			return schema.DiffResult{}, fmt.Errorf("parse base: %w", err)
		}
		head, err := parsergraphql.Parse(headPath)
		if err != nil {
			return schema.DiffResult{}, fmt.Errorf("parse head: %w", err)
		}
		return classifier.Classify(basePath, headPath, differgraphql.Diff(base, head)), nil
	default: // openapi
		base, err := parseropenapi.Parse(basePath)
		if err != nil {
			return schema.DiffResult{}, fmt.Errorf("parse base: %w", err)
		}
		head, err := parseropenapi.Parse(headPath)
		if err != nil {
			return schema.DiffResult{}, fmt.Errorf("parse head: %w", err)
		}
		return classifier.Classify(basePath, headPath, differopenapi.Diff(base, head)), nil
	}
}

func writeTempFile(content []byte, originalName string) (string, error) {
	ext := filepath.Ext(originalName)
	f, err := os.CreateTemp("", "driftabot-*"+ext)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(content); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func resolveSchemaType(explicit, baseName string) string {
	if explicit != "" {
		return strings.ToLower(explicit)
	}
	ext := strings.ToLower(filepath.Ext(baseName))
	if ext == ".graphql" || ext == ".gql" {
		return "graphql"
	}
	return "openapi"
}

func toProto(r schema.DiffResult) *pb.DiffResponse {
	changes := make([]*pb.Change, len(r.Changes))
	for i, c := range r.Changes {
		changes[i] = &pb.Change{
			Type:        string(c.Type),
			Severity:    string(c.Severity),
			Path:        c.Path,
			Method:      c.Method,
			Location:    c.Location,
			Description: c.Description,
			Before:      c.Before,
			After:       c.After,
		}
	}
	return &pb.DiffResponse{
		Changes: changes,
		Summary: &pb.Summary{
			Total:       int32(r.Summary.Total),
			Breaking:    int32(r.Summary.Breaking),
			NonBreaking: int32(r.Summary.NonBreaking),
			Info:        int32(r.Summary.Info),
		},
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterDiffEngineServer(srv, &server{})

	log.Printf("driftabot gRPC server listening on :%s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
