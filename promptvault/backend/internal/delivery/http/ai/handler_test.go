package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/openrouter"
	authmw "promptvault/internal/middleware/auth"
	aiuc "promptvault/internal/usecases/ai"
)

func testHandlerNoAPIKey(client aiuc.AIClient) *Handler {
	cfg := &config.AIConfig{
		OpenRouterAPIKey: "",
		RateLimitRPM:     100,
		Models: []config.ModelConfig{
			{ID: "test/model", Name: "Test", MaxTokens: 4096},
		},
	}
	svc := aiuc.NewService(client, cfg)
	return NewHandler(svc)
}

// --- helpers ---

type mockStreamClient struct {
	err error
}

func (m *mockStreamClient) Stream(_ context.Context, _ openrouter.ChatRequest, cb openrouter.StreamCallback) (*openrouter.Usage, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, cb("chunk1")
}

func testHandler(client aiuc.AIClient) *Handler {
	cfg := &config.AIConfig{
		OpenRouterAPIKey: "test-key",
		RateLimitRPM:     100,
		Models: []config.ModelConfig{
			{ID: "test/model", Name: "Test", MaxTokens: 4096},
		},
	}
	svc := aiuc.NewService(client, cfg)
	return NewHandler(svc)
}

func testHandlerWithRPM(client aiuc.AIClient, rpm int) *Handler {
	cfg := &config.AIConfig{
		OpenRouterAPIKey: "test-key",
		RateLimitRPM:     rpm,
		Models: []config.ModelConfig{
			{ID: "test/model", Name: "Test", MaxTokens: 4096},
		},
	}
	svc := aiuc.NewService(client, cfg)
	return NewHandler(svc)
}

type flusherRecorder struct {
	*httptest.ResponseRecorder
}

func (f *flusherRecorder) Flush() {}

func withUserID(r *http.Request, userID uint) *http.Request {
	ctx := context.WithValue(r.Context(), authmw.UserIDKey, userID)
	return r.WithContext(ctx)
}

// --- tests ---

func TestModels_Success(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	req := httptest.NewRequest(http.MethodGet, "/api/ai/models", nil)
	rec := httptest.NewRecorder()

	h.Models(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var models []ModelResponse
	err := json.NewDecoder(rec.Body).Decode(&models)
	assert.NoError(t, err)
	assert.Len(t, models, 1)
	assert.Equal(t, "test/model", models[0].ID)
	assert.Equal(t, "Test", models[0].Name)
	assert.Equal(t, 4096, models[0].MaxTokens)
}

func TestEnhance_InvalidJSON(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{bad json`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/enhance", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Enhance(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp["error"])
}

func TestEnhance_MissingFields(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/enhance", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Enhance(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp["error"])
}

func TestEnhance_RateLimited(t *testing.T) {
	h := testHandlerWithRPM(&mockStreamClient{}, 1)

	makeReq := func() *flusherRecorder {
		body := bytes.NewBufferString(`{"content":"hello","model":"test/model"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/ai/enhance", body)
		req.Header.Set("Content-Type", "application/json")
		req = withUserID(req, 42)
		rec := &flusherRecorder{httptest.NewRecorder()}
		h.Enhance(rec, req)
		return rec
	}

	first := makeReq()
	// First request should succeed (SSE starts → 200 implicit)
	assert.NotEqual(t, http.StatusTooManyRequests, first.Code)

	second := makeReq()
	assert.Equal(t, http.StatusTooManyRequests, second.Code)
}

func TestEnhance_SSEHeaders(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/enhance", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := &flusherRecorder{httptest.NewRecorder()}

	h.Enhance(rec, req)

	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", rec.Header().Get("Connection"))
}

// --- Analyze endpoint ---

func TestAnalyze_Success(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/analyze", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := &flusherRecorder{httptest.NewRecorder()}

	h.Analyze(rec, req)

	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "data: chunk1")
	assert.Contains(t, rec.Body.String(), "data: [DONE]")
}

func TestAnalyze_InvalidJSON(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{bad json`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/analyze", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Analyze(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAnalyze_MissingFields(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/analyze", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Analyze(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Variations endpoint ---

func TestVariations_Success(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model","count":3}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/variations", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := &flusherRecorder{httptest.NewRecorder()}

	h.Variations(rec, req)

	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "data: chunk1")
	assert.Contains(t, rec.Body.String(), "data: [DONE]")
}

func TestVariations_InvalidCount(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model","count":6}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/variations", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Variations(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestVariations_MissingFields(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/variations", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Variations(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Rewrite endpoint (extended) ---

func TestRewrite_Success(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model","style":"formal"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/rewrite", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := &flusherRecorder{httptest.NewRecorder()}

	h.Rewrite(rec, req)

	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "data: chunk1")
	assert.Contains(t, rec.Body.String(), "data: [DONE]")
}

func TestRewrite_AllStyles(t *testing.T) {
	styles := []string{"formal", "concise", "creative", "detailed", "technical"}
	for _, style := range styles {
		t.Run(style, func(t *testing.T) {
			h := testHandler(&mockStreamClient{})

			body := bytes.NewBufferString(`{"content":"hello","model":"test/model","style":"` + style + `"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/ai/rewrite", body)
			req.Header.Set("Content-Type", "application/json")
			req = withUserID(req, 1)
			rec := &flusherRecorder{httptest.NewRecorder()}

			h.Rewrite(rec, req)

			assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"),
				"style %q should produce SSE response", style)
		})
	}
}

func TestRewrite_MissingStyle(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/rewrite", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Rewrite(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Service error handling ---

func TestEnhance_ServiceError(t *testing.T) {
	h := testHandler(&mockStreamClient{err: openrouter.ErrEmptyResponse})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/enhance", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := &flusherRecorder{httptest.NewRecorder()}

	h.Enhance(rec, req)

	assert.Contains(t, rec.Body.String(), "event: error")
	assert.Contains(t, rec.Body.String(), "Модель вернула пустой ответ")
}

func TestEnhance_APIKeyMissing(t *testing.T) {
	// ErrAPIKeyMissing occurs inside service.validate(), after SSE headers are sent.
	// So it appears as an SSE error event, not as HTTP 503.
	h := testHandlerNoAPIKey(&mockStreamClient{})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/enhance", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := &flusherRecorder{httptest.NewRecorder()}

	h.Enhance(rec, req)

	assert.Contains(t, rec.Body.String(), "event: error")
	assert.Contains(t, rec.Body.String(), "Ошибка AI-сервиса")
}

// --- SSE body verification ---

func TestEnhance_SSEBody(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/enhance", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := &flusherRecorder{httptest.NewRecorder()}

	h.Enhance(rec, req)

	responseBody := rec.Body.String()
	assert.Contains(t, responseBody, "data: chunk1\n")
	assert.Contains(t, responseBody, "data: [DONE]\n")
}

func TestEnhance_SSEErrorEvent(t *testing.T) {
	h := testHandler(&mockStreamClient{err: openrouter.ErrUnauthorized})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/enhance", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := &flusherRecorder{httptest.NewRecorder()}

	h.Enhance(rec, req)

	responseBody := rec.Body.String()
	assert.Contains(t, responseBody, "event: error\n")
	assert.Contains(t, responseBody, "AI-сервис временно недоступен")
	assert.NotContains(t, responseBody, "data: [DONE]")
}

// --- userFriendlyError mapping ---

func TestUserFriendlyError_Mapping(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"Unauthorized", openrouter.ErrUnauthorized, "AI-сервис временно недоступен"},
		{"RateLimited", openrouter.ErrRateLimited, "Превышен лимит запросов к AI"},
		{"InsufficientCredits", openrouter.ErrInsufficientCredits, "AI-сервис временно недоступен"},
		{"EmptyResponse", openrouter.ErrEmptyResponse, "Модель вернула пустой ответ"},
		{"UnknownError", errors.New("unknown"), "Ошибка AI-сервиса. Попробуйте позже"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := testHandler(&mockStreamClient{err: tc.err})

			body := bytes.NewBufferString(`{"content":"hello","model":"test/model"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/ai/enhance", body)
			req.Header.Set("Content-Type", "application/json")
			req = withUserID(req, 1)
			rec := &flusherRecorder{httptest.NewRecorder()}

			h.Enhance(rec, req)

			assert.Contains(t, rec.Body.String(), tc.expected)
		})
	}
}

// --- existing tests ---

func TestRewrite_InvalidStyle(t *testing.T) {
	h := testHandler(&mockStreamClient{})

	body := bytes.NewBufferString(`{"content":"hello","model":"test/model","style":"invalid"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/rewrite", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.Rewrite(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp["error"])
}
