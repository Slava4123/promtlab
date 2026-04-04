package openrouter

// StreamCallback is called for each content chunk from OpenRouter.
// Return a non-nil error to abort the stream.
type StreamCallback func(chunk string) error

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model        string        `json:"model"`
	Messages     []Message     `json:"messages"`
	MaxTokens    int           `json:"max_tokens,omitempty"`
	Temperature  *float64      `json:"temperature,omitempty"`
	Reasoning    *Reasoning    `json:"reasoning,omitempty"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

// Reasoning configures reasoning effort for models that support reasoning (via OpenRouter).
// OpenRouter format: {"reasoning": {"effort": "medium"}}
type Reasoning struct {
	Effort string `json:"effort"`
}

// chatRequestWire is the wire format sent to OpenRouter (adds stream field).
type chatRequestWire struct {
	ChatRequest
	Stream bool `json:"stream"`
}

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CacheControl enables prompt caching for providers that require explicit opt-in (Anthropic).
type CacheControl struct {
	Type string `json:"type"`
}

// Usage contains token and cost data from the final SSE chunk.
type Usage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	Cost             float64 `json:"cost"`
	CachedTokens     int     `json:"cached_tokens"`
}

// sseChunk matches the OpenAI-compatible SSE chunk format from OpenRouter.
type sseChunk struct {
	Choices []sseChoice `json:"choices"`
	Usage   *sseUsage   `json:"usage,omitempty"`
}

type sseChoice struct {
	Delta        sseDelta `json:"delta"`
	FinishReason *string  `json:"finish_reason"`
}

type sseDelta struct {
	Content string `json:"content"`
}

type sseUsage struct {
	PromptTokens     int                `json:"prompt_tokens"`
	CompletionTokens int                `json:"completion_tokens"`
	TotalTokens      int                `json:"total_tokens"`
	Cost             float64            `json:"cost"`
	PromptDetails    *ssePromptDetails  `json:"prompt_tokens_details,omitempty"`
}

type ssePromptDetails struct {
	CachedTokens     int `json:"cached_tokens"`
	CacheWriteTokens int `json:"cache_write_tokens"`
}

// apiError represents an error response from OpenRouter.
type apiError struct {
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}
