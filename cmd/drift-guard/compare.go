package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pgomes13/drift-guard-engine/internal/compare"
	"github.com/pgomes13/drift-guard-engine/internal/generate/node/express"
	"github.com/pgomes13/drift-guard-engine/internal/generate/node/nest"
	"github.com/pgomes13/drift-guard-engine/internal/languages"
	"github.com/pgomes13/drift-guard-engine/internal/reporter"
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
	fmt.Fprintf(os.Stderr, "Project detected: %s\n", info.TypeName)
	if !promptYesNo("Proceed?") {
		return nil
	}

	// --- Step 2: scaffold generation script if needed ---
	specFound := swaggerSpecExists(cwd)
	scriptFound := swaggerScriptExists(cwd)
	fmt.Fprintf(os.Stderr, "\nSwagger (openapi) file detected: %s\n", yesNo(specFound || scriptFound))

	if !(specFound || scriptFound) {
		switch info.TypeName {
		case "NestJS":
			if !promptYesNo("Proceed to add script?") {
				return nil
			}
			written, err := nest.ScaffoldNestSwaggerScript(cwd)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "scaffold written to %s\n", written)

		case "Express", "Node.js":
			if !express.HasTsoaControllers(cwd) {
				if !promptYesNo("Set up swagger-autogen for plain Express generation?") {
					break
				}
				written, err := express.ScaffoldSwaggerAutogenScript(cwd)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "script written to %s\n", written)
				fmt.Fprintf(os.Stderr, "Installing swagger-autogen...\n")
				if err := express.InstallSwaggerAutogen(cwd); err != nil {
					return err
				}
				break
			}
			if !promptYesNo("Set up tsoa for zero-config generation?") {
				return nil
			}
			written, err := express.ScaffoldTsoa(cwd)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "tsoa.json written to %s\n", written)
			fmt.Fprintf(os.Stderr, "Installing tsoa...\n")
			if err := express.InstallTsoa(cwd); err != nil {
				return err
			}
		}
	}

	// --- Step 3: generate head spec from current branch ---
	fmt.Fprintf(os.Stderr, "\nGenerating head spec from current branch...\n")
	headDir, err := os.MkdirTemp("", "drift-guard-head-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(headDir)

	if err := info.Generate(cwd, headDir); err != nil {
		return fmt.Errorf("generate head schema: %w", err)
	}
	headSpec, err := findSchemaFile(headDir)
	if err != nil {
		return fmt.Errorf("head spec: %w", err)
	}

	// --- Step 4: checkout base branch and generate base spec ---
	baseRef := resolveBaseRef("origin/main")
	fmt.Fprintf(os.Stderr, "Generating base spec from %s...\n", baseRef)

	worktreeDir, err := os.MkdirTemp("", "drift-guard-base-*")
	if err != nil {
		return fmt.Errorf("create worktree dir: %w", err)
	}
	defer func() {
		exec.Command("git", "worktree", "remove", "--force", worktreeDir).Run()
		os.RemoveAll(worktreeDir)
	}()

	if out, err := exec.Command("git", "worktree", "add", "--detach", worktreeDir, baseRef).CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add %s: %w\n%s", baseRef, err, out)
	}

	baseDir, err := os.MkdirTemp("", "drift-guard-base-out-*")
	if err != nil {
		return fmt.Errorf("create base output dir: %w", err)
	}
	defer os.RemoveAll(baseDir)

	if err := runGenerate(worktreeDir, baseDir); err != nil {
		return fmt.Errorf("generate base schema: %w", err)
	}
	baseSpec, err := findSchemaFile(baseDir)
	if err != nil {
		return fmt.Errorf("base spec: %w", err)
	}

	// --- Step 5: copy specs to output locations ---
	headOut := filepath.Join(filepath.Dir(flagHeadOut), "head.json")
	baseOut := filepath.Join(filepath.Dir(flagHeadOut), "base.json")
	if err := copyFile(headSpec, headOut); err != nil {
		return err
	}
	if err := copyFile(baseSpec, baseOut); err != nil {
		return err
	}

	// --- Step 6: diff base vs head ---
	fmt.Fprintf(os.Stderr, "\nDiffing %s vs %s...\n", baseOut, headOut)
	result, err := compare.OpenAPI(baseOut, headOut)
	if err != nil {
		return err
	}
	if err := reporter.Write(cmd.OutOrStdout(), result, reporter.Format(flagFormat)); err != nil {
		return err
	}
	if flagFailOnBreak && reporter.HasBreakingChanges(result) {
		os.Exit(1)
	}
	return nil
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
			fmt.Fprintf(os.Stderr, "base-ref %q not found, using %q\n", baseRef, candidate)
			return candidate
		}
	}
	return baseRef
}

func refExists(ref string) bool {
	return exec.Command("git", "rev-parse", "--verify", ref).Run() == nil
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

// swaggerScriptExists reports whether a swagger generation script or tsoa
// config is already present in the project.
func swaggerScriptExists(dir string) bool {
	candidates := []string{
		"tsoa.json",
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

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
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

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "schema written to %s\n", dst)
	return nil
}

var flagHeadOut string

func init() {
	compareCmd.Flags().StringVar(&flagHeadOut, "output-dir", ".", "Directory to write base.json and head.json")
	addOutputFlags(compareCmd)
}
