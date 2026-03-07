# Supported Languages & Frameworks

## Schema formats

`drift-guard` can diff any two schema files of these types:

| Format | Command | File types |
|--------|---------|------------|
| OpenAPI 3.x | `openapi` | `.yaml`, `.json` |
| GraphQL SDL | `graphql` | `.graphql`, `.gql` |
| Protobuf | `grpc` | `.proto` |

## Auto-detection (`compare`)

`drift-guard compare` auto-detects your project type and generates schemas automatically.

### Node.js

| Framework | REST (OpenAPI) | GraphQL | gRPC |
|-----------|---------------|---------|------|
| Express | Yes | Yes | Yes |
| NestJS | Yes | Yes | Yes |

### Go

| Framework | REST (OpenAPI) |
|-----------|---------------|
| Any (swag annotations) | Yes |

