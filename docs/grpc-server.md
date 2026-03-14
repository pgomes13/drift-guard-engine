# gRPC Server

The engine ships as a standalone gRPC server (`DiffEngine` service, port `50051`) for programmatic use.

## Run with Docker

```sh
docker build -t api-drift-engine .
docker run -p 50051:50051 api-drift-engine
```

Override the port via the `PORT` environment variable:

```sh
docker run -e PORT=9090 -p 9090:9090 api-drift-engine
```

## Proto API

```protobuf
service DiffEngine {
  rpc Diff(DiffRequest) returns (DiffResponse);
}
```

### `DiffRequest` fields

| Field          | Type     | Description                                                                                          |
| -------------- | -------- | ---------------------------------------------------------------------------------------------------- |
| `base_content` | `bytes`  | Raw content of the base schema file                                                                  |
| `head_content` | `bytes`  | Raw content of the head schema file                                                                  |
| `base_name`    | `string` | Original filename (used for extension-based type detection)                                          |
| `head_name`    | `string` | Original filename of the head file                                                                   |
| `type`         | `string` | Explicit schema type: `openapi`, `graphql`, or `grpc`. Auto-detected from `base_name` extension if omitted. |

The proto definition lives at [`api/drift-guard/v1/drift-guard.proto`](https://github.com/pgomes13/api-drift-engine/blob/main/api/drift-guard/v1/drift-guard.proto).
