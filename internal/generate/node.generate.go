package generate

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
//  1. tsoa    — if tsoa.json is present, run `npx tsoa spec` and copy the result.
//  2. Script  — look for an existing scripts/generate-swagger.ts (or .js).
//  3. Scaffold — write a temporary NestJS bootstrap script and run it.
func Node(projectDir, outputDir string) error {
	// 1. tsoa
	if _, err := os.Stat(filepath.Join(projectDir, "tsoa.json")); err == nil {
		return tsoaSpec(projectDir, outputDir)
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
			return runScript(projectDir, full, outputPath)
		}
	}

	// 3. Scaffold a temporary NestJS bootstrap script.
	scriptPath := filepath.Join(projectDir, ".drift-guard-swagger-gen.ts")
	defer os.Remove(scriptPath) //nolint:errcheck

	content := nestJSScript(projectDir)
	if err := os.WriteFile(scriptPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("scaffold Node swagger script: %w", err)
	}
	return runScript(projectDir, scriptPath, outputPath)
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

func tsoaSpec(projectDir, outputDir string) error {
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

func runScript(projectDir, scriptPath, outputPath string) error {
	// Try with tsconfig-paths first; suppress output on this probe.
	probe := exec.Command("npx", "ts-node", "--transpile-only", "-r", "tsconfig-paths/register", scriptPath)
	probe.Dir = projectDir
	probe.Env = append(os.Environ(), "SWAGGER_OUTPUT="+outputPath)
	if err := probe.Run(); err == nil {
		return nil
	}

	// Fallback: without tsconfig-paths.
	cmd := exec.Command("npx", "ts-node", "--transpile-only", scriptPath)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "SWAGGER_OUTPUT="+outputPath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"run Node swagger generator: %w\n\n"+
				"Hint: create scripts/generate-swagger.ts in your project that writes the\n"+
				"OpenAPI document to process.env.SWAGGER_OUTPUT, then re-run drift-guard.",
			err,
		)
	}
	return nil
}

// --------------------------------------------------------------------------
// NestJS scaffold
// --------------------------------------------------------------------------

func nestJSScript(projectDir string) string {
	appModule := "./src/app.module"
	candidates := []string{"src/app.module.ts", "src/app.module.js"}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(projectDir, c)); err == nil {
			appModule = "./" + strings.TrimSuffix(filepath.ToSlash(c), filepath.Ext(c))
			break
		}
	}

	return fmt.Sprintf(`import { NestFactory } from '@nestjs/core';
import { SwaggerModule, DocumentBuilder } from '@nestjs/swagger';
import { writeFileSync } from 'fs';
import { AppModule } from '%s';

async function generate() {
  const app = await NestFactory.create(AppModule, { logger: false });
  const config = new DocumentBuilder()
    .setTitle('API')
    .setVersion('1.0')
    .build();
  const document = SwaggerModule.createDocument(app, config);
  const outputPath = process.env.SWAGGER_OUTPUT || 'swagger.json';
  writeFileSync(outputPath, JSON.stringify(document, null, 2));
  await app.close();
}

generate().catch(err => { console.error(err); process.exit(1); });
`, appModule)
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
