package adminservices

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	dorequest "github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/do_request"
	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/types"
	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/types/schemas"
)

type OverridesService struct {
	Client         *types.Client
	CollectionName string
}

func (s *OverridesService) Upsert(ctx context.Context, id string, schema schemas.OverrideSchema) (schemas.Override, error) {
	path := fmt.Sprintf("/collections/%s/overrides/%s", url.PathEscape(s.CollectionName), url.PathEscape(id))
	return dorequest.DoRequest[schemas.Override](ctx, s.Client, http.MethodPut, path, schema)
}

func (s *OverridesService) Retrieve(ctx context.Context, id string) (schemas.Override, error) {
	path := fmt.Sprintf("/collections/%s/overrides/%s", url.PathEscape(s.CollectionName), url.PathEscape(id))
	return dorequest.DoRequest[schemas.Override](ctx, s.Client, http.MethodGet, path, nil)
}

func (s *OverridesService) Delete(ctx context.Context, id string) (schemas.Override, error) {
	path := fmt.Sprintf("/collections/%s/overrides/%s", url.PathEscape(s.CollectionName), url.PathEscape(id))
	return dorequest.DoRequest[schemas.Override](ctx, s.Client, http.MethodDelete, path, nil)
}

func (s *OverridesService) RetrieveAll(ctx context.Context) ([]schemas.Override, error) {
	type responseWrapper struct {
		Overrides []schemas.Override `json:"overrides"`
	}
	path := fmt.Sprintf("/collections/%s/overrides", url.PathEscape(s.CollectionName))
	w, err := dorequest.DoRequest[responseWrapper](ctx, s.Client, http.MethodGet, path, nil)
	return w.Overrides, err
}

// synonymsService
type SynonymsService struct {
	Client         *types.Client
	CollectionName string
}

func (s *SynonymsService) Upsert(ctx context.Context, id string, schema schemas.SynonymSchema) (schemas.Synonym, error) {
	path := fmt.Sprintf("/collections/%s/synonyms/%s", url.PathEscape(s.CollectionName), url.PathEscape(id))
	return dorequest.DoRequest[schemas.Synonym](ctx, s.Client, http.MethodPut, path, schema)
}

func (s *SynonymsService) Retrieve(ctx context.Context, id string) (schemas.Synonym, error) {
	path := fmt.Sprintf("/collections/%s/synonyms/%s", url.PathEscape(s.CollectionName), url.PathEscape(id))
	return dorequest.DoRequest[schemas.Synonym](ctx, s.Client, http.MethodGet, path, nil)
}

func (s *SynonymsService) Delete(ctx context.Context, id string) (schemas.Synonym, error) {
	path := fmt.Sprintf("/collections/%s/synonyms/%s", url.PathEscape(s.CollectionName), url.PathEscape(id))
	return dorequest.DoRequest[schemas.Synonym](ctx, s.Client, http.MethodDelete, path, nil)
}

func (s *SynonymsService) RetrieveAll(ctx context.Context) ([]schemas.Synonym, error) {
	type wrapper struct {
		Synonyms []schemas.Synonym `json:"synonyms"`
	}
	path := fmt.Sprintf("/collections/%s/synonyms", url.PathEscape(s.CollectionName))
	w, err := dorequest.DoRequest[wrapper](ctx, s.Client, http.MethodGet, path, nil)
	return w.Synonyms, err
}

type AliasesService struct {
	Client *types.Client
}

func (s *AliasesService) Upsert(ctx context.Context, name string, schema schemas.AliasSchema) (schemas.Alias, error) {
	path := fmt.Sprintf("/aliases/%s", url.PathEscape(name))
	return dorequest.DoRequest[schemas.Alias](ctx, s.Client, http.MethodPut, path, schema)
}

func (s *AliasesService) Retrieve(ctx context.Context, name string) (schemas.Alias, error) {
	path := fmt.Sprintf("/aliases/%s", url.PathEscape(name))
	return dorequest.DoRequest[schemas.Alias](ctx, s.Client, http.MethodGet, path, nil)
}

func (s *AliasesService) Delete(ctx context.Context, name string) (schemas.Alias, error) {
	path := fmt.Sprintf("/aliases/%s", url.PathEscape(name))
	return dorequest.DoRequest[schemas.Alias](ctx, s.Client, http.MethodDelete, path, nil)
}

func (s *AliasesService) RetrieveAll(ctx context.Context) ([]schemas.Alias, error) {
	type wrapper struct {
		Aliases []schemas.Alias `json:"aliases"`
	}
	w, err := dorequest.DoRequest[wrapper](ctx, s.Client, http.MethodGet, "/aliases", nil)
	return w.Aliases, err
}

type KeysService struct {
	Client *types.Client
}

func (s *KeysService) Create(ctx context.Context, schema schemas.KeySchema) (schemas.Key, error) {
	return dorequest.DoRequest[schemas.Key](ctx, s.Client, http.MethodPost, "/keys", schema)
}

func (s *KeysService) Retrieve(ctx context.Context, id int64) (schemas.Key, error) {
	path := fmt.Sprintf("/keys/%d", id)
	return dorequest.DoRequest[schemas.Key](ctx, s.Client, http.MethodGet, path, nil)
}

func (s *KeysService) Delete(ctx context.Context, id int64) (schemas.Key, error) {
	path := fmt.Sprintf("/keys/%d", id)
	return dorequest.DoRequest[schemas.Key](ctx, s.Client, http.MethodDelete, path, nil)
}

func (s *KeysService) RetrieveAll(ctx context.Context) ([]schemas.Key, error) {
	type wrapper struct {
		Keys []schemas.Key `json:"keys"`
	}
	w, err := dorequest.DoRequest[wrapper](ctx, s.Client, http.MethodGet, "/keys", nil)
	return w.Keys, err
}

type AnalyticsService struct {
	Client *types.Client
}

func (s *AnalyticsService) CreateRule(ctx context.Context, name string, schema schemas.AnalyticsRuleSchema) (schemas.AnalyticsRule, error) {
	path := fmt.Sprintf("/analytics/rules/%s", url.PathEscape(name))
	return dorequest.DoRequest[schemas.AnalyticsRule](ctx, s.Client, http.MethodPut, path, schema)
}

func (s *AnalyticsService) RetrieveAll(ctx context.Context) ([]schemas.AnalyticsRule, error) {
	type wrapper struct {
		Rules []schemas.AnalyticsRule `json:"rules"`
	}
	w, err := dorequest.DoRequest[wrapper](ctx, s.Client, http.MethodGet, "/analytics/rules", nil)
	return w.Rules, err
}

type OperationsService struct {
	Client *types.Client
}

func (s *OperationsService) Health(ctx context.Context) (schemas.HealthResponse, error) {
	return dorequest.DoRequest[schemas.HealthResponse](ctx, s.Client, http.MethodGet, "/health", nil)
}

func (s *OperationsService) Metrics(ctx context.Context) (schemas.MetricsResponse, error) {
	return dorequest.DoRequest[schemas.MetricsResponse](ctx, s.Client, http.MethodGet, "/metrics.json", nil)
}
