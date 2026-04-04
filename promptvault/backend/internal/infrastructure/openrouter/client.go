package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://openrouter.ai/api/v1"

// Client is an HTTP client for the OpenRouter API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new OpenRouter client with the given API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// NewClientWithBaseURL creates a client with a custom base URL (for testing).
func NewClientWithBaseURL(apiKey, baseURL string) *Client {
	c := NewClient(apiKey)
	c.baseURL = baseURL
	return c
}

// Stream sends a chat completion request with streaming enabled and calls cb
// for each content delta. It blocks until the stream completes, the context
// is cancelled, or cb returns an error. Returns usage data from the final chunk.
func (c *Client) Stream(ctx context.Context, req ChatRequest, cb StreamCallback) (*Usage, error) {
	wire := chatRequestWire{ChatRequest: req, Stream: true}

	body, err := json.Marshal(wire)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://promptvault.app")
	httpReq.Header.Set("X-Title", "PromptVault")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	return c.readSSEStream(resp.Body, cb)
}

func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusPaymentRequired:
		return ErrInsufficientCredits
	case http.StatusNotFound:
		return ErrModelNotFound
	default:
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("OpenRouter API error %d: %s", resp.StatusCode, apiErr.Error.Message)
		}
		return fmt.Errorf("OpenRouter API error %d: %s", resp.StatusCode, string(body))
	}
}

func (c *Client) readSSEStream(body io.Reader, cb StreamCallback) (*Usage, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	gotContent := false
	var usage *Usage

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			if !gotContent {
				return nil, ErrEmptyResponse
			}
			return usage, nil
		}

		var chunk sseChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			slog.Warn("malformed SSE chunk from OpenRouter", "data", data, "error", err)
			continue
		}

		// Parse usage from the final chunk (where finish_reason is set).
		if chunk.Usage != nil {
			usage = &Usage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
				Cost:             chunk.Usage.Cost,
			}
			if chunk.Usage.PromptDetails != nil {
				usage.CachedTokens = chunk.Usage.PromptDetails.CachedTokens
			}
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		content := chunk.Choices[0].Delta.Content
		if content == "" {
			continue
		}

		gotContent = true
		if err := cb(content); err != nil {
			return usage, err
		}
	}

	if err := scanner.Err(); err != nil {
		return usage, fmt.Errorf("read stream: %w", err)
	}

	if !gotContent {
		return nil, ErrEmptyResponse
	}

	return usage, nil
}
