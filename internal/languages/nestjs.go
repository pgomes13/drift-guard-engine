package languages

import "github.com/pgomes13/api-drift-engine/internal/generate/node/nest"

// GenerateNest delegates to the nest generate package.
var GenerateNest = nest.Nest

// GenerateNestGraphQL delegates to the nest GraphQL generate package.
var GenerateNestGraphQL GeneratorFunc = nest.NestGraphQL

// GenerateNestGRPC delegates to the nest gRPC generate package.
var GenerateNestGRPC GeneratorFunc = nest.NestGRPC
