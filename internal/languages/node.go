package languages

import "github.com/pgomes13/drift-guard-engine/internal/generate/node/express"

// GenerateNode delegates to the express generate package.
var GenerateNode = express.Node

// GenerateNodeGraphQL delegates to the express GraphQL generate package.
var GenerateNodeGraphQL GeneratorFunc = express.NodeGraphQL
