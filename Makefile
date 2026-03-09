BIN              := drift-guard
MCP_BIN          := drift-guard-mcp
API_BIN          := drift-guard-api
CMD              := ./cmd/drift-guard
MCP_CMD          := ./cmd/mcp-server
API_CMD          := ./cmd/playground
HOMEBREW_TAP     := pgomes13/homebrew-tap
FORMULA          := drift-guard

include make/build.mk
include make/test.mk
include make/run.mk
include make/deploy.mk
include make/git.mk
include make/release.mk
