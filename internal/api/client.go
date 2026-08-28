package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	BaseURL    = "https://rms.teltonika-networks.com/api"
	UserAgent  = "Terraform-Provider-Teltonika-RMS/1.0.0"
	APIVersion = "v3"
	Timeout    = 30 * time.Second
	MaxRetries = 3
	RetryDelay = 1 * time.Second
)

// ErrNotFound is returned when a resource is not found (404)
var ErrNotFound = fmt.Errorf("resource not found")

// Client represents the Teltonika RMS API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	maxRetries int
}

// NewClient creates a new Teltonika RMS API client
func NewClient(ctx context.Context, token string) *Client {
	tflog.Info(ctx, "Creating Teltonika RMS API client", map[string]interface{}{
		"base_url": BaseURL,
	})

	return &Client{
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token:      token,
		maxRetries: MaxRetries,
	}
}

// NewClientWithBaseURL creates a client with a custom base URL (for testing)
func NewClientWithBaseURL(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
		token:      token,
		maxRetries: MaxRetries,
	}
}

// NewClientWithOptions creates a client with custom options
func NewClientWithOptions(ctx context.Context, token, baseURL string, timeout time.Duration, maxRetries int) *Client {
	tflog.Info(ctx, "Creating Teltonika RMS API client", map[string]interface{}{
		"base_url":  baseURL,
		"timeout":   timeout.String(),
		"max_retry": maxRetries,
	})

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		token:      token,
		maxRetries: maxRetries,
	}
}

// setRequestHeaders sets common headers for all requests
func (c *Client) setRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
}

// do executes an HTTP request with retry logic
func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) error {
	c.setRequestHeaders(req)

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			tflog.Warn(ctx, "Retrying request", map[string]interface{}{
				"attempt": attempt + 1,
				"delay":   RetryDelay.String(),
			})
			timer := time.NewTimer(RetryDelay * time.Duration(attempt))
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			_ = resp.Body.Close()
			return fmt.Errorf("unauthorized: invalid or expired token")
		}

		if resp.StatusCode == http.StatusForbidden {
			_ = resp.Body.Close()
			return fmt.Errorf("forbidden: insufficient permissions")
		}

		if resp.StatusCode >= 400 {
			err := c.handleErrorResponse(resp)
			_ = resp.Body.Close()
			return err
		}

		err = c.decodeResponse(resp, v)
		_ = resp.Body.Close()
		return err
	}

	return fmt.Errorf("request failed after %d attempts", c.maxRetries)
}

// handleErrorResponse decodes and returns error details from response
func (c *Client) handleErrorResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, resp.Status)
	}

	var errorResp map[string]interface{}
	if err := json.Unmarshal(body, &errorResp); err == nil {
		if errMsg, ok := errorResp["error"].(string); ok {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, errMsg)
		}
	}

	return fmt.Errorf("API error %d: %s", resp.StatusCode, resp.Status)
}

// decodeResponse decodes JSON response into target struct
func (c *Client) decodeResponse(resp *http.Response, v interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if v == nil || len(body) == 0 {
		return nil
	}

	return json.Unmarshal(body, v)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, params map[string]string, v interface{}) error {
	reqURL := c.baseURL + path

	// Add query parameters with proper escaping
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Add(k, v)
		}
		reqURL += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	return c.do(ctx, req, v)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body, v interface{}) error {
	reqURL := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bodyReader)
	if err != nil {
		return err
	}

	return c.do(ctx, req, v)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body, v interface{}) error {
	reqURL := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bodyReader)
	if err != nil {
		return err
	}

	return c.do(ctx, req, v)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string, v interface{}) error {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	return c.do(ctx, req, v)
}
