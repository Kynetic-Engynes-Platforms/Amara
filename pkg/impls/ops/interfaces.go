package ops

import (
	"context"
	"io"
	"net/http"

	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/types/schemas"
)

// HTTPClient defines the interface for making HTTP requests, allowing for easy mocking during tests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NodeManager handles cluster node routing and failover.
type NodeManager interface {
	GetNextNode() string
	MarkUnhealthy(rawURL string)
	MarkHealthy(rawURL string)
}

// CollectionsService manages Typesense collections.
type CollectionsService interface {
	Create(ctx context.Context, schema schemas.Schema) (schemas.CollectionResponse, error)
	Retrieve(ctx context.Context, name string) (schemas.CollectionResponse, error)
	RetrieveAll(ctx context.Context) ([]schemas.CollectionResponse, error)
	Delete(ctx context.Context, name string) (schemas.CollectionResponse, error)
	Update(ctx context.Context, name string, schema schemas.Schema) (schemas.CollectionResponse, error)
}

// DocumentsService manages documents within a specific collection.
type DocumentsService[T any] interface {
	Create(ctx context.Context, doc T) (T, error)
	Upsert(ctx context.Context, doc T) (T, error)
	Update(ctx context.Context, id string, doc T) (T, error)
	Retrieve(ctx context.Context, id string) (T, error)
	Delete(ctx context.Context, id string) (T, error)
	DeleteByQuery(ctx context.Context, filterBy string) (schemas.DeleteByQueryResponse, error)
	Search(ctx context.Context, params schemas.SearchParams) (schemas.SearchResult[T], error)
	ImportBatch(ctx context.Context, docs []T, action string) ([]byte, error)
	ImportStream(ctx context.Context, body io.Reader, action string) ([]byte, error)
}

// Admin Services Interfaces
type OverridesService interface {
	Upsert(ctx context.Context, id string, schema schemas.OverrideSchema) (schemas.Override, error)
	Retrieve(ctx context.Context, id string) (schemas.Override, error)
	Delete(ctx context.Context, id string) (schemas.Override, error)
	RetrieveAll(ctx context.Context) ([]schemas.Override, error)
}

type SynonymsService interface {
	Upsert(ctx context.Context, id string, schema schemas.SynonymSchema) (schemas.Synonym, error)
	Retrieve(ctx context.Context, id string) (schemas.Synonym, error)
	Delete(ctx context.Context, id string) (schemas.Synonym, error)
	RetrieveAll(ctx context.Context) ([]schemas.Synonym, error)
}

type AliasesService interface {
	Upsert(ctx context.Context, name string, schema schemas.AliasSchema) (schemas.Alias, error)
	Retrieve(ctx context.Context, name string) (schemas.Alias, error)
	Delete(ctx context.Context, name string) (schemas.Alias, error)
	RetrieveAll(ctx context.Context) ([]schemas.Alias, error)
}

type KeysService interface {
	Create(ctx context.Context, schema schemas.KeySchema) (schemas.Key, error)
	Retrieve(ctx context.Context, id int64) (schemas.Key, error)
	Delete(ctx context.Context, id int64) (schemas.Key, error)
	RetrieveAll(ctx context.Context) ([]schemas.Key, error)
}

type AnalyticsService interface {
	CreateRule(ctx context.Context, name string, schema schemas.AnalyticsRuleSchema) (schemas.AnalyticsRule, error)
	RetrieveAll(ctx context.Context) ([]schemas.AnalyticsRule, error)
}

type OperationsService interface {
	Health(ctx context.Context) (schemas.HealthResponse, error)
	Metrics(ctx context.Context) (schemas.MetricsResponse, error)
}
