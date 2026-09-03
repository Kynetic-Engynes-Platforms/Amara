package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// HTTPClient defines the interface for making HTTP requests, allowing for easy mocking during tests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewClient initializes a new Typesense SDK client.
// It configures connection pooling, timeouts, and wires up all service interfaces.
func NewClient(cfg Config) (*Client, error) {
	if len(cfg.Nodes) == 0 {
		cfg.Nodes = []string{"http://localhost:8108"}
	}
	if cfg.NumRetries == 0 {
		cfg.NumRetries = 3
	}
	if cfg.HealthWaitTime == 0 {
		cfg.HealthWaitTime = 60 * time.Second
	}
	if cfg.RetryInterval == 0 {
		cfg.RetryInterval = 100 * time.Millisecond
	}

	nm, err := NewNodeManager(cfg.Nodes, cfg.HealthWaitTime)
	if err != nil {
		return nil, err
	}

	// Institutional-grade pooled HTTP transport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	client := &Client{
		config: cfg,
		http:   httpClient,
		nodes:  nm,
	}

	// Wire up the unexported concrete implementations to the public interfaces
	client.Collections = &collectionsService{client: client}
	client.Aliases = &aliasesService{client: client}
	client.Keys = &keysService{client: client}
	client.Analytics = &analyticsService{client: client}
	client.Operations = &operationsService{client: client}

	return client, nil
}

// doRequest executes an HTTP request across configured nodes with automatic retry and failover.
func doRequest[T any](ctx context.Context, c *Client, method, path string, payload any) (T, error) {
	var zero T
	var bodyBytes []byte
	var err error

	if payload != nil {
		bodyBytes, err = json.Marshal(payload)
		if err != nil {
			return zero, fmt.Errorf("marshaling request payload failed: %w", err)
		}
	}

	retries := c.config.NumRetries
	if retries <= 0 {
		retries = 1
	}

	var lastErr error

	for attempt := 0; attempt < retries; attempt++ {
		nodeURL := c.nodes.GetNextNode()
		fullURL := strings.TrimRight(nodeURL, "/") + path

		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return zero, fmt.Errorf("creating HTTP request failed: %w", err)
		}

		req.Header.Set("X-TYPESENSE-API-KEY", c.config.APIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			// Network error, mark unhealthy and retry
			c.nodes.MarkUnhealthy(nodeURL)
			lastErr = fmt.Errorf("node %s connection error: %w", nodeURL, err)
			time.Sleep(c.config.RetryInterval)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			c.nodes.MarkUnhealthy(nodeURL)
			lastErr = fmt.Errorf("reading response body from %s failed: %w", nodeURL, err)
			continue
		}

		// Handle 5xx server errors via failover
		if resp.StatusCode >= http.StatusInternalServerError {
			c.nodes.MarkUnhealthy(nodeURL)
			lastErr = &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
			time.Sleep(c.config.RetryInterval)
			continue
		}

		// Handle 4xx client errors directly without retrying (e.g., Not Found, Unauthorized)
		if resp.StatusCode >= http.StatusBadRequest {
			var apiErr APIError
			apiErr.StatusCode = resp.StatusCode
			if err := json.Unmarshal(respBody, &apiErr); err != nil {
				apiErr.Message = string(respBody)
			}
			return zero, &apiErr
		}

		// Success: Mark node healthy
		c.nodes.MarkHealthy(nodeURL)

		// Handle empty responses gracefully
		if len(respBody) == 0 {
			return zero, nil
		}

		var result T
		if err := json.Unmarshal(respBody, &result); err != nil {
			return zero, fmt.Errorf("decoding response failed: %w", err)
		}

		return result, nil
	}

	return zero, fmt.Errorf("request failed after %d retries across nodes; last error: %w", retries, lastErr)
}
