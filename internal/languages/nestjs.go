package languages

import "github.com/pgomes13/drift-guard-engine/internal/generate/node/nest"

// GenerateNest delegates to the nest generate package.
var GenerateNest = nest.Nest

// GenerateNestGraphQL delegates to the nest GraphQL generate package.
var GenerateNestGraphQL GeneratorFunc = nest.NestGraphQL
