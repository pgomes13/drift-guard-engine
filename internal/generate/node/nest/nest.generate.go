package nest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DriftaBot/driftabot-engine/internal/generate/node/express"
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
		return express.TsoaSpec(projectDir, outputDir)
	}

	outputPath := filepath.Join(outputDir, "swagger.json")

	// 2. Existing generation script.
	candidates := []string{
		"drift-agent/scripts/generate-swagger.ts",
		"drift-agent/scripts/generate-swagger.js",
		"scripts/generate-swagger.ts",
		"scripts/generate-swagger.js",
		"src/generate-swagger.ts",
		"generate-swagger.ts",
	}
	for _, rel := range candidates {
		full := filepath.Join(projectDir, rel)
		if _, err := os.Stat(full); err == nil {
			return express.RunScript(projectDir, full, outputPath)
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
			`  drift-agent compare openapi --cmd "npm run generate-swagger" --output swagger.json`,
		projectDir,
	)
}

// nestSwagger scaffolds a temporary TypeScript script that uses @nestjs/swagger
// to generate the OpenAPI document, then runs it via ts-node with a CJS preload
// that patches database drivers before any app module is imported.
func nestSwagger(projectDir, outputPath string) error {
	appModulePath, err := detectAppModulePath(projectDir)
	if err != nil {
		return err
	}

	// Preload: pure CJS, runs before ts-node imports any module.
	preload, err := os.CreateTemp(projectDir, ".dg-nestjs-preload-*.js")
	if err != nil {
		return fmt.Errorf("create preload: %w", err)
	}
	defer os.Remove(preload.Name())
	if _, err := preload.WriteString(nestPreloadScript()); err != nil {
		preload.Close()
		return fmt.Errorf("write preload: %w", err)
	}
	preload.Close()

	// Main script: TS generation logic only (no patches).
	tmp, err := os.CreateTemp(projectDir, ".dg-nestjs-swagger-*.ts")
	if err != nil {
		return fmt.Errorf("create temp script: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(buildNestSwaggerScript(appModulePath)); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp script: %w", err)
	}
	tmp.Close()

	if err := runNestScript(projectDir, preload.Name(), tmp.Name(), outputPath); err != nil {
		return fmt.Errorf("nestjs/swagger auto-generation failed")
	}
	return nil
}

// runNestScript runs the NestJS swagger generation script via ts-node, loading
// preloadPath first via -r so DB patches execute before any module is imported.
func runNestScript(projectDir, preloadPath, scriptPath, outputPath string) error {
	args := []string{
		"ts-node",
		"-r", preloadPath,
		"--transpile-only", "--skip-project",
		"--compiler-options", `{"module":"CommonJS","moduleResolution":"node","experimentalDecorators":true,"emitDecoratorMetadata":true,"esModuleInterop":true,"allowSyntheticDefaultImports":true,"target":"ES2021","skipLibCheck":true}`,
	}
	if express.HasTsconfigPaths(projectDir) {
		args = append(args, "-r", "tsconfig-paths/register")
	}
	args = append(args, scriptPath)

	var errBuf strings.Builder
	cmd := exec.Command("npx", args...)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "SWAGGER_OUTPUT="+outputPath)
	cmd.Stdout = express.SubprocessOutput
	cmd.Stderr = io.MultiWriter(express.SubprocessOutput, &errBuf)
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(errBuf.String())
		if detail != "" {
			return fmt.Errorf("run NestJS swagger generator: %w\n\n%s", err, detail)
		}
		return fmt.Errorf("run NestJS swagger generator: %w", err)
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

// nestPreloadScript returns a CJS snippet that patches database drivers to
// no-ops. It is loaded via ts-node -r before the main script so patches run
// before any app module (and therefore any DB driver) is imported.
func nestPreloadScript() string {
	return `// drift-agent preload: patch DB drivers before any module is imported.
try { require('dotenv').config({ quiet: true }); } catch (_) {}

try {
  const t = require('typeorm');
  if (t && t.DataSource) {
    t.DataSource.prototype.initialize = async function() {
      this.isInitialized = true;
      return this;
    };
  }
} catch (_) {}

try {
  const m = require('mongoose');
  if (m && m.Connection) {
    m.Connection.prototype.openUri = async function() { return this; };
  }
} catch (_) {}
`
}

// buildNestSwaggerScript returns a TypeScript snippet that boots the NestJS app,
// generates the OpenAPI document, writes it to SWAGGER_OUTPUT, and exits.
// DB patches are applied via a separate preload file (see nestPreloadScript).
func buildNestSwaggerScript(appModuleAbsPath string) string {
	return fmt.Sprintf(`import { NestFactory } from '@nestjs/core';
import { SwaggerModule, DocumentBuilder } from '@nestjs/swagger';
import * as fs from 'fs';

async function generate(): Promise<void> {
  const { AppModule } = await import('%s');
  const app = await NestFactory.create(AppModule, { abortOnError: false, logger: false });
  const config = new DocumentBuilder()
    .setTitle('API')
    .setVersion('1.0')
    .build();
  const document = SwaggerModule.createDocument(app, config);
  const output = process.env.SWAGGER_OUTPUT ?? 'swagger.json';
  fs.writeFileSync(output, JSON.stringify(document, null, 2));
  process.exit(0);
}

generate().catch((err) => {
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
