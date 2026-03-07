package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pgomes13/drift-guard-engine/internal/generate"
	"github.com/pgomes13/drift-guard-engine/internal/languages"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an API schema from source code",
	Long: `Interactively detect the project type, optionally scaffold a swagger
generation script, and build the OpenAPI schema.

Sub-commands are also available for non-interactive use:
  drift-guard generate openapi
  drift-guard generate swagger-script`,
	RunE: runGenerateWizard,
}

var flagGenOutPath string

func runGenerateWizard(cmd *cobra.Command, args []string) error {
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

	// --- Step 2: check for existing swagger spec or generation script ---
	specFound := swaggerSpecExists(cwd)
	scriptFound := swaggerScriptExists(cwd)
	fmt.Fprintf(os.Stderr, "\nSwagger (openapi) file detected: %s\n", yesNo(specFound || scriptFound))

	if !(specFound || scriptFound) && info.TypeName == "NestJS" {
		if !promptYesNo("Proceed to add script?") {
			return nil
		}
		written, err := generate.ScaffoldNestSwaggerScript(cwd)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "scaffold written to %s\n", written)
	}

	// --- Step 3: build the swagger spec ---
	if !promptYesNo("Build swagger spec?") {
		return nil
	}

	tmpDir, err := os.MkdirTemp("", "drift-guard-generate-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := info.Generate(cwd, tmpDir); err != nil {
		return fmt.Errorf("generate schema: %w", err)
	}

	return copySchema(tmpDir, flagGenOutPath)
}

// --------------------------------------------------------------------------
// generate openapi
// --------------------------------------------------------------------------

var generateOpenapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: "Generate an OpenAPI schema from the current project (non-interactive)",
	Example: `  # Auto-detect project type and write to swagger.json
  drift-guard generate openapi

  # Write to a custom path
  drift-guard generate openapi --output docs/openapi.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		gen, err := languages.DetectGenerator(cwd)
		if err != nil {
			return err
		}

		tmpDir, err := os.MkdirTemp("", "drift-guard-generate-*")
		if err != nil {
			return fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		if err := gen(cwd, tmpDir); err != nil {
			return fmt.Errorf("generate schema: %w", err)
		}

		return copySchema(tmpDir, flagGenOutPath)
	},
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

// --------------------------------------------------------------------------
// generate swagger-script
// --------------------------------------------------------------------------

var flagSwaggerScriptForce bool

var generateSwaggerScriptCmd = &cobra.Command{
	Use:   "swagger-script",
	Short: "Scaffold a scripts/generate-swagger.ts file in the current project",
	Long: `Write a starter scripts/generate-swagger.ts to the current directory.

The generated script uses NestFactory and SwaggerModule to produce an OpenAPI
document and write it to the path given by SWAGGER_OUTPUT — exactly what
drift-guard generate openapi expects.

Inline comments in the file explain how to mock heavy providers (TypeORM,
Redis, etc.) so the script can run without a live database.`,
	Example: `  # Scaffold in the current directory
  drift-guard generate swagger-script

  # Overwrite an existing file
  drift-guard generate swagger-script --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		outPath := filepath.Join(cwd, "scripts", "generate-swagger.ts")
		if !flagSwaggerScriptForce {
			if _, err := os.Stat(outPath); err == nil {
				return fmt.Errorf("%s already exists (use --force to overwrite)", outPath)
			}
		}

		written, err := generate.ScaffoldNestSwaggerScript(cwd)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "scaffold written to %s\n\n", written)
		fmt.Fprintf(os.Stderr, "Next steps:\n")
		fmt.Fprintf(os.Stderr, "  1. Review %s and adjust the AppModule import / provider mocks\n", written)
		fmt.Fprintf(os.Stderr, "  2. Run: drift-guard generate openapi\n")
		return nil
	},
}

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

// promptYesNo prints prompt and reads a Y/n response from stdin.
// An empty response (just Enter) defaults to Yes.
func promptYesNo(prompt string) bool {
	fmt.Fprintf(os.Stderr, "%s [Y/n]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
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

// swaggerScriptExists reports whether a swagger generation script is already
// present in the project.
func swaggerScriptExists(dir string) bool {
	candidates := []string{
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

// copySchema finds the generated schema file in srcDir and copies it to dst.
func copySchema(srcDir, dst string) error {
	candidates := []string{"swagger.yaml", "swagger.json", "docs.yaml", "docs.json"}
	for _, name := range candidates {
		src := filepath.Join(srcDir, name)
		if _, err := os.Stat(src); err == nil {
			return copyFile(src, dst)
		}
	}
	return fmt.Errorf("no schema file found in %s", srcDir)
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

func init() {
	generateCmd.PersistentFlags().StringVar(&flagGenOutPath, "output", "swagger.json", "Path to write the generated schema file")
	generateSwaggerScriptCmd.Flags().BoolVar(&flagSwaggerScriptForce, "force", false, "Overwrite the file if it already exists")
	generateCmd.AddCommand(generateOpenapiCmd)
	generateCmd.AddCommand(generateSwaggerScriptCmd)
}
