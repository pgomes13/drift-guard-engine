package nest

import (
	"fmt"
	"os"
	"path/filepath"
)

// ScaffoldNestSwaggerScript writes a starter drift-agent/scripts/generate-swagger.ts
// to projectDir (creating the directory if needed). It returns the path of the
// file that was written.
func ScaffoldNestSwaggerScript(projectDir string) (string, error) {
	outPath := filepath.Join(projectDir, "drift-agent", "scripts", "generate-swagger.ts")

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", fmt.Errorf("create scripts directory: %w", err)
	}

	appModuleRelPath := detectAppModuleRelPath(projectDir)
	content := buildNestSwaggerScaffold(appModuleRelPath)

	if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write scaffold: %w", err)
	}
	return outPath, nil
}

// detectAppModuleRelPath returns a relative import path for the AppModule,
// suitable for use inside scripts/generate-swagger.ts.
func detectAppModuleRelPath(projectDir string) string {
	// Script lives at drift-agent/scripts/generate-swagger.ts, so imports
	// need two levels up (../../) to reach the project root.
	candidates := []struct {
		rel     string
		import_ string
	}{
		{"src/app.module.ts", "../../src/app.module"},
		{"src/app.module.js", "../../src/app.module"},
		{"app.module.ts", "../../app.module"},
		{"app.module.js", "../../app.module"},
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(projectDir, c.rel)); err == nil {
			return c.import_
		}
	}
	return "../../src/app.module"
}

// buildNestSwaggerScaffold returns the TypeScript source for the scaffold.
func buildNestSwaggerScaffold(appModuleRelPath string) string {
	return fmt.Sprintf(`/**
 * scripts/generate-swagger.ts
 *
 * Generates an OpenAPI (swagger) document for your NestJS application and
 * writes it to the path specified by the SWAGGER_OUTPUT environment variable.
 *
 * Run via drift-agent:
 *   drift-agent generate
 *
 * Or directly:
 *   SWAGGER_OUTPUT=swagger.json npx ts-node --transpile-only scripts/generate-swagger.ts
 *
 * -----------------------------------------------------------------------
 * If your app requires a live database or other services to start, you
 * have two options:
 *
 * Option A — start the real services, then run this script.
 *
 * Option B — override the heavy providers with no-op mocks so the app can
 * boot without infrastructure. Example using @nestjs/testing:
 *
 *   import { Test } from '@nestjs/testing';
 *   import { TypeOrmModule } from '@nestjs/typeorm';
 *   import { getDataSourceToken } from '@nestjs/typeorm';
 *
 *   const moduleRef = await Test.createTestingModule({ imports: [AppModule] })
 *     .overrideProvider(getDataSourceToken())
 *     .useValue({ isInitialized: true })
 *     .compile();
 *   const app = moduleRef.createNestApplication();
 *   await app.init();
 * -----------------------------------------------------------------------
 */

// Load .env into process.env using only Node builtins (no dotenv package needed).
// This ensures TypeORM / ConfigModule can read DB credentials before app init.
(function loadEnv() {
  const fs = require('fs');
  const path = require('path');
  try {
    const lines = fs.readFileSync(path.join(process.cwd(), '.env'), 'utf8').split('\n');
    for (const line of lines) {
      const m = line.match(/^\s*([^#\s][^=]*?)\s*=\s*(.*?)\s*$/);
      if (m && process.env[m[1]] === undefined) {
        process.env[m[1]] = m[2].replace(/^(['"])(.*)\1$/, '$2');
      }
    }
  } catch (_) {}
})();

import { NestFactory } from '@nestjs/core';
import { SwaggerModule, DocumentBuilder } from '@nestjs/swagger';
import * as fs from 'fs';
import { AppModule } from '%s';

async function generate(): Promise<void> {
  // abortOnError: false makes NestJS throw on failure instead of silently
  // calling process.exit(1), so errors are visible in the output below.
  const app = await NestFactory.create(AppModule, { abortOnError: false });

  const config = new DocumentBuilder()
    .setTitle('API')
    .setVersion('1.0')
    .build();

  const document = SwaggerModule.createDocument(app, config);

  const output = process.env.SWAGGER_OUTPUT ?? 'swagger.json';
  fs.writeFileSync(output, JSON.stringify(document, null, 2));

  // Force-exit so open handles (DB pools, queues) don't block the process.
  process.exit(0);
}

generate().catch((err) => {
  console.error(err);
  process.exit(1);
});
`, appModuleRelPath)
}
