package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"drift-guard-diff-engine/internal/classifier"
	"drift-guard-diff-engine/internal/differ"
	parsergraphql "drift-guard-diff-engine/internal/parser/graphql"
	parseropenapi "drift-guard-diff-engine/internal/parser/openapi"
	"drift-guard-diff-engine/internal/reporter"
)

func main() {
	var (
		baseFile    = flag.String("base", "", "Path to the base schema (OpenAPI YAML/JSON or GraphQL SDL)")
		headFile    = flag.String("head", "", "Path to the head schema (OpenAPI YAML/JSON or GraphQL SDL)")
		schemaType  = flag.String("type", "", "Schema type: openapi, graphql (auto-detected from extension if omitted)")
		format      = flag.String("format", "text", "Output format: text, json, github")
		failOnBreak = flag.Bool("fail-on-breaking", false, "Exit with code 1 if breaking changes are detected")
	)
	flag.Parse()

	if *baseFile == "" || *headFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --base and --head are required")
		flag.Usage()
		os.Exit(2)
	}

	kind := resolveSchemaType(*schemaType, *baseFile)

	switch kind {
	case "graphql":
		baseSchema, err := parsergraphql.Parse(*baseFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing base schema: %v\n", err)
			os.Exit(2)
		}
		headSchema, err := parsergraphql.Parse(*headFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing head schema: %v\n", err)
			os.Exit(2)
		}
		res := classifier.Classify(*baseFile, *headFile, differ.DiffGQL(baseSchema, headSchema))
		if err := reporter.Write(os.Stdout, res, reporter.Format(*format)); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
			os.Exit(2)
		}
		if *failOnBreak && reporter.HasBreakingChanges(res) {
			os.Exit(1)
		}

	default: // openapi
		baseSchema, err := parseropenapi.Parse(*baseFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing base schema: %v\n", err)
			os.Exit(2)
		}
		headSchema, err := parseropenapi.Parse(*headFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing head schema: %v\n", err)
			os.Exit(2)
		}
		res := classifier.Classify(*baseFile, *headFile, differ.Diff(baseSchema, headSchema))
		if err := reporter.Write(os.Stdout, res, reporter.Format(*format)); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
			os.Exit(2)
		}
		if *failOnBreak && reporter.HasBreakingChanges(res) {
			os.Exit(1)
		}
	}
}

// resolveSchemaType returns "graphql" or "openapi" based on the explicit flag
// or the file extension of the base file.
func resolveSchemaType(flagVal, baseFile string) string {
	if flagVal != "" {
		return strings.ToLower(flagVal)
	}
	ext := strings.ToLower(filepath.Ext(baseFile))
	if ext == ".graphql" || ext == ".gql" {
		return "graphql"
	}
	return "openapi"
}
