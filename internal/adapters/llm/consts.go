package llm

import _ "embed"

const (
	temperature = 0.0
)

//go:embed SUMMARIZATION_SYSTEM.md
var systemPrompt string
