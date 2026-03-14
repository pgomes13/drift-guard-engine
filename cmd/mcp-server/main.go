package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/DriftAgent/api-drift-engine/pkg/compare"
	"github.com/DriftAgent/api-drift-engine/internal/languages"
	"github.com/DriftAgent/api-drift-engine/internal/reporter"
)

func main() {
	s := server.NewMCPServer(
		"drift-guard",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("diff_openapi",
		mcp.WithDescription("Compare two OpenAPI 3.x schema files (YAML or JSON) and report breaking and non-breaking changes"),
		mcp.WithString("base_file",
			mcp.Description("Path to the base (old) OpenAPI schema file"),
			mcp.Required(),
		),
		mcp.WithString("head_file",
			mcp.Description("Path to the head (new) OpenAPI schema file"),
			mcp.Required(),
		),
		mcp.WithString("format",
			mcp.Description("Output format: text (default), json, markdown"),
		),
	), diffOpenAPIHandler)

	s.AddTool(mcp.NewTool("diff_graphql",
		mcp.WithDescription("Compare two GraphQL SDL schema files (.graphql or .gql) and report breaking and non-breaking changes"),
		mcp.WithString("base_file",
			mcp.Description("Path to the base (old) GraphQL schema file"),
			mcp.Required(),
		),
		mcp.WithString("head_file",
			mcp.Description("Path to the head (new) GraphQL schema file"),
			mcp.Required(),
		),
		mcp.WithString("format",
			mcp.Description("Output format: text (default), json, markdown"),
		),
	), diffGraphQLHandler)

	s.AddTool(mcp.NewTool("diff_grpc",
		mcp.WithDescription("Compare two Protobuf (.proto) schema files and report breaking and non-breaking changes"),
		mcp.WithString("base_file",
			mcp.Description("Path to the base (old) .proto file"),
			mcp.Required(),
		),
		mcp.WithString("head_file",
			mcp.Description("Path to the head (new) .proto file"),
			mcp.Required(),
		),
		mcp.WithString("format",
			mcp.Description("Output format: text (default), json, markdown"),
		),
	), diffGRPCHandler)

	s.AddTool(mcp.NewTool("detect_project",
		mcp.WithDescription("Detect the project type, framework, and available schema types (OpenAPI, GraphQL, gRPC) for a given directory"),
		mcp.WithString("dir",
			mcp.Description("Absolute path to the project directory to inspect"),
			mcp.Required(),
		),
	), detectProjectHandler)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "drift-guard mcp server error: %v\n", err)
		os.Exit(1)
	}
}

func resolveFormat(req mcp.CallToolRequest) reporter.Format {
	switch strings.ToLower(req.GetString("format", "text")) {
	case "json":
		return reporter.FormatJSON
	case "markdown":
		return reporter.FormatMarkdown
	default:
		return reporter.FormatText
	}
}

func diffOpenAPIHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	base, err := req.RequireString("base_file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	head, err := req.RequireString("head_file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := compare.OpenAPI(base, head)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var buf bytes.Buffer
	if err := reporter.Write(&buf, result, resolveFormat(req)); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(buf.String()), nil
}

func diffGraphQLHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	base, err := req.RequireString("base_file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	head, err := req.RequireString("head_file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := compare.GraphQL(base, head)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var buf bytes.Buffer
	if err := reporter.Write(&buf, result, resolveFormat(req)); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(buf.String()), nil
}

func diffGRPCHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	base, err := req.RequireString("base_file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	head, err := req.RequireString("head_file")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := compare.GRPC(base, head)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var buf bytes.Buffer
	if err := reporter.Write(&buf, result, resolveFormat(req)); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(buf.String()), nil
}

func detectProjectHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir, err := req.RequireString("dir")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Project type: %s\n", info.TypeName)

	if gql := languages.DetectGraphQLInfo(dir); gql != nil {
		fmt.Fprintf(&sb, "GraphQL: detected (%s)\n", gql.TypeName)
	} else {
		fmt.Fprintln(&sb, "GraphQL: not detected")
	}

	if grpc := languages.DetectGRPCInfo(dir); grpc != nil {
		fmt.Fprintf(&sb, "gRPC: detected (%s)\n", grpc.TypeName)
	} else {
		fmt.Fprintln(&sb, "gRPC: not detected")
	}

	return mcp.NewToolResultText(sb.String()), nil
}
