package languages

import "github.com/DriftAgent/api-drift-engine/internal/generate/golang"

// GenerateGo delegates to the golang generate package.
var GenerateGo = golang.Go

// GenerateGoGraphQL delegates to the golang GraphQL generator.
var GenerateGoGraphQL GeneratorFunc = golang.GoGraphQL

// GenerateGoGRPC delegates to the golang gRPC generator.
var GenerateGoGRPC GeneratorFunc = golang.GoGRPC
