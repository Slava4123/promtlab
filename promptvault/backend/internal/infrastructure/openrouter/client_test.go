package openrouter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sseServer(t *testing.T, lines ...string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for _, line := range lines {
			fmt.Fprintln(w, line)
		}
	}))
}

func errorServer(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

func defaultRequest() ChatRequest {
	return ChatRequest{
		Model:    "test",
		Messages: []Message{{Role: "user", Content: "hi"}},
	}
}

// TestStream_ValidChunks verifies that the callback is called once per content
// chunk and receives the correct text.
func TestStream_ValidChunks(t *testing.T) {
	srv := sseServer(t,
		`data: {"choices":[{"delta":{"content":"Hello"}}]}`,
		"",
		`data: {"choices":[{"delta":{"content":" world"}}]}`,
		"",
		"data: [DONE]",
		"",
	)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	var chunks []string
	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"Hello", " world"}, chunks)
}

// TestStream_DoneSignal verifies that [DONE] terminates the stream without error.
func TestStream_DoneSignal(t *testing.T) {
	srv := sseServer(t,
		`data: {"choices":[{"delta":{"content":"ok"}}]}`,
		"",
		"data: [DONE]",
		"",
	)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	callCount := 0
	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		callCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

// TestStream_EmptyResponse verifies that a stream with only [DONE] and no
// content chunks returns ErrEmptyResponse.
func TestStream_EmptyResponse(t *testing.T) {
	srv := sseServer(t,
		"data: [DONE]",
		"",
	)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		t.Fatal("callback should not be called")
		return nil
	})

	assert.True(t, errors.Is(err, ErrEmptyResponse))
}

// TestStream_CallbackError verifies that when the callback returns an error
// the stream stops and that error is propagated.
func TestStream_CallbackError(t *testing.T) {
	srv := sseServer(t,
		`data: {"choices":[{"delta":{"content":"first"}}]}`,
		"",
		`data: {"choices":[{"delta":{"content":"second"}}]}`,
		"",
		"data: [DONE]",
		"",
	)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	cbErr := errors.New("callback failed")
	callCount := 0
	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		callCount++
		return cbErr
	})

	assert.ErrorIs(t, err, cbErr)
	assert.Equal(t, 1, callCount)
}

// TestStream_MalformedJSON verifies that a malformed JSON chunk is skipped and
// subsequent valid chunks are still processed.
func TestStream_MalformedJSON(t *testing.T) {
	srv := sseServer(t,
		`data: {not json at all`,
		"",
		`data: {"choices":[{"delta":{"content":"valid"}}]}`,
		"",
		"data: [DONE]",
		"",
	)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	var chunks []string
	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"valid"}, chunks)
}

// TestStream_EmptyChoices verifies that a chunk with an empty choices array is
// skipped and does not invoke the callback.
func TestStream_EmptyChoices(t *testing.T) {
	srv := sseServer(t,
		`data: {"choices":[]}`,
		"",
		`data: {"choices":[{"delta":{"content":"after"}}]}`,
		"",
		"data: [DONE]",
		"",
	)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	var chunks []string
	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"after"}, chunks)
}

// TestHandleError_401 verifies that a 401 response returns ErrUnauthorized.
func TestHandleError_401(t *testing.T) {
	srv := errorServer(t, http.StatusUnauthorized)
	defer srv.Close()

	client := NewClientWithBaseURL("bad-key", srv.URL)

	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		t.Fatal("callback should not be called")
		return nil
	})

	assert.True(t, errors.Is(err, ErrUnauthorized))
}

// TestHandleError_429 verifies that a 429 response returns ErrRateLimited.
func TestHandleError_429(t *testing.T) {
	srv := errorServer(t, http.StatusTooManyRequests)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		t.Fatal("callback should not be called")
		return nil
	})

	assert.True(t, errors.Is(err, ErrRateLimited))
}

// TestHandleError_402 verifies that a 402 response returns ErrInsufficientCredits.
func TestHandleError_402(t *testing.T) {
	srv := errorServer(t, http.StatusPaymentRequired)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		t.Fatal("callback should not be called")
		return nil
	})

	assert.True(t, errors.Is(err, ErrInsufficientCredits))
}

// TestStream_MultilineChunk verifies that a chunk containing a newline
// is delivered to the callback as a single string.
func TestStream_MultilineChunk(t *testing.T) {
	srv := sseServer(t,
		`data: {"choices":[{"delta":{"content":"line1\nline2"}}]}`,
		"",
		"data: [DONE]",
		"",
	)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	var chunks []string
	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"line1\nline2"}, chunks)
}

// TestStream_ContextCanceled verifies that canceling the context stops the stream.
func TestStream_ContextCanceled(t *testing.T) {
	// Create a server that sends one chunk, then blocks forever
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"choices":[{"delta":{"content":"hello"}}]}`)
		fmt.Fprintln(w, "")
		w.(http.Flusher).Flush()
		// Block until client cancels
		<-r.Context().Done()
	}))
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	called := make(chan struct{}, 1)
	go func() {
		<-called
		cancel()
	}()

	callCount := 0
	_, err := client.Stream(ctx, defaultRequest(), func(chunk string) error {
		callCount++
		select {
		case called <- struct{}{}:
		default:
		}
		return nil
	})

	assert.Error(t, err)
	assert.Equal(t, 1, callCount)
}

// TestHandleError_500 verifies that a 500 response returns a generic error.
func TestHandleError_500(t *testing.T) {
	srv := errorServer(t, http.StatusInternalServerError)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		t.Fatal("callback should not be called")
		return nil
	})

	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrUnauthorized))
	assert.False(t, errors.Is(err, ErrRateLimited))
	assert.False(t, errors.Is(err, ErrInsufficientCredits))
	assert.False(t, errors.Is(err, ErrModelNotFound))
	assert.Contains(t, err.Error(), "500")
}

// TestHandleError_404 verifies that a 404 response returns ErrModelNotFound.
func TestHandleError_404(t *testing.T) {
	srv := errorServer(t, http.StatusNotFound)
	defer srv.Close()

	client := NewClientWithBaseURL("test-key", srv.URL)

	_, err := client.Stream(context.Background(), defaultRequest(), func(chunk string) error {
		t.Fatal("callback should not be called")
		return nil
	})

	assert.True(t, errors.Is(err, ErrModelNotFound))
}
