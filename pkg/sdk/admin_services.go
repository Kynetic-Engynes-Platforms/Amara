package sdk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// overridesService
type overridesService struct {
	client         *Client
	collectionName string
}

func (s *overridesService) Upsert(ctx context.Context, id string, schema OverrideSchema) (Override, error) {
	path := fmt.Sprintf("/collections/%s/overrides/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
	return doRequest[Override](ctx, s.client, http.MethodPut, path, schema)
}

func (s *overridesService) Retrieve(ctx context.Context, id string) (Override, error) {
	path := fmt.Sprintf("/collections/%s/overrides/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
	return doRequest[Override](ctx, s.client, http.MethodGet, path, nil)
}

func (s *overridesService) Delete(ctx context.Context, id string) (Override, error) {
	path := fmt.Sprintf("/collections/%s/overrides/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
	return doRequest[Override](ctx, s.client, http.MethodDelete, path, nil)
}

func (s *overridesService) RetrieveAll(ctx context.Context) ([]Override, error) {
	type responseWrapper struct {
		Overrides []Override `json:"overrides"`
	}
	path := fmt.Sprintf("/collections/%s/overrides", url.PathEscape(s.collectionName))
	w, err := doRequest[responseWrapper](ctx, s.client, http.MethodGet, path, nil)
	return w.Overrides, err
}

// synonymsService
type synonymsService struct {
	client         *Client
	collectionName string
}

func (s *synonymsService) Upsert(ctx context.Context, id string, schema SynonymSchema) (Synonym, error) {
	path := fmt.Sprintf("/collections/%s/synonyms/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
	return doRequest[Synonym](ctx, s.client, http.MethodPut, path, schema)
}

func (s *synonymsService) Retrieve(ctx context.Context, id string) (Synonym, error) {
	path := fmt.Sprintf("/collections/%s/synonyms/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
	return doRequest[Synonym](ctx, s.client, http.MethodGet, path, nil)
}

func (s *synonymsService) Delete(ctx context.Context, id string) (Synonym, error) {
	path := fmt.Sprintf("/collections/%s/synonyms/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
	return doRequest[Synonym](ctx, s.client, http.MethodDelete, path, nil)
}

func (s *synonymsService) RetrieveAll(ctx context.Context) ([]Synonym, error) {
	type wrapper struct {
		Synonyms []Synonym `json:"synonyms"`
	}
	path := fmt.Sprintf("/collections/%s/synonyms", url.PathEscape(s.collectionName))
	w, err := doRequest[wrapper](ctx, s.client, http.MethodGet, path, nil)
	return w.Synonyms, err
}

// aliasesService
type aliasesService struct {
	client *Client
}

func (s *aliasesService) Upsert(ctx context.Context, name string, schema AliasSchema) (Alias, error) {
	path := fmt.Sprintf("/aliases/%s", url.PathEscape(name))
	return doRequest[Alias](ctx, s.client, http.MethodPut, path, schema)
}

func (s *aliasesService) Retrieve(ctx context.Context, name string) (Alias, error) {
	path := fmt.Sprintf("/aliases/%s", url.PathEscape(name))
	return doRequest[Alias](ctx, s.client, http.MethodGet, path, nil)
}

func (s *aliasesService) Delete(ctx context.Context, name string) (Alias, error) {
	path := fmt.Sprintf("/aliases/%s", url.PathEscape(name))
	return doRequest[Alias](ctx, s.client, http.MethodDelete, path, nil)
}

func (s *aliasesService) RetrieveAll(ctx context.Context) ([]Alias, error) {
	type wrapper struct {
		Aliases []Alias `json:"aliases"`
	}
	w, err := doRequest[wrapper](ctx, s.client, http.MethodGet, "/aliases", nil)
	return w.Aliases, err
}

// keysService
type keysService struct {
	client *Client
}

func (s *keysService) Create(ctx context.Context, schema KeySchema) (Key, error) {
	return doRequest[Key](ctx, s.client, http.MethodPost, "/keys", schema)
}

func (s *keysService) Retrieve(ctx context.Context, id int64) (Key, error) {
	path := fmt.Sprintf("/keys/%d", id)
	return doRequest[Key](ctx, s.client, http.MethodGet, path, nil)
}

func (s *keysService) Delete(ctx context.Context, id int64) (Key, error) {
	path := fmt.Sprintf("/keys/%d", id)
	return doRequest[Key](ctx, s.client, http.MethodDelete, path, nil)
}

func (s *keysService) RetrieveAll(ctx context.Context) ([]Key, error) {
	type wrapper struct {
		Keys []Key `json:"keys"`
	}
	w, err := doRequest[wrapper](ctx, s.client, http.MethodGet, "/keys", nil)
	return w.Keys, err
}

// analyticsService
type analyticsService struct {
	client *Client
}

func (s *analyticsService) CreateRule(ctx context.Context, name string, schema AnalyticsRuleSchema) (AnalyticsRule, error) {
	path := fmt.Sprintf("/analytics/rules/%s", url.PathEscape(name))
	return doRequest[AnalyticsRule](ctx, s.client, http.MethodPut, path, schema)
}

func (s *analyticsService) RetrieveAll(ctx context.Context) ([]AnalyticsRule, error) {
	type wrapper struct {
		Rules []AnalyticsRule `json:"rules"`
	}
	w, err := doRequest[wrapper](ctx, s.client, http.MethodGet, "/analytics/rules", nil)
	return w.Rules, err
}

// operationsService
type operationsService struct {
	client *Client
}

func (s *operationsService) Health(ctx context.Context) (HealthResponse, error) {
	return doRequest[HealthResponse](ctx, s.client, http.MethodGet, "/health", nil)
}

func (s *operationsService) Metrics(ctx context.Context) (MetricsResponse, error) {
	return doRequest[MetricsResponse](ctx, s.client, http.MethodGet, "/metrics.json", nil)
}
