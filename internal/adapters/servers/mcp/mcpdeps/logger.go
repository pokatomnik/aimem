package mcpdeps

import "github.com/pokatomnik/gomem/internal/shared/logging"

// Logger is the logging seam of the MCP adapter.
//
// It aliases the shared interface so the mcp and tools packages keep their local
// spelling while every layer depends on the same contract.
type Logger = logging.Logger
