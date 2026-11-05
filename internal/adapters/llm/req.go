package llm

type (
	chatRequest struct {
		Model       string        `json:"model"`
		Messages    []chatMessage `json:"messages"`
		Stream      bool          `json:"stream"`
		Temperature *float64      `json:"temperature,omitempty"`
		MaxTokens   *int          `json:"max_tokens,omitempty"`
	}

	chatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	chatResponse struct {
		ID      string       `json:"id"`
		Model   string       `json:"model"`
		Choices []chatChoice `json:"choices"`
		Usage   chatUsage    `json:"usage"`
	}

	chatChoice struct {
		Index        int         `json:"index"`
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	}

	chatUsage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	}
)
