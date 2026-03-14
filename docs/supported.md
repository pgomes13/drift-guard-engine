# Supported

## Schema formats

`driftabot` can diff any two schema files of these types:

| Format      | Command   | File types         |
| ----------- | --------- | ------------------ |
| OpenAPI 3.x | `openapi` | `.yaml`, `.json`   |
| GraphQL SDL | `graphql` | `.graphql`, `.gql` |
| Protobuf    | `grpc`    | `.proto`           |

## Auto-detection (`compare`)

`driftabot compare` auto-detects your project type and generates schemas automatically.

### Node.js

| Framework | REST (OpenAPI) | GraphQL | gRPC |
| --------- | -------------- | ------- | ---- |
| Express   | Yes            | Yes     | Yes  |
| NestJS    | Yes            | Yes     | Yes  |

### Go

Detected by `go.mod`, legacy manifests (`Gopkg.toml`, `glide.yaml`), or `.go` files in the project root. The specific framework (Gin, Echo, Fiber, Chi, Gorilla Mux) is identified from the module graph and shown at startup.

| Framework   | REST (OpenAPI)                                                       | GraphQL | gRPC |
| ----------- | -------------------------------------------------------------------- | ------- | ---- |
| Gin         | Yes (requires [`swag`](https://github.com/swaggo/swag) annotations) | Yes     | Yes  |
| Echo        | Yes                                                                  | Yes     | Yes  |
| Fiber       | Yes                                                                  | Yes     | Yes  |
| Chi         | Yes                                                                  | Yes     | Yes  |
| Gorilla Mux | Yes                                                                  | Yes     | Yes  |
| plain Go    | Yes                                                                  | Yes     | Yes  |

> More languages and frameworks coming soon.
