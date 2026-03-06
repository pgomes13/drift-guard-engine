package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pgomes13/drift-guard-engine/internal/classifier"
	differgraphql "github.com/pgomes13/drift-guard-engine/internal/differ/graphql"
	differgrpc "github.com/pgomes13/drift-guard-engine/internal/differ/grpc"
	differopenapi "github.com/pgomes13/drift-guard-engine/internal/differ/openapi"
	"github.com/pgomes13/drift-guard-engine/internal/languages"
	parsergraphql "github.com/pgomes13/drift-guard-engine/internal/parser/graphql"
	parsergrpc "github.com/pgomes13/drift-guard-engine/internal/parser/grpc"
	parseropenapi "github.com/pgomes13/drift-guard-engine/internal/parser/openapi"
	"github.com/pgomes13/drift-guard-engine/internal/reporter"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Generate schemas from code and diff them",
	Long: `Generate API schemas by running against both the base and head revisions of the
repository, then diff the results.

For OpenAPI, schema generation is automatic (uses swaggo/swag) when --cmd is omitted.
Uses git worktree to check out the base ref without modifying the working tree.`,
}

// shared compare flags
var (
	flagGenBaseRef string
	flagGenCmd     string
	flagGenOutput  string
)

func addCompareFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&flagGenBaseRef, "base-ref", "origin/main", "Git ref to use as the base (before) revision")
	cmd.Flags().StringVar(&flagGenCmd, "cmd", "", "Command to run to generate the schema file (optional; auto-detected for OpenAPI)")
	cmd.Flags().StringVar(&flagGenOutput, "output", "", "Relative path where --cmd writes the schema file (required when --cmd is set)")
}

func addCompareFlagsRequired(cmd *cobra.Command) {
	addCompareFlags(cmd)
	_ = cmd.MarkFlagRequired("cmd")
	_ = cmd.MarkFlagRequired("output")
}

// --------------------------------------------------------------------------
// worktree helpers
// --------------------------------------------------------------------------

// setupWorktree checks out baseRef into a temp directory and returns
// (worktreeDir, cwd, cleanup, err).
func setupWorktree(baseRef string) (worktreeDir, cwd string, cleanup func(), err error) {
	cleanup = func() {}

	cwd, err = os.Getwd()
	if err != nil {
		return "", "", cleanup, fmt.Errorf("get working directory: %w", err)
	}

	worktreeDir, err = os.MkdirTemp("", "drift-guard-base-*")
	if err != nil {
		return "", "", cleanup, fmt.Errorf("create temp dir: %w", err)
	}
	cleanup = func() {
		exec.Command("git", "worktree", "remove", "--force", worktreeDir).Run() //nolint:errcheck
		os.RemoveAll(worktreeDir)
	}

	if out, err := exec.Command("git", "worktree", "add", "--detach", worktreeDir, baseRef).CombinedOutput(); err != nil {
		return "", "", cleanup, fmt.Errorf("git worktree add %s: %w\n%s", baseRef, err, out)
	}
	return worktreeDir, cwd, cleanup, nil
}

// runCompare runs genCmd in both the base worktree and current dir, returning
// the two schema file paths.
func runCompare(baseRef, genCmd, outputPath string) (basePath, headPath string, cleanup func(), err error) {
	worktreeDir, cwd, cleanup, err := setupWorktree(baseRef)
	if err != nil {
		return "", "", cleanup, err
	}

	if err := runCmd(genCmd, worktreeDir); err != nil {
		return "", "", cleanup, fmt.Errorf("generate base schema: %w", err)
	}
	if err := runCmd(genCmd, cwd); err != nil {
		return "", "", cleanup, fmt.Errorf("generate head schema: %w", err)
	}

	return filepath.Join(worktreeDir, outputPath), filepath.Join(cwd, outputPath), cleanup, nil
}

// runCompareAutoOpenAPI detects the project type and generates OpenAPI schemas
// for both base and head revisions.
func runCompareAutoOpenAPI(baseRef string) (basePath, headPath string, cleanup func(), err error) {
	worktreeDir, cwd, worktreeCleanup, err := setupWorktree(baseRef)
	if err != nil {
		return "", "", worktreeCleanup, err
	}

	baseOut, err := os.MkdirTemp("", "drift-guard-schema-base-*")
	if err != nil {
		return "", "", worktreeCleanup, err
	}
	headOut, err := os.MkdirTemp("", "drift-guard-schema-head-*")
	if err != nil {
		os.RemoveAll(baseOut)
		return "", "", worktreeCleanup, err
	}

	cleanup = func() {
		worktreeCleanup()
		os.RemoveAll(baseOut)
		os.RemoveAll(headOut)
	}

	gen, err := languages.DetectGenerator(cwd)
	if err != nil {
		return "", "", cleanup, err
	}

	if err := gen(worktreeDir, baseOut); err != nil {
		return "", "", cleanup, fmt.Errorf("generate base schema: %w", err)
	}
	if err := gen(cwd, headOut); err != nil {
		return "", "", cleanup, fmt.Errorf("generate head schema: %w", err)
	}

	return filepath.Join(baseOut, "swagger.yaml"), filepath.Join(headOut, "swagger.yaml"), cleanup, nil
}

func runCmd(command, dir string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// --------------------------------------------------------------------------
// compare openapi
// --------------------------------------------------------------------------

var compareOpenapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: "Generate OpenAPI schemas from code and diff them",
	Long: `Generates OpenAPI schemas from both the base ref and the current branch, then diffs them.

Auto mode (no --cmd): uses swaggo/swag with auto-detected project structure.
Custom mode (--cmd): runs your command and reads the schema from --output.`,
	Example: `  # Auto mode — zero config, requires swag
  drift-guard compare openapi

  # Custom generator
  drift-guard compare openapi --cmd "swag init --generalInfo cmd/api/main.go" --output docs/swagger.yaml

  # Against a specific base ref
  drift-guard compare openapi --base-ref origin/release --fail-on-breaking`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagGenCmd != "" && flagGenOutput == "" {
			return fmt.Errorf("--output is required when --cmd is set")
		}

		var basePath, headPath string
		var cleanup func()
		var err error

		if flagGenCmd != "" {
			basePath, headPath, cleanup, err = runCompare(flagGenBaseRef, flagGenCmd, flagGenOutput)
		} else {
			basePath, headPath, cleanup, err = runCompareAutoOpenAPI(flagGenBaseRef)
		}
		defer cleanup()
		if err != nil {
			return err
		}

		baseSchema, err := parseropenapi.Parse(basePath)
		if err != nil {
			return fmt.Errorf("parsing base: %w", err)
		}
		headSchema, err := parseropenapi.Parse(headPath)
		if err != nil {
			return fmt.Errorf("parsing head: %w", err)
		}

		result := classifier.Classify(basePath, headPath, differopenapi.Diff(baseSchema, headSchema))
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
// compare graphql
// --------------------------------------------------------------------------

var compareGraphqlCmd = &cobra.Command{
	Use:     "graphql",
	Short:   "Generate GraphQL schemas from code and diff them",
	Example: `  drift-guard compare graphql --cmd "go run ./tools/gen-schema.go" --output schema/schema.graphql --fail-on-breaking`,
	RunE: func(cmd *cobra.Command, args []string) error {
		basePath, headPath, cleanup, err := runCompare(flagGenBaseRef, flagGenCmd, flagGenOutput)
		defer cleanup()
		if err != nil {
			return err
		}

		baseSchema, err := parsergraphql.Parse(basePath)
		if err != nil {
			return fmt.Errorf("parsing base: %w", err)
		}
		headSchema, err := parsergraphql.Parse(headPath)
		if err != nil {
			return fmt.Errorf("parsing head: %w", err)
		}

		result := classifier.Classify(basePath, headPath, differgraphql.Diff(baseSchema, headSchema))
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
// compare grpc
// --------------------------------------------------------------------------

var compareGrpcCmd = &cobra.Command{
	Use:     "grpc",
	Short:   "Generate Protobuf schemas from code and diff them",
	Example: `  drift-guard compare grpc --cmd "buf export . --output /tmp/proto" --output api.proto --fail-on-breaking`,
	RunE: func(cmd *cobra.Command, args []string) error {
		basePath, headPath, cleanup, err := runCompare(flagGenBaseRef, flagGenCmd, flagGenOutput)
		defer cleanup()
		if err != nil {
			return err
		}

		baseSchema, err := parsergrpc.Parse(basePath)
		if err != nil {
			return fmt.Errorf("parsing base: %w", err)
		}
		headSchema, err := parsergrpc.Parse(headPath)
		if err != nil {
			return fmt.Errorf("parsing head: %w", err)
		}

		result := classifier.Classify(basePath, headPath, differgrpc.Diff(baseSchema, headSchema))
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
	addCompareFlags(compareOpenapiCmd)
	addOutputFlags(compareOpenapiCmd)

	addCompareFlagsRequired(compareGraphqlCmd)
	addOutputFlags(compareGraphqlCmd)

	addCompareFlagsRequired(compareGrpcCmd)
	addOutputFlags(compareGrpcCmd)

	compareCmd.AddCommand(compareOpenapiCmd, compareGraphqlCmd, compareGrpcCmd)
}
