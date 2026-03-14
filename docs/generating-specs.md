# Generating Specs

The API Drift Agent auto-detects schema files in your repo. If it can't find one (you'll see "No OpenAPI schema found" in the action logs), you need to generate and commit a schema file first.

Below are the most common tools for each language and schema type.

## OpenAPI

### Node.js

| Tool | Framework | Command | Output |
|------|-----------|---------|--------|
| [`@nestjs/swagger`](https://docs.nestjs.com/openapi/introduction) | NestJS | `nest build` + export via `SwaggerModule` | `openapi.json` |
| [`swagger-jsdoc`](https://github.com/Surnet/swagger-jsdoc) | Express / Fastify | `npx swagger-jsdoc -d swaggerDef.js routes/*.js -o openapi.json` | `openapi.json` |
| [`fastify-swagger`](https://github.com/fastify/fastify-swagger) | Fastify | `app.ready(() => writeFileSync('openapi.json', JSON.stringify(app.swagger())))` | `openapi.json` |
| [`tsoa`](https://tsoa-community.github.io/docs/) | Express / Koa | `npx tsoa spec` | `build/swagger.json` |

### Go

| Tool | Framework | Command | Output |
|------|-----------|---------|--------|
| [`swag`](https://github.com/swaggo/swag) | Gin / Echo / Fiber | `swag init -g main.go -o docs` | `docs/swagger.yaml` |
| [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) | net/http / Chi | schema-first — write `.yaml`, generate code | hand-authored |
| [`huma`](https://huma.rocks/) | net/http / Chi | schema auto-generated at startup; dump via `/openapi.json` | served at runtime |

---

## GraphQL

### Node.js

| Tool | Framework | Command | Output |
|------|-----------|---------|--------|
| [`graphql-inspector`](https://the-guild.dev/graphql/inspector) | Apollo / Yoga | `npx graphql-inspector introspect http://localhost:4000 --write schema.graphql` | `schema.graphql` |
| [`get-graphql-schema`](https://github.com/prisma-labs/get-graphql-schema) | Any | `npx get-graphql-schema http://localhost:4000 > schema.graphql` | `schema.graphql` |
| [`@graphql-codegen`](https://the-guild.dev/graphql/codegen) | Apollo / Yoga | `npx graphql-codegen --config codegen.ts` | configurable |
| Apollo Server | Apollo | `rover graph introspect http://localhost:4000 > schema.graphql` | `schema.graphql` |

### Go

| Tool | Framework | Command | Output |
|------|-----------|---------|--------|
| [`gqlgen`](https://gqlgen.com/) | gqlgen | schema-first — `graph/schema.graphqls` is your source | hand-authored |
| [`graph-gophers/graphql-go`](https://github.com/graph-gophers/graphql-go) | net/http | schema defined as Go string literal; export to file | hand-authored |
| Introspection | Any server | `curl -X POST http://localhost:8080/graphql -d '{"query":"{__schema{...}}"}' \| jq > schema.json` | `schema.json` |

---

## gRPC / Protobuf

### Node.js

| Tool | Framework | Command | Output |
|------|-----------|---------|--------|
| [`@grpc/proto-loader`](https://github.com/grpc/grpc-node/tree/master/packages/proto-loader) | grpc-js | `.proto` files are the source of truth — check them in directly | hand-authored |
| [`ts-proto`](https://github.com/stephenh/ts-proto) | grpc-js | `protoc --plugin=protoc-gen-ts_proto --ts_proto_out=. service.proto` | generated TS; `.proto` is source |
| [`buf`](https://buf.build/) | Any | `buf build --output image.bin` (or use `.proto` files directly) | `.proto` / BSR |

### Go

| Tool | Framework | Command | Output |
|------|-----------|---------|--------|
| [`protoc`](https://grpc.io/docs/languages/go/quickstart/) | google.golang.org/grpc | `.proto` files are the source of truth — check them in directly | hand-authored |
| [`buf`](https://buf.build/) | Any | `buf generate` | generated Go; `.proto` is source |
| [`grpc-gateway`](https://grpc-ecosystem.github.io/grpc-gateway/) | grpc-gateway | annotate `.proto`, generate OpenAPI via `protoc-gen-openapiv2` | `apidocs.swagger.json` |

---

## Quick reference

| Schema type | Minimum drift-agent input |
|-------------|--------------------------|
| OpenAPI | Any `.yaml` or `.json` file valid against OpenAPI 3.x |
| GraphQL | SDL file (`.graphql` / `.gql`) or introspection JSON |
| gRPC | A `.proto` file using `syntax = "proto3"` |
