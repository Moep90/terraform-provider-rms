package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	client.setRequestHeaders(req)

	assert.Equal(t, UserAgent, req.Header.Get("User-Agent"))
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
		w.Write([]byte(`{"data": "test"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token: "test-token",
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
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token: "test-token",
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
		w.Write([]byte(`{"updated": true}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token: "test-token",
	}

	var result map[string]interface{}
	err := client.Put(context.Background(), "/test", nil, &result)

	require.NoError(t, err)
	assert.True(t, result["updated"].(bool))
}

func TestClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"deleted": true}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token: "test-token",
	}

	var result map[string]interface{}
	err := client.Delete(context.Background(), "/test", &result)

	require.NoError(t, err)
	assert.True(t, result["deleted"].(bool))
}

func TestClientAuthenticationErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token: "invalid-token",
	}

	var result map[string]interface{}
	err := client.Get(context.Background(), "/test", nil, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestClientForbiddenErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "forbidden"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token: "test-token",
	}

	var result map[string]interface{}
	err := client.Get(context.Background(), "/test", nil, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}
