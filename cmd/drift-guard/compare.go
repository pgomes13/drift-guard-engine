package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/pgomes13/drift-guard-engine/internal/compare"
	"github.com/pgomes13/drift-guard-engine/internal/generate/node/express"
	"github.com/pgomes13/drift-guard-engine/internal/generate/node/nest"
	"github.com/pgomes13/drift-guard-engine/internal/languages"
	"github.com/pgomes13/drift-guard-engine/internal/reporter"
	"github.com/pgomes13/drift-guard-engine/pkg/schema"
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

	// --- Step 1: detect project type ---
	info, err := languages.DetectProjectInfo(cwd)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s framework detected\n", info.TypeName)
	fmt.Fprintf(os.Stderr, "%s API detected\n", strings.Join(detectAPITypes(cwd), " | "))

	// --- Step 1b: offer GraphQL comparison if detected ---
	if gqlInfo := languages.DetectGraphQLInfo(cwd); gqlInfo != nil {
		if promptYesNo("Compare GraphQL schemas?") {
			return runGraphQLCompare(cmd, cwd, gqlInfo)
		}
	}

	if !promptYesNo("Proceed?") {
		return nil
	}

	// All drift-guard generated artifacts live under drift-guard/ and are
	// removed after the comparison completes.
	driftGuardDir := filepath.Join(cwd, "drift-guard")

	// --- Step 2: scaffold generation script if needed ---
	specFound := swaggerSpecExists(cwd)
	scriptFound := swaggerScriptExists(cwd)

	if !(specFound || scriptFound) {
		switch info.TypeName {
		case "NestJS":
			if !promptYesNo("Proceed to add script?") {
				return nil
			}
			if _, err := nest.ScaffoldNestSwaggerScript(driftGuardDir); err != nil {
				return err
			}

		case "Express", "Node.js":
			if !express.HasTsoaControllers(cwd) {
				if !promptYesNo("Set up swagger-autogen for plain Express generation?") {
					break
				}
				if _, err := express.ScaffoldSwaggerAutogenScript(driftGuardDir); err != nil {
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

	// Suppress subprocess output — spinner is active from here on.
	express.SubprocessOutput = io.Discard

	// --- Step 3: create temp dir inside drift-guard/ ---
	tmpDir := filepath.Join(driftGuardDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(driftGuardDir)

	// --- Step 4: generate head spec from current branch ---
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

	// --- Step 5: checkout base branch and generate base spec ---
	baseRef := resolveBaseRef("origin/main")
	baseOut := filepath.Join(tmpDir, "base.json")
	worktreeDir := filepath.Join(tmpDir, "worktree")
	defer exec.Command("git", "worktree", "remove", "--force", worktreeDir).Run()

	if err := runStep(fmt.Sprintf("Generating base spec (%s)", baseRef), func() error {
		if out, err := exec.Command("git", "worktree", "add", "--detach", worktreeDir, baseRef).CombinedOutput(); err != nil {
			return fmt.Errorf("git worktree add %s: %w\n%s", baseRef, err, out)
		}
		if rel := findSwaggerScript(cwd); rel != "" {
			src := filepath.Join(cwd, rel)
			dst := filepath.Join(worktreeDir, rel)
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("copy script to worktree: %w", err)
			}
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

	// --- Step 6: diff base vs head ---
	var diffResult schema.DiffResult
	if err := runStep("Comparing", func() error {
		var err error
		diffResult, err = compare.OpenAPI(baseOut, headOut)
		return err
	}); err != nil {
		return err
	}
	if err := reporter.Write(cmd.OutOrStdout(), diffResult, reporter.Format(flagFormat)); err != nil {
		return err
	}
	if flagFailOnBreak && reporter.HasBreakingChanges(diffResult) {
		os.Exit(1)
	}
	return nil
}

// runGraphQLCompare generates GraphQL schemas for head and base branches and diffs them.
func runGraphQLCompare(cmd *cobra.Command, cwd string, info *languages.GraphQLProjectInfo) error {
	driftGuardDir := filepath.Join(cwd, "drift-guard")
	tmpDir := filepath.Join(driftGuardDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(driftGuardDir)

	express.SubprocessOutput = io.Discard

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
	worktreeDir := filepath.Join(tmpDir, "worktree")
	defer exec.Command("git", "worktree", "remove", "--force", worktreeDir).Run()

	if err := runStep(fmt.Sprintf("Generating base GraphQL schema (%s)", baseRef), func() error {
		if out, err := exec.Command("git", "worktree", "add", "--detach", worktreeDir, baseRef).CombinedOutput(); err != nil {
			return fmt.Errorf("git worktree add %s: %w\n%s", baseRef, err, out)
		}
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
	if err := reporter.Write(cmd.OutOrStdout(), diffResult, reporter.Format(flagFormat)); err != nil {
		return err
	}
	if flagFailOnBreak && reporter.HasBreakingChanges(diffResult) {
		os.Exit(1)
	}
	return nil
}

// runGraphQLGenerate auto-detects the project type and generates a GraphQL schema
// for the project at projectDir, writing output files into outputDir.
func runGraphQLGenerate(projectDir, outputDir string) error {
	gen, err := languages.DetectGraphQLGenerator(projectDir)
	if err != nil {
		return err
	}
	return gen(projectDir, outputDir)
}

// findGraphQLFile finds the generated GraphQL schema file in dir.
func findGraphQLFile(dir string) (string, error) {
	for _, name := range []string{"schema.graphql", "schema.gql"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no GraphQL schema file found in %s", dir)
}

// runGenerate auto-detects the project type and generates an OpenAPI schema
// for the project at projectDir, writing output files into outputDir.
func runGenerate(projectDir, outputDir string) error {
	gen, err := languages.DetectGenerator(projectDir)
	if err != nil {
		return err
	}
	return gen(projectDir, outputDir)
}

// resolveBaseRef returns the first of the candidates that exists as a valid git ref.
func resolveBaseRef(baseRef string) string {
	if refExists(baseRef) {
		return baseRef
	}
	for _, candidate := range []string{"origin/master", "HEAD~1"} {
		if refExists(candidate) {
			return candidate
		}
	}
	return baseRef
}

func refExists(ref string) bool {
	return exec.Command("git", "rev-parse", "--verify", ref).Run() == nil
}

// detectAPITypes returns the API types present in the project.
// REST is always included; GraphQL and gRPC are added when detected.
func detectAPITypes(dir string) []string {
	types := []string{"REST"}
	if languages.DetectGraphQLInfo(dir) != nil {
		types = append(types, "GraphQL")
	}
	if hasProtoFiles(dir) {
		types = append(types, "gRPC")
	}
	return types
}

// hasProtoFiles reports whether any .proto files exist under dir.
func hasProtoFiles(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".proto") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// --------------------------------------------------------------------------
// progress spinner
// --------------------------------------------------------------------------

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// runStep displays an animated spinner while fn executes, then prints ✓ or ✗.
func runStep(label string, fn func() error) error {
	stop := make(chan struct{})
	spinDone := make(chan struct{})
	var fnErr error

	go func() {
		defer close(spinDone)
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		fmt.Fprintf(os.Stderr, "  %s  %s", spinnerFrames[0], label)
		for {
			select {
			case <-stop:
				if fnErr != nil {
					fmt.Fprintf(os.Stderr, "\r  ✗  %s\n", label)
				} else {
					fmt.Fprintf(os.Stderr, "\r  ✓  %s\n", label)
				}
				return
			case <-ticker.C:
				i++
				fmt.Fprintf(os.Stderr, "\r  %s  %s", spinnerFrames[i%len(spinnerFrames)], label)
			}
		}
	}()

	fnErr = fn()
	close(stop)
	<-spinDone
	return fnErr
}

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

// stdinReader is shared across all promptYesNo calls so that bufio buffering
// doesn't consume input intended for a subsequent prompt.
var stdinReader = bufio.NewReader(os.Stdin)

// promptYesNo prints prompt and reads a Y/n response from stdin.
// An empty response (just Enter) defaults to Yes.
func promptYesNo(prompt string) bool {
	fmt.Fprintf(os.Stderr, "%s [Y/n]: ", prompt)
	line, _ := stdinReader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "" || line == "y" || line == "yes"
}

// swaggerSpecExists reports whether a swagger/openapi spec file already exists
// in common locations under dir.
func swaggerSpecExists(dir string) bool {
	candidates := []string{
		"swagger.json", "swagger.yaml", "swagger.yml",
		"openapi.json", "openapi.yaml", "openapi.yml",
		"docs/swagger.json", "docs/swagger.yaml",
		"api/swagger.json", "api/openapi.json",
	}
	for _, rel := range candidates {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			return true
		}
	}
	return false
}

// findSwaggerScript returns the relative path of the first swagger generation
// script found in dir, or empty string if none is found.
func findSwaggerScript(dir string) string {
	candidates := []string{
		"drift-guard/scripts/generate-swagger.ts",
		"drift-guard/scripts/generate-swagger.js",
		"scripts/generate-swagger.ts",
		"scripts/generate-swagger.js",
		"src/generate-swagger.ts",
		"generate-swagger.ts",
	}
	for _, rel := range candidates {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			return rel
		}
	}
	return ""
}

func swaggerScriptExists(dir string) bool {
	candidates := []string{
		"tsoa.json",
		"drift-guard/scripts/generate-swagger.ts",
		"drift-guard/scripts/generate-swagger.js",
		"scripts/generate-swagger.ts",
		"scripts/generate-swagger.js",
		"src/generate-swagger.ts",
		"generate-swagger.ts",
	}
	for _, rel := range candidates {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			return true
		}
	}
	return false
}

// findSchemaFile finds the generated schema file in dir.
func findSchemaFile(dir string) (string, error) {
	for _, name := range []string{"swagger.yaml", "swagger.json", "docs.yaml", "docs.json"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no schema file found in %s", dir)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func init() {
	addOutputFlags(compareCmd)
}
