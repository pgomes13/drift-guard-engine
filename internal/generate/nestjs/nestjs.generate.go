package nestjs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pgomes13/drift-guard-engine/internal/generate/node"
)

// Nest generates a swagger.json in outputDir for the NestJS project rooted at
// projectDir using @nestjs/swagger.
//
// Strategy (in order):
//  1. tsoa   — if tsoa.json is present, run `npx tsoa spec` and copy the result.
//  2. Script — look for an existing scripts/generate-swagger.ts (or .js).
//  3. Swagger — if @nestjs/swagger is installed, scaffold and run a temp script.
//  4. Error  — instruct the user to add a generation script.
func Nest(projectDir, outputDir string) error {
	// 1. tsoa
	if _, err := os.Stat(filepath.Join(projectDir, "tsoa.json")); err == nil {
		return node.TsoaSpec(projectDir, outputDir)
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
			return node.RunScript(projectDir, full, outputPath)
		}
	}

	// 3. @nestjs/swagger present — scaffold a temporary generation script.
	if nestHasSwaggerDep(projectDir) {
		if err := nestSwagger(projectDir, outputPath); err == nil {
			return nil
		}
		// Fall through to the actionable error on failure.
		// (The subprocess already printed diagnostics to stderr.)
	}

	// 4. No auto-generation possible — guide the user.
	return fmt.Errorf(
		"no OpenAPI generator found in %s\n\n"+
			"NestJS auto-generation requires a dedicated generation script.\n\n"+
			"Option A — add scripts/generate-swagger.ts that writes the spec to\n"+
			"  process.env.SWAGGER_OUTPUT without starting the full app, e.g. using\n"+
			"  a mocked AppModule that omits database providers.\n\n"+
			"Option B — use tsoa (no running app required):\n"+
			"  npm install --save-dev tsoa && npx tsoa init\n\n"+
			"Option C — use --cmd with a command that starts your app and generates the spec:\n"+
			`  drift-guard compare openapi --cmd "npm run generate-swagger" --output swagger.json`,
		projectDir,
	)
}

// nestSwagger scaffolds a temporary TypeScript script that uses @nestjs/swagger
// to generate the OpenAPI document, then runs it via ts-node.
func nestSwagger(projectDir, outputPath string) error {
	appModulePath, err := detectAppModulePath(projectDir)
	if err != nil {
		return err
	}

	script := buildNestSwaggerScript(appModulePath)

	tmp, err := os.CreateTemp(projectDir, ".dg-nestjs-swagger-*.ts")
	if err != nil {
		return fmt.Errorf("create temp script: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(script); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp script: %w", err)
	}
	tmp.Close()

	if err := node.RunScript(projectDir, tmp.Name(), outputPath); err != nil {
		return fmt.Errorf("nestjs/swagger auto-generation failed")
	}
	return nil
}

// detectAppModulePath returns the absolute path to the project's AppModule
// (without extension), trying common NestJS conventions.
func detectAppModulePath(projectDir string) (string, error) {
	candidates := []string{
		"src/app.module.ts",
		"src/app.module.js",
		"app.module.ts",
		"app.module.js",
	}
	for _, rel := range candidates {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err == nil {
			abs := filepath.Join(projectDir, strings.TrimSuffix(rel, filepath.Ext(rel)))
			return filepath.ToSlash(abs), nil
		}
	}
	return "", fmt.Errorf(
		"could not find AppModule (tried src/app.module.ts and others); " +
			"add scripts/generate-swagger.ts to your project instead",
	)
}

// buildNestSwaggerScript returns a TypeScript snippet that boots the NestJS app,
// generates the OpenAPI document, writes it to SWAGGER_OUTPUT, and exits.
func buildNestSwaggerScript(appModuleAbsPath string) string {
	return fmt.Sprintf(`// Load .env so ConfigModule / TypeORM can read credentials.
try { require('dotenv').config({ quiet: true }); } catch (_) {}

import { NestFactory } from '@nestjs/core';
import { SwaggerModule, DocumentBuilder } from '@nestjs/swagger';
import * as fs from 'fs';

// If app initialisation hangs (e.g. TypeORM retrying a lost DB connection),
// kill the process after 15 s with an actionable message.
const deadline = setTimeout(() => {
  process.stderr.write(
    '\ndrift-guard: NestJS app did not finish initialising within 15 s.\n' +
    'Ensure your database and other services are running, then retry.\n' +
    'Alternatively, add scripts/generate-swagger.ts with a mocked AppModule.\n',
  );
  process.exit(1);
}, 15_000);
deadline.unref();

async function generate(): Promise<void> {
  const { AppModule } = await import('%s');
  // abortOnError: false makes NestJS throw on failure instead of process.exit(1).
  const app = await NestFactory.create(AppModule, { abortOnError: false });
  clearTimeout(deadline);
  const config = new DocumentBuilder()
    .setTitle('API')
    .setVersion('1.0')
    .build();
  const document = SwaggerModule.createDocument(app, config);
  const output = process.env.SWAGGER_OUTPUT ?? 'swagger.json';
  fs.writeFileSync(output, JSON.stringify(document, null, 2));
  // Force exit so open handles (DB pools, queues) do not block the process.
  process.exit(0);
}

generate().catch((err) => {
  clearTimeout(deadline);
  process.stderr.write('Error: ' + String(err?.message ?? err) + '\n');
  process.exit(1);
});
`, appModuleAbsPath)
}

// nestHasSwaggerDep reports whether @nestjs/swagger is listed as a dependency
// in the project's package.json.
func nestHasSwaggerDep(projectDir string) bool {
	data, err := os.ReadFile(filepath.Join(projectDir, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Dependencies    map[string]json.RawMessage `json:"dependencies"`
		DevDependencies map[string]json.RawMessage `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return strings.Contains(string(data), `"@nestjs/swagger"`)
	}
	_, inDeps := pkg.Dependencies["@nestjs/swagger"]
	_, inDev := pkg.DevDependencies["@nestjs/swagger"]
	return inDeps || inDev
}
