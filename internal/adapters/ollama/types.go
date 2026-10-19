package ollama

type (
	ModelCapability string

	OllamaModel struct {
		Name         string            `json:"name"`
		Model        string            `json:"model"`
		Capabilities []ModelCapability `json:"capabilities"`
	}

	OllamaModelsResponse struct {
		Models []OllamaModel `json:"models"`
	}

	OllamaEmbedRequest struct {
		Model      string   `json:"model"`
		Input      []string `json:"input"`
		Dimensions uint32   `json:"dimensions"`
	}

	OllamaEmbedResponse struct {
		Model      string      `json:"model"`
		Embeddings [][]float32 `json:"embeddings"`
	}

	OllamaGenerateOptions struct {
		NumPredict uint32 `json:"num_predict"`
	}

	OllamaSummarizationRequest struct {
		Model   string                `json:"model"`
		Prompt  string                `json:"prompt"`
		System  string                `json:"system"`
		Stream  bool                  `json:"stream"`
		Think   bool                  `json:"think"`
		Options OllamaGenerateOptions `json:"options"`
	}

	OllamaSummarizationResponse struct {
		Model    string `json:"model"`
		Response string `json:"response"`
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
