package nest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DriftAgent/api-drift-engine/internal/generate/node/express"
)

// NestGraphQL finds or generates the GraphQL SDL schema for the NestJS project
// and copies it to outputDir/schema.graphql.
//
// Strategy (in order):
//  1. Schema-first: look for an existing schema.gql/schema.graphql and copy it.
//  2. Code-first: generate via a temp script that boots the app and captures
//     the schema emitted by @nestjs/graphql with autoSchemaFile configured.
//  3. Error: guide the user.
func NestGraphQL(projectDir, outputDir string) error {
	outputPath := filepath.Join(outputDir, "schema.graphql")

	// 1. Existing schema file (schema-first or already-generated code-first).
	if src := express.FindGraphQLSchema(projectDir); src != "" {
		return nestCopyFile(src, outputPath)
	}

	// 2. Code-first: generate via @nestjs/graphql auto-schema.
	if nestHasGraphQLDep(projectDir) {
		if err := nestGenerateGraphQL(projectDir, outputPath); err == nil {
			return nil
		}
		// Fall through to the actionable error on failure.
	}

	return fmt.Errorf(
		"no GraphQL schema found in %s\n\n"+
			"Schema-first: commit your schema file at schema.gql or src/schema.graphql.\n\n"+
			"Code-first: configure autoSchemaFile in GraphQLModule and run your app once\n"+
			"to generate schema.gql, then commit it so drift-guard can compare across branches.",
		projectDir,
	)
}

// nestHasGraphQLDep reports whether @nestjs/graphql is listed in package.json.
func nestHasGraphQLDep(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), `"@nestjs/graphql"`)
}

// nestGenerateGraphQL boots the NestJS app via a temp TypeScript script,
// waits for @nestjs/graphql to emit schema.gql, then copies it to outputPath.
func nestGenerateGraphQL(projectDir, outputPath string) error {
	appModulePath, err := detectAppModulePath(projectDir)
	if err != nil {
		return err
	}

	script := buildNestGraphQLScript(appModulePath)

	tmp, err := os.CreateTemp(projectDir, ".dg-nestjs-graphql-*.ts")
	if err != nil {
		return fmt.Errorf("create temp script: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(script); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp script: %w", err)
	}
	tmp.Close()

	if err := express.RunScript(projectDir, tmp.Name(), outputPath); err != nil {
		return fmt.Errorf("nestjs/graphql schema generation failed")
	}
	return nil
}

// buildNestGraphQLScript returns a TypeScript snippet that boots the NestJS app
// (triggering @nestjs/graphql to emit schema.gql) and copies it to GRAPHQL_OUTPUT.
func buildNestGraphQLScript(appModuleAbsPath string) string {
	return fmt.Sprintf(`try { require('dotenv').config({ quiet: true }); } catch (_) {}
import { NestFactory } from '@nestjs/core';
import * as fs from 'fs';

const deadline = setTimeout(() => {
  process.stderr.write(
    '\ndrift-guard: NestJS app did not finish initialising within 15 s.\n' +
    'Ensure your services are running or use schema-first GraphQL.\n',
  );
  process.exit(1);
}, 15_000);
deadline.unref();

async function generate(): Promise<void> {
  const { AppModule } = await import('%s');
  const app = await NestFactory.create(AppModule, { abortOnError: false });
  clearTimeout(deadline);

  // @nestjs/graphql writes schema.gql when autoSchemaFile is configured.
  const schemaLocations = ['schema.gql', 'src/schema.gql', 'schema.graphql'];
  for (const loc of schemaLocations) {
    if (fs.existsSync(loc)) {
      const output = process.env.GRAPHQL_OUTPUT ?? 'schema.graphql';
      fs.copyFileSync(loc, output);
      await app.close();
      process.exit(0);
    }
  }

  process.stderr.write('drift-guard: schema.gql was not generated after app boot.\n');
  process.exit(1);
}

generate().catch((err) => {
  clearTimeout(deadline);
  process.stderr.write('Error: ' + String(err?.message ?? err) + '\n');
  process.exit(1);
});
`, appModuleAbsPath)
}

func nestCopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read GraphQL schema %s: %w", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
