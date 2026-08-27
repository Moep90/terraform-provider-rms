package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		token: token,
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

	var lastErr error
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			tflog.Warn(ctx, "Retrying request", map[string]interface{}{
				"attempt": attempt + 1,
				"delay":   RetryDelay.String(),
			})
			time.Sleep(RetryDelay * time.Duration(attempt))
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %s", resp.Status)
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("unauthorized: invalid or expired token")
		}

		if resp.StatusCode == http.StatusForbidden {
			return fmt.Errorf("forbidden: insufficient permissions")
		}

		if resp.StatusCode >= 400 {
			return c.handleErrorResponse(resp)
		}

		return c.decodeResponse(resp, v)
	}

	return fmt.Errorf("request failed after %d attempts: %w", MaxRetries, lastErr)
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

	return json.Unmarshal(body, v)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, params map[string]string, v interface{}) error {
	reqURL := c.baseURL + path

	// Add query parameters
	if len(params) > 0 {
		q := ""
		for k, v := range params {
			if q != "" {
				q += "&"
			}
			q += k + "=" + v
		}
		if q != "" {
			reqURL += "?" + q
		}
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
