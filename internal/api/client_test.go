package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	ctx := context.Background()
	token := "test-token-123"

	client := NewClient(ctx, token)

	assert.NotNil(t, client)
	assert.Equal(t, BaseURL, client.baseURL)
	assert.Equal(t, token, client.token)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, Timeout, client.httpClient.Timeout)
}

func TestClientSetRequestHeaders(t *testing.T) {
	ctx := context.Background()
	client := NewClient(ctx, "test-token")

	req, err := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	require.NoError(t, err)

	client.setRequestHeaders(req)

	assert.Equal(t, DefaultUserAgent, req.Header.Get("User-Agent"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
}

func TestClientGet(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "test"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token:      "test-token",
		maxRetries: MaxRetries,
	}

	var result map[string]interface{}
	err := client.Get(context.Background(), "/test", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, "test", result["data"])
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token:      "test-token",
		maxRetries: MaxRetries,
	}

	var result map[string]interface{}
	err := client.Post(context.Background(), "/test", nil, &result)

	require.NoError(t, err)
	assert.Equal(t, float64(1), result["id"])
}

func TestClientPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"updated": true}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token:      "test-token",
		maxRetries: MaxRetries,
	}

	var result map[string]interface{}
	err := client.Put(context.Background(), "/test", nil, &result)

	require.NoError(t, err)
	updated, ok := result["updated"].(bool)
	assert.True(t, ok && updated)
}

func TestClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"deleted": true}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token:      "test-token",
		maxRetries: MaxRetries,
	}

	var result map[string]interface{}
	err := client.Delete(context.Background(), "/test", &result)

	require.NoError(t, err)
	deleted, ok := result["deleted"].(bool)
	assert.True(t, ok && deleted)
}

func TestClientAuthenticationErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token:      "invalid-token",
		maxRetries: MaxRetries,
	}

	var result map[string]interface{}
	err := client.Get(context.Background(), "/test", nil, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestClientForbiddenErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "forbidden"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token:      "test-token",
		maxRetries: MaxRetries,
	}

	var result map[string]interface{}
	err := client.Get(context.Background(), "/test", nil, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestClientForbiddenErrorSpecific(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "forbidden"}`))
	}))
	defer server.Close()

	client := NewClientWithOptions(context.Background(), "test-token", server.URL, Timeout, MaxRetries, "test")

	err := client.Get(context.Background(), "/test", nil, nil)
	if err == nil {
		t.Fatal("Expected error for 403, got nil")
	}

	// The error message should contain "forbidden"
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("Expected error to contain 'forbidden', got: %s", err.Error())
	}
}

func TestClientTypeSafety(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": 1, "company_name": "Test"}`))
	}))
	defer server.Close()

	client := NewClientWithOptions(context.Background(), "test-token", server.URL, Timeout, MaxRetries, "test")

	var result map[string]interface{}
	err := client.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Expected no error, got: %s", err)
	}

	id, ok := result["id"].(float64)
	if !ok {
		t.Error("Expected id to be float64")
	} else if id != 1 {
		t.Errorf("Expected id to be 1, got %f", id)
	}

	name, ok := result["company_name"].(string)
	if !ok {
		t.Error("Expected company_name to be string")
	} else if name != "Test" {
		t.Errorf("Expected company_name to be 'Test', got '%s'", name)
	}
}

func TestClientNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	client := NewClientWithOptions(context.Background(), "test-token", server.URL, Timeout, MaxRetries, "test")

	var result map[string]interface{}
	err := client.Get(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Fatal("Expected error for 404, got nil")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got: %s", err)
	}
}

func TestClientDeleteWithNilTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	client := NewClientWithOptions(context.Background(), "test-token", server.URL, Timeout, MaxRetries, "test")

	// Delete with nil target should not error
	err := client.Delete(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Expected no error when deleting with nil target, got: %s", err)
	}
}

func TestClientDeleteWithEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty body
	}))
	defer server.Close()

	client := NewClientWithOptions(context.Background(), "test-token", server.URL, Timeout, MaxRetries, "test")

	// Delete with empty body should not error
	err := client.Delete(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("Expected no error when deleting with empty body, got: %s", err)
	}
}

func TestClientRetryRespectsContextCancellation(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// Always return 500 to trigger retries
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClientWithOptions(context.Background(), "test-token", server.URL, Timeout, 3, "test")

	// Create a context that will be cancelled after first attempt completes
	ctx, cancel := context.WithCancel(context.Background())

	// Make the request in a goroutine
	done := make(chan error, 1)
	go func() {
		err := client.Get(ctx, "/test", nil, nil)
		done <- err
	}()

	// Wait for first request to complete (should be fast since server responds immediately)
	time.Sleep(50 * time.Millisecond)

	// Cancel context to prevent retries
	cancel()

	// Wait for completion
	err := <-done

	// Should fail with context error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "canceled")
	// Should have made at least one attempt
	assert.GreaterOrEqual(t, attempts, 1)
}

func TestClientGetResponseEnvelope(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []map[string]interface{}
	}{
		{
			name: "envelope with success and data",
			body: `{"success":true,"data":[{"id":1}],"errors":[],"meta":{"total":1}}`,
			want: []map[string]interface{}{{"id": float64(1)}},
		},
		{
			name: "bare top level array",
			body: `[{"id":1}]`,
			want: []map[string]interface{}{{"id": float64(1)}},
		},
		{
			name: "empty envelope data",
			body: `{"success":true,"data":[]}`,
			want: []map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClientWithOptions(context.Background(), "test-token", server.URL, Timeout, MaxRetries, "test")

			var result []map[string]interface{}
			require.NoError(t, client.Get(context.Background(), "/test", nil, &result))
			assert.Equal(t, tt.want, result)
		})
	}
}

// The alerts-configurations and email-configurations endpoints use "data" as
// their own payload format, without a "success" field. Those bodies must be
// decoded unchanged rather than unwrapped.
func TestClientGetDataWithoutSuccessIsNotUnwrapped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":1}]}`))
	}))
	defer server.Close()

	client := NewClientWithOptions(context.Background(), "test-token", server.URL, Timeout, MaxRetries, "test")

	var result map[string]interface{}
	require.NoError(t, client.Get(context.Background(), "/test", nil, &result))
	assert.Equal(t, []interface{}{map[string]interface{}{"id": float64(1)}}, result["data"])
}
