package ollama

type (
	ModelCapability string

	OllamaModel struct {
		Name         string
		Model        string
		Capabilities []ModelCapability
	}
)

const (
	CapabilityEmbedding   ModelCapability = "embedding"
	CapabilityCompletions ModelCapability = "completion"
	CapabilityVision      ModelCapability = "vision"
	CapabilityAudio       ModelCapability = "audio"
	CapabilityTools       ModelCapability = "tools"
	CapabilityThinking    ModelCapability = "thinking"
)
