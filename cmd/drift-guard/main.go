package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pgomes13/drift-guard-engine/internal/classifier"
	differgraphql "github.com/pgomes13/drift-guard-engine/internal/differ/graphql"
	differgrpc "github.com/pgomes13/drift-guard-engine/internal/differ/grpc"
	differopenapi "github.com/pgomes13/drift-guard-engine/internal/differ/openapi"
	parsergraphql "github.com/pgomes13/drift-guard-engine/internal/parser/graphql"
	parsergrpc "github.com/pgomes13/drift-guard-engine/internal/parser/grpc"
	parseropenapi "github.com/pgomes13/drift-guard-engine/internal/parser/openapi"
	"github.com/pgomes13/drift-guard-engine/internal/reporter"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "drift-guard",
	Short: "API contract diff engine for OpenAPI, GraphQL, and gRPC schemas",
	Long: `drift-guard detects and classifies breaking vs. non-breaking changes
between two versions of an API schema.

Supported schema types: openapi, graphql, grpc`,
}

// shared flags
var (
	flagFormat      string
	flagFailOnBreak bool
)

func init() {
	rootCmd.AddCommand(openapiCmd, graphqlCmd, grpcCmd, compareCmd, generateCmd)
}

func addOutputFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&flagFormat, "format", "f", "text", "Output format: text, json, github")
	cmd.Flags().BoolVar(&flagFailOnBreak, "fail-on-breaking", false, "Exit with code 1 if breaking changes are detected")
}

// --------------------------------------------------------------------------
// openapi sub-command
// --------------------------------------------------------------------------

var openapiCmd = &cobra.Command{
	Use:   "openapi --base <file> --head <file>",
	Short: "Diff two OpenAPI 3.x schemas (YAML or JSON)",
	Example: `  drift-guard openapi --base api/v1.yaml --head api/v2.yaml
  drift-guard openapi --base old.json --head new.json --format json --fail-on-breaking`,
	RunE: func(cmd *cobra.Command, args []string) error {
		base, _ := cmd.Flags().GetString("base")
		head, _ := cmd.Flags().GetString("head")

		baseSchema, err := parseropenapi.Parse(base)
		if err != nil {
			return fmt.Errorf("parsing base: %w", err)
		}
		headSchema, err := parseropenapi.Parse(head)
		if err != nil {
			return fmt.Errorf("parsing head: %w", err)
		}

		result := classifier.Classify(base, head, differopenapi.Diff(baseSchema, headSchema))
		if err := reporter.Write(cmd.OutOrStdout(), result, reporter.Format(flagFormat)); err != nil {
			return err
		}
		if flagFailOnBreak && reporter.HasBreakingChanges(result) {
			os.Exit(1)
		}
		return nil
	},
}

// --------------------------------------------------------------------------
// graphql sub-command
// --------------------------------------------------------------------------

var graphqlCmd = &cobra.Command{
	Use:   "graphql --base <file> --head <file>",
	Short: "Diff two GraphQL SDL schemas",
	Example: `  drift-guard graphql --base schema/base.graphql --head schema/head.graphql
  drift-guard graphql --base old.gql --head new.gql --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		base, _ := cmd.Flags().GetString("base")
		head, _ := cmd.Flags().GetString("head")

		baseSchema, err := parsergraphql.Parse(base)
		if err != nil {
			return fmt.Errorf("parsing base: %w", err)
		}
		headSchema, err := parsergraphql.Parse(head)
		if err != nil {
			return fmt.Errorf("parsing head: %w", err)
		}

		result := classifier.Classify(base, head, differgraphql.Diff(baseSchema, headSchema))
		if err := reporter.Write(cmd.OutOrStdout(), result, reporter.Format(flagFormat)); err != nil {
			return err
		}
		if flagFailOnBreak && reporter.HasBreakingChanges(result) {
			os.Exit(1)
		}
		return nil
	},
}

// --------------------------------------------------------------------------
// grpc sub-command
// --------------------------------------------------------------------------

var grpcCmd = &cobra.Command{
	Use:   "grpc --base <file> --head <file>",
	Short: "Diff two Protobuf schemas (.proto)",
	Example: `  drift-guard grpc --base proto/base.proto --head proto/head.proto
  drift-guard grpc --base old.proto --head new.proto --format json --fail-on-breaking`,
	RunE: func(cmd *cobra.Command, args []string) error {
		base, _ := cmd.Flags().GetString("base")
		head, _ := cmd.Flags().GetString("head")

		baseSchema, err := parsergrpc.Parse(base)
		if err != nil {
			return fmt.Errorf("parsing base: %w", err)
		}
		headSchema, err := parsergrpc.Parse(head)
		if err != nil {
			return fmt.Errorf("parsing head: %w", err)
		}

		result := classifier.Classify(base, head, differgrpc.Diff(baseSchema, headSchema))
		if err := reporter.Write(cmd.OutOrStdout(), result, reporter.Format(flagFormat)); err != nil {
			return err
		}
		if flagFailOnBreak && reporter.HasBreakingChanges(result) {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	for _, cmd := range []*cobra.Command{openapiCmd, graphqlCmd, grpcCmd} {
		cmd.Flags().String("base", "", "Path to the base (before) schema file (required)")
		cmd.Flags().String("head", "", "Path to the head (after) schema file (required)")
		_ = cmd.MarkFlagRequired("base")
		_ = cmd.MarkFlagRequired("head")
		addOutputFlags(cmd)
	}
}
