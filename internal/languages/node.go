package languages

import "github.com/DriftAgent/api-drift-engine/internal/generate/node/express"

// GenerateNode delegates to the express generate package.
var GenerateNode = express.Node

// GenerateNodeGraphQL delegates to the express GraphQL generate package.
var GenerateNodeGraphQL GeneratorFunc = express.NodeGraphQL

// GenerateNodeGRPC delegates to the express gRPC generate package.
var GenerateNodeGRPC GeneratorFunc = express.NodeGRPC
