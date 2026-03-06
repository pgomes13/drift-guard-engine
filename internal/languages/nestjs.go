package languages

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateNestJS generates a swagger.json in outputDir for the NestJS project
// rooted at projectDir. It looks for an existing generation script first; if
// none is found it scaffolds a temporary one.
func GenerateNestJS(projectDir, outputDir string) error {
	outputPath := filepath.Join(outputDir, "swagger.json")

	// Look for an existing generation script committed to the repo.
	candidates := []string{
		"scripts/generate-swagger.ts",
		"scripts/generate-swagger.js",
		"src/generate-swagger.ts",
		"generate-swagger.ts",
	}
	for _, rel := range candidates {
		full := filepath.Join(projectDir, rel)
		if _, err := os.Stat(full); err == nil {
			return runNestJSScript(projectDir, full, outputPath)
		}
	}

	// No existing script — scaffold a temporary one inside the project so
	// that relative TypeScript imports resolve correctly.
	scriptPath := filepath.Join(projectDir, ".drift-guard-swagger-gen.ts")
	defer os.Remove(scriptPath) //nolint:errcheck

	content := buildNestJSScript(projectDir)
	if err := os.WriteFile(scriptPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("scaffold NestJS swagger script: %w", err)
	}
	return runNestJSScript(projectDir, scriptPath, outputPath)
}

// isJavascriptProject returns true when the project has a package.json but
// does NOT include @nestjs/swagger (i.e. it is a plain JS/TS project, not NestJS).
func isJavascriptProject(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Dependencies    map[string]json.RawMessage `json:"dependencies"`
		DevDependencies map[string]json.RawMessage `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return !strings.Contains(string(data), `"@nestjs/swagger"`)
	}
	_, inDeps := pkg.Dependencies["@nestjs/swagger"]
	_, inDev := pkg.DevDependencies["@nestjs/swagger"]
	return !inDeps && !inDev
}

// runNestJSScript executes a TypeScript generation script via npx ts-node and
// passes the output path through the SWAGGER_OUTPUT environment variable.
func runNestJSScript(projectDir, scriptPath, outputPath string) error {
	// Try with tsconfig-paths first (needed when path aliases are used).
	cmd := exec.Command("npx", "ts-node", "-r", "tsconfig-paths/register", scriptPath)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "SWAGGER_OUTPUT="+outputPath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Fallback: without tsconfig-paths.
	cmd = exec.Command("npx", "ts-node", scriptPath)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "SWAGGER_OUTPUT="+outputPath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"run NestJS swagger generator: %w\n\n"+
				"Hint: create scripts/generate-swagger.ts in your project that writes the\n"+
				"OpenAPI document to process.env.SWAGGER_OUTPUT, then re-run drift-guard.",
			err,
		)
	}
	return nil
}

// buildNestJSScript returns the content of a minimal NestJS swagger generation
// script that bootstraps the app in standalone mode and writes swagger.json.
func buildNestJSScript(projectDir string) string {
	// Determine the relative import path for AppModule.
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
