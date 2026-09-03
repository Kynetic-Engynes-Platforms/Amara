package types

import (
	"fmt"
	"time"

	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/ops"
)

// APIError represents structured error responses returned by Typesense server.
type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("typesense api error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("typesense api error (status %d)", e.StatusCode)
}

// Config defines the configuration needed to connect to Typesense cluster nodes.
type Config struct {
	APIKey         string        `json:"api_key"`
	Nodes          []string      `json:"nodes"`
	NumRetries     int           `json:"num_retries,omitempty"`
	RetryInterval  time.Duration `json:"retry_interval,omitempty"`
	HealthWaitTime time.Duration `json:"health_wait_time,omitempty"`
}

// Client is the thread-safe root SDK instance connecting all sub-services via interfaces.
type Client struct {
	Config      Config
	Http        ops.HTTPClient
	Nodes       ops.NodeManager
	Collections ops.CollectionsService
	Aliases     ops.AliasesService
	Keys        ops.KeysService
	Analytics   ops.AnalyticsService
	Operations  ops.OperationsService
}

// Collection represents a scoped handle to a single collection.
type Collection[T any] struct {
	Name      string
	Documents ops.DocumentsService[T]
	Overrides ops.OverridesService
	Synonyms  ops.SynonymsService
}
