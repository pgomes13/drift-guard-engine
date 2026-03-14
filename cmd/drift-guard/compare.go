package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DriftAgent/api-drift-engine/pkg/compare"
	"github.com/DriftAgent/api-drift-engine/internal/generate/node/express"
	"github.com/DriftAgent/api-drift-engine/internal/generate/node/nest"
	"github.com/DriftAgent/api-drift-engine/internal/languages"
	"github.com/DriftAgent/api-drift-engine/internal/reporter"
	"github.com/DriftAgent/api-drift-engine/pkg/schema"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare API schemas between current branch and base branch",
	Long: `Detect the project type, generate an OpenAPI spec from the current branch
(head.json) and from the base branch (base.json), then diff them.`,
	RunE: runCompare,
}

func runCompare(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	info, err := languages.DetectProjectInfo(cwd)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s framework detected\n", info.TypeName)
	fmt.Fprintf(os.Stderr, "%s API detected\n", strings.Join(detectAPITypes(cwd), " | "))

	if gqlInfo := languages.DetectGraphQLInfo(cwd); gqlInfo != nil {
		if promptYesNo("Compare GraphQL schemas?") {
			return runGraphQLCompare(cmd, cwd, gqlInfo)
		}
	}

	if grpcInfo := languages.DetectGRPCInfo(cwd); grpcInfo != nil {
		if promptYesNo("Compare gRPC schemas?") {
			return runGRPCCompare(cmd, cwd, grpcInfo)
		}
	}

	if !promptYesNo("Proceed?") {
		return nil
	}

	// Scaffold generation script if needed.
	// Go projects use swag annotations directly — no scaffold step required.
	if openAPIScaffoldNeeded(info.TypeName) && !(swaggerSpecExists(cwd) || swaggerScriptExists(cwd)) {
		switch info.TypeName {
		case "NestJS":
			if _, err := nest.ScaffoldNestSwaggerScript(cwd); err != nil {
				return err
			}

		default: // Express / Node.js
			if !express.HasTsoaControllers(cwd) {
				if !promptYesNo("Set up swagger-autogen for plain Express generation?") {
					break
				}
				if _, err := express.ScaffoldSwaggerAutogenScript(cwd); err != nil {
					return err
				}
				if err := runStep("Installing swagger-autogen", func() error {
					return express.InstallSwaggerAutogen(cwd)
				}); err != nil {
					return err
				}
				break
			}
			if !promptYesNo("Set up tsoa for zero-config generation?") {
				return nil
			}
			if _, err := express.ScaffoldTsoa(cwd); err != nil {
				return err
			}
			if err := runStep("Installing tsoa", func() error {
				return express.InstallTsoa(cwd)
			}); err != nil {
				return err
			}
		}
	}

	express.SubprocessOutput = io.Discard

	tmpDir, cleanup, err := setupWorkspace(cwd)
	if err != nil {
		return err
	}
	defer cleanup()

	headOut := filepath.Join(tmpDir, "head.json")
	if err := runStep("Generating head spec", func() error {
		headGenDir := filepath.Join(tmpDir, "head-gen")
		if err := os.MkdirAll(headGenDir, 0o755); err != nil {
			return err
		}
		if err := info.Generate(cwd, headGenDir); err != nil {
			return fmt.Errorf("generate head schema: %w", err)
		}
		headSpec, err := findSchemaFile(headGenDir)
		if err != nil {
			return fmt.Errorf("head spec: %w", err)
		}
		return copyFile(headSpec, headOut)
	}); err != nil {
		return err
	}

	baseRef := resolveBaseRef("origin/main")
	baseOut := filepath.Join(tmpDir, "base.json")
	if err := runStep(fmt.Sprintf("Generating base spec (%s)", baseRef), func() error {
		worktreeDir, removeWorktree, err := addWorktree(tmpDir, baseRef)
		if err != nil {
			return err
		}
		defer removeWorktree()

		if rel := findSwaggerScript(cwd); rel != "" {
			if err := copyFile(filepath.Join(cwd, rel), filepath.Join(worktreeDir, rel)); err != nil {
				return fmt.Errorf("copy script to worktree: %w", err)
			}
		}
		if src := filepath.Join(cwd, ".env"); pathExists(src) {
			_ = copyFile(src, filepath.Join(worktreeDir, ".env"))
		}

		baseGenDir := filepath.Join(tmpDir, "base-gen")
		if err := os.MkdirAll(baseGenDir, 0o755); err != nil {
			return err
		}
		if err := runGenerate(worktreeDir, baseGenDir); err != nil {
			return fmt.Errorf("generate base schema: %w", err)
		}
		baseSpec, err := findSchemaFile(baseGenDir)
		if err != nil {
			return fmt.Errorf("base spec: %w", err)
		}
		return copyFile(baseSpec, baseOut)
	}); err != nil {
		return err
	}

	var diffResult schema.DiffResult
	if err := runStep("Comparing", func() error {
		var err error
		diffResult, err = compare.OpenAPI(baseOut, headOut)
		return err
	}); err != nil {
		return err
	}
	return writeResult(cmd, diffResult)
}

// runGraphQLCompare generates GraphQL schemas for head and base branches and diffs them.
func runGraphQLCompare(cmd *cobra.Command, cwd string, info *languages.GraphQLProjectInfo) error {
	express.SubprocessOutput = io.Discard

	tmpDir, cleanup, err := setupWorkspace(cwd)
	if err != nil {
		return err
	}
	defer cleanup()

	headOut := filepath.Join(tmpDir, "head.graphql")
	if err := runStep("Generating head GraphQL schema", func() error {
		headGenDir := filepath.Join(tmpDir, "head-gen")
		if err := os.MkdirAll(headGenDir, 0o755); err != nil {
			return err
		}
		if err := info.GenerateGQL(cwd, headGenDir); err != nil {
			return fmt.Errorf("generate head schema: %w", err)
		}
		headSchema, err := findGraphQLFile(headGenDir)
		if err != nil {
			return fmt.Errorf("head schema: %w", err)
		}
		return copyFile(headSchema, headOut)
	}); err != nil {
		return err
	}

	baseRef := resolveBaseRef("origin/main")
	baseOut := filepath.Join(tmpDir, "base.graphql")
	if err := runStep(fmt.Sprintf("Generating base GraphQL schema (%s)", baseRef), func() error {
		worktreeDir, removeWorktree, err := addWorktree(tmpDir, baseRef)
		if err != nil {
			return err
		}
		defer removeWorktree()

		baseGenDir := filepath.Join(tmpDir, "base-gen")
		if err := os.MkdirAll(baseGenDir, 0o755); err != nil {
			return err
		}
		if err := runGraphQLGenerate(worktreeDir, baseGenDir); err != nil {
			return fmt.Errorf("generate base schema: %w", err)
		}
		baseSchema, err := findGraphQLFile(baseGenDir)
		if err != nil {
			return fmt.Errorf("base schema: %w", err)
		}
		return copyFile(baseSchema, baseOut)
	}); err != nil {
		return err
	}

	var diffResult schema.DiffResult
	if err := runStep("Comparing", func() error {
		var err error
		diffResult, err = compare.GraphQL(baseOut, headOut)
		return err
	}); err != nil {
		return err
	}
	return writeResult(cmd, diffResult)
}

// runGRPCCompare finds gRPC proto schemas for head and base branches and diffs them.
func runGRPCCompare(cmd *cobra.Command, cwd string, info *languages.GRPCProjectInfo) error {
	tmpDir, cleanup, err := setupWorkspace(cwd)
	if err != nil {
		return err
	}
	defer cleanup()

	headOut := filepath.Join(tmpDir, "head.proto")
	if err := runStep("Collecting head gRPC schema", func() error {
		headGenDir := filepath.Join(tmpDir, "head-gen")
		if err := os.MkdirAll(headGenDir, 0o755); err != nil {
			return err
		}
		if err := info.GenerateRPC(cwd, headGenDir); err != nil {
			return fmt.Errorf("collect head schema: %w", err)
		}
		headProto, err := findProtoFile(headGenDir)
		if err != nil {
			return fmt.Errorf("head schema: %w", err)
		}
		return copyFile(headProto, headOut)
	}); err != nil {
		return err
	}

	baseRef := resolveBaseRef("origin/main")
	baseOut := filepath.Join(tmpDir, "base.proto")
	if err := runStep(fmt.Sprintf("Collecting base gRPC schema (%s)", baseRef), func() error {
		worktreeDir, removeWorktree, err := addWorktree(tmpDir, baseRef)
		if err != nil {
			return err
		}
		defer removeWorktree()

		baseGenDir := filepath.Join(tmpDir, "base-gen")
		if err := os.MkdirAll(baseGenDir, 0o755); err != nil {
			return err
		}
		if err := runGRPCGenerate(worktreeDir, baseGenDir); err != nil {
			return fmt.Errorf("collect base schema: %w", err)
		}
		baseProto, err := findProtoFile(baseGenDir)
		if err != nil {
			return fmt.Errorf("base schema: %w", err)
		}
		return copyFile(baseProto, baseOut)
	}); err != nil {
		return err
	}

	var diffResult schema.DiffResult
	if err := runStep("Comparing", func() error {
		var err error
		diffResult, err = compare.GRPC(baseOut, headOut)
		return err
	}); err != nil {
		return err
	}
	return writeResult(cmd, diffResult)
}

// openAPIScaffoldNeeded reports whether the project type requires swagger
// generation tooling to be scaffolded. Go projects use swag annotations
// directly and never need scaffolding.
func openAPIScaffoldNeeded(typeName string) bool {
	return !strings.HasPrefix(typeName, "Go")
}

// setupWorkspace creates the drift-guard/tmp directory and returns its path
// along with a cleanup function that removes the entire drift-guard directory.
func setupWorkspace(cwd string) (tmpDir string, cleanup func(), err error) {
	driftGuardDir := filepath.Join(cwd, "drift-guard")
	tmpDir = filepath.Join(driftGuardDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", nil, fmt.Errorf("create temp dir: %w", err)
	}
	return tmpDir, func() { os.RemoveAll(driftGuardDir) }, nil
}

// writeResult writes the diff report to cmd's stdout and exits with code 1
// if --fail-on-breaking is set and breaking changes were found.
func writeResult(cmd *cobra.Command, result schema.DiffResult) error {
	if err := reporter.Write(cmd.OutOrStdout(), result, reporter.Format(flagFormat)); err != nil {
		return err
	}
	if flagFailOnBreak && reporter.HasBreakingChanges(result) {
		os.Exit(1)
	}
	return nil
}

func init() {
	addOutputFlags(compareCmd)
}
