package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pgomes13/drift-guard-engine/internal/languages"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an API schema from source code",
	Long: `Generate an API schema from the current project's source code.

The project type is auto-detected from the directory contents:
  - Go:     go.mod present → swaggo/swag
  - NestJS: package.json with @nestjs/swagger → npx ts-node`,
}

var flagGenOutPath string

// --------------------------------------------------------------------------
// generate openapi
// --------------------------------------------------------------------------

var generateOpenapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: "Generate an OpenAPI schema from the current project",
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

		// Generate into a temp dir, then copy the result to --output.
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

// copySchema finds the generated schema file in srcDir and copies it to dst.
// It tries common filenames produced by the generators.
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
	generateOpenapiCmd.Flags().StringVar(&flagGenOutPath, "output", "swagger.json", "Path to write the generated schema file")
	generateCmd.AddCommand(generateOpenapiCmd)
}
