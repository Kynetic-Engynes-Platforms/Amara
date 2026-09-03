package sdk

import "context"

// NodeManager handles cluster node routing and failover.
type NodeManager interface {
	GetNextNode() string
	MarkUnhealthy(rawURL string)
	MarkHealthy(rawURL string)
}

// CollectionsService manages Typesense collections.
type CollectionsService interface {
	Create(ctx context.Context, schema Schema) (CollectionResponse, error)
	Retrieve(ctx context.Context, name string) (CollectionResponse, error)
	RetrieveAll(ctx context.Context) ([]CollectionResponse, error)
	Delete(ctx context.Context, name string) (CollectionResponse, error)
	Update(ctx context.Context, name string, schema Schema) (CollectionResponse, error)
}

// DocumentsService manages documents within a specific collection.
type DocumentsService[T any] interface {
	Create(ctx context.Context, doc T) (T, error)
	Upsert(ctx context.Context, doc T) (T, error)
	Update(ctx context.Context, id string, doc T) (T, error)
	Retrieve(ctx context.Context, id string) (T, error)
	Delete(ctx context.Context, id string) (T, error)
	DeleteByQuery(ctx context.Context, filterBy string) (DeleteByQueryResponse, error)
	Search(ctx context.Context, params SearchParams) (SearchResult[T], error)
	ImportBatch(ctx context.Context, docs []T, action string) ([]byte, error)
}

// Admin Services Interfaces
type OverridesService interface {
	Upsert(ctx context.Context, id string, schema OverrideSchema) (Override, error)
	Retrieve(ctx context.Context, id string) (Override, error)
	Delete(ctx context.Context, id string) (Override, error)
	RetrieveAll(ctx context.Context) ([]Override, error)
}

type SynonymsService interface {
	Upsert(ctx context.Context, id string, schema SynonymSchema) (Synonym, error)
	Retrieve(ctx context.Context, id string) (Synonym, error)
	Delete(ctx context.Context, id string) (Synonym, error)
	RetrieveAll(ctx context.Context) ([]Synonym, error)
}

type AliasesService interface {
	Upsert(ctx context.Context, name string, schema AliasSchema) (Alias, error)
	Retrieve(ctx context.Context, name string) (Alias, error)
	Delete(ctx context.Context, name string) (Alias, error)
	RetrieveAll(ctx context.Context) ([]Alias, error)
}

type KeysService interface {
	Create(ctx context.Context, schema KeySchema) (Key, error)
	Retrieve(ctx context.Context, id int64) (Key, error)
	Delete(ctx context.Context, id int64) (Key, error)
	RetrieveAll(ctx context.Context) ([]Key, error)
}

type AnalyticsService interface {
	CreateRule(ctx context.Context, name string, schema AnalyticsRuleSchema) (AnalyticsRule, error)
	RetrieveAll(ctx context.Context) ([]AnalyticsRule, error)
}

type OperationsService interface {
	Health(ctx context.Context) (HealthResponse, error)
	Metrics(ctx context.Context) (MetricsResponse, error)
}
