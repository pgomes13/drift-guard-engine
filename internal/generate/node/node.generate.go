package node

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Node generates a swagger.json in outputDir for the Node project rooted at
// projectDir.
//
// Strategy (in order):
//  1. tsoa   — if tsoa.json is present, run `npx tsoa spec` and copy the result.
//  2. Script — look for an existing scripts/generate-swagger.ts (or .js).
//  3. Error  — instruct the user to add tsoa.
func Node(projectDir, outputDir string) error {
	// 1. tsoa
	if _, err := os.Stat(filepath.Join(projectDir, "tsoa.json")); err == nil {
		return TsoaSpec(projectDir, outputDir)
	}

	outputPath := filepath.Join(outputDir, "swagger.json")

	// 2. Existing generation script.
	candidates := []string{
		"scripts/generate-swagger.ts",
		"scripts/generate-swagger.js",
		"src/generate-swagger.ts",
		"generate-swagger.ts",
	}
	for _, rel := range candidates {
		full := filepath.Join(projectDir, rel)
		if _, err := os.Stat(full); err == nil {
			return RunScript(projectDir, full, outputPath)
		}
	}

	// 3. No auto-generation possible — guide the user to set up tsoa.
	return fmt.Errorf(
		"no OpenAPI generator found in %s\n\n"+
			"Add tsoa for zero-config generation:\n\n"+
			"  npm install --save-dev tsoa\n"+
			"  npx tsoa init          # creates tsoa.json\n\n"+
			"Or use --cmd to provide your own generator:\n\n"+
			`  drift-guard compare openapi --cmd "node scripts/gen.js" --output swagger.json`,
		projectDir,
	)
}

// --------------------------------------------------------------------------
// tsoa
// --------------------------------------------------------------------------

type tsoaConfig struct {
	Spec struct {
		OutputDirectory  string `json:"outputDirectory"`
		SpecFileBaseName string `json:"specFileBaseName"`
	} `json:"spec"`
}

// TsoaSpec runs `npx tsoa spec` in projectDir and copies the result to outputDir.
func TsoaSpec(projectDir, outputDir string) error {
	cmd := exec.Command("npx", "tsoa", "spec")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npx tsoa spec: %w", err)
	}

	src, err := tsoaSpecFile(projectDir)
	if err != nil {
		return err
	}

	return copyFile(src, filepath.Join(outputDir, "swagger.json"))
}

func tsoaSpecFile(projectDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(projectDir, "tsoa.json"))
	if err != nil {
		return "", fmt.Errorf("read tsoa.json: %w", err)
	}

	var cfg tsoaConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse tsoa.json: %w", err)
	}

	outDir := cfg.Spec.OutputDirectory
	if outDir == "" {
		outDir = "."
	}
	baseName := cfg.Spec.SpecFileBaseName
	if baseName == "" {
		baseName = "swagger"
	}

	return filepath.Join(projectDir, filepath.FromSlash(outDir), baseName+".json"), nil
}

// --------------------------------------------------------------------------
// ts-node script runner
// --------------------------------------------------------------------------

// RunScript executes scriptPath with the project's settings, setting
// SWAGGER_OUTPUT to outputPath. Plain .js files are run with node; .ts files
// are run with ts-node.
func RunScript(projectDir, scriptPath, outputPath string) error {
	if filepath.Ext(scriptPath) == ".js" {
		return runJsScript(projectDir, scriptPath, outputPath)
	}
	return runTsScript(projectDir, scriptPath, outputPath)
}

func runJsScript(projectDir, scriptPath, outputPath string) error {
	cmd := exec.Command("node", scriptPath)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "SWAGGER_OUTPUT="+outputPath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run Node swagger generator: %w", err)
	}
	return nil
}

func runTsScript(projectDir, scriptPath, outputPath string) error {
	// --skip-project avoids inheriting the project's tsconfig (which may use
	// "module":"nodenext" / ESM, incompatible with require-hooks). We supply a
	// minimal known-good CJS config. tsconfig-paths/register still reads the
	// project tsconfig.json independently to resolve baseUrl/paths.
	args := []string{
		"ts-node", "--transpile-only", "--skip-project",
		"--compiler-options", `{"module":"CommonJS","moduleResolution":"node","experimentalDecorators":true,"emitDecoratorMetadata":true,"esModuleInterop":true,"allowSyntheticDefaultImports":true,"target":"ES2021","skipLibCheck":true}`,
	}
	if hasTsconfigPaths(projectDir) {
		args = append(args, "-r", "tsconfig-paths/register")
	}
	args = append(args, scriptPath)

	cmd := exec.Command("npx", args...)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "SWAGGER_OUTPUT="+outputPath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		hint := "Hint: create scripts/generate-swagger.ts in your project that writes the\n" +
			"OpenAPI document to process.env.SWAGGER_OUTPUT, then re-run drift-guard."
		if hasTsconfigPaths(projectDir) {
			hint = "Your tsconfig.json uses baseUrl/paths. Ensure tsconfig-paths is installed:\n\n" +
				"  npm install --save-dev tsconfig-paths"
		}
		return fmt.Errorf("run Node swagger generator: %w\n\n%s", err, hint)
	}
	return nil
}

// hasTsconfigPaths reports whether tsconfig.json in projectDir uses baseUrl or
// paths-based module resolution, which requires tsconfig-paths/register at
// runtime so that bare imports (e.g. "iam/auth") resolve correctly.
func hasTsconfigPaths(projectDir string) bool {
	data, err := os.ReadFile(filepath.Join(projectDir, "tsconfig.json"))
	if err != nil {
		return false
	}
	var tsconfig struct {
		CompilerOptions struct {
			BaseURL string                     `json:"baseUrl"`
			Paths   map[string]json.RawMessage `json:"paths"`
		} `json:"compilerOptions"`
	}
	if err := json.Unmarshal(data, &tsconfig); err != nil {
		// Likely JSONC — fall back to raw string search.
		return strings.Contains(string(data), `"baseUrl"`) ||
			strings.Contains(string(data), `"paths"`)
	}
	return tsconfig.CompilerOptions.BaseURL != "" ||
		len(tsconfig.CompilerOptions.Paths) > 0
}

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read generated spec %s: %w", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}
