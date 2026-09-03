package dorequest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/types"
)

// DoRequest executes an HTTP request across configured nodes with automatic retry and failover.
func DoRequest[T any](ctx context.Context, c *types.Client, method, path string, payload any) (T, error) {
	var zero T
	var bodyBytes []byte
	var err error

	if payload != nil {
		bodyBytes, err = json.Marshal(payload)
		if err != nil {
			return zero, fmt.Errorf("marshaling request payload failed: %w", err)
		}
	}

	retries := c.Config.NumRetries
	if retries <= 0 {
		retries = 1
	}

	var lastErr error

	for attempt := 0; attempt < retries; attempt++ {
		nodeURL := c.Nodes.GetNextNode()
		fullURL := strings.TrimRight(nodeURL, "/") + path

		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return zero, fmt.Errorf("creating HTTP request failed: %w", err)
		}

		req.Header.Set("X-TYPESENSE-API-KEY", c.Config.APIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := c.Http.Do(req)
		if err != nil {
			c.Nodes.MarkUnhealthy(nodeURL)
			lastErr = fmt.Errorf("node %s connection error: %w", nodeURL, err)

			// Context-aware wait
			select {
			case <-ctx.Done():
				return zero, fmt.Errorf("request context cancelled: %w", ctx.Err())
			case <-time.After(c.Config.RetryInterval):
			}
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			c.Nodes.MarkUnhealthy(nodeURL)
			lastErr = fmt.Errorf("reading response body from %s failed: %w", nodeURL, err)
			continue
		}

		if resp.StatusCode >= http.StatusInternalServerError {
			c.Nodes.MarkUnhealthy(nodeURL)
			lastErr = &types.APIError{StatusCode: resp.StatusCode, Message: string(respBody)}

			select {
			case <-ctx.Done():
				return zero, fmt.Errorf("request context cancelled: %w", ctx.Err())
			case <-time.After(c.Config.RetryInterval):
			}
			continue
		}

		if resp.StatusCode >= http.StatusBadRequest {
			var apiErr types.APIError
			apiErr.StatusCode = resp.StatusCode
			if err := json.Unmarshal(respBody, &apiErr); err != nil {
				apiErr.Message = string(respBody)
			}
			return zero, &apiErr
		}

		c.Nodes.MarkHealthy(nodeURL)

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
