package sdk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type Field struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Optional bool   `json:"optional,omitempty"`
	Facet    bool   `json:"facet,omitempty"`
	Index    bool   `json:"index,omitempty"`
	Sort     bool   `json:"sort,omitempty"`
	Infix    bool   `json:"infix,omitempty"`
	Locale   string `json:"locale,omitempty"`
}

type Schema struct {
	Name                string   `json:"name"`
	Fields              []Field  `json:"fields"`
	DefaultSortingField string   `json:"default_sorting_field,omitempty"`
	EnableNestedFields  bool     `json:"enable_nested_fields,omitempty"`
	SymbolsToIndex      []string `json:"symbols_to_index,omitempty"`
	TokenSeparators     []string `json:"token_separators,omitempty"`
}

type CollectionResponse struct {
	Name                string  `json:"name"`
	NumDocuments        int64   `json:"num_documents"`
	Fields              []Field `json:"fields"`
	DefaultSortingField string  `json:"default_sorting_field,omitempty"`
	CreatedAt           int64   `json:"created_at,omitempty"`
}

type collectionsService struct {
	client *Client
}

func NewCollection[T any](client *Client, name string) *Collection[T] {
	return &Collection[T]{
		Name:      name,
		Documents: NewDocumentsService[T](client, name),
		Overrides: &overridesService{client: client, collectionName: name},
		Synonyms:  &synonymsService{client: client, collectionName: name},
	}
}

func (s *collectionsService) Create(ctx context.Context, schema Schema) (CollectionResponse, error) {
	return doRequest[CollectionResponse](ctx, s.client, http.MethodPost, "/collections", schema)
}

func (s *collectionsService) Retrieve(ctx context.Context, name string) (CollectionResponse, error) {
	if name == "" {
		return CollectionResponse{}, fmt.Errorf("collection name cannot be empty")
	}
	path := fmt.Sprintf("/collections/%s", url.PathEscape(name))
	return doRequest[CollectionResponse](ctx, s.client, http.MethodGet, path, nil)
}

func (s *collectionsService) RetrieveAll(ctx context.Context) ([]CollectionResponse, error) {
	return doRequest[[]CollectionResponse](ctx, s.client, http.MethodGet, "/collections", nil)
}

func (s *collectionsService) Delete(ctx context.Context, name string) (CollectionResponse, error) {
	if name == "" {
		return CollectionResponse{}, fmt.Errorf("collection name cannot be empty")
	}
	path := fmt.Sprintf("/collections/%s", url.PathEscape(name))
	return doRequest[CollectionResponse](ctx, s.client, http.MethodDelete, path, nil)
}

func (s *collectionsService) Update(ctx context.Context, name string, schema Schema) (CollectionResponse, error) {
	if name == "" {
		return CollectionResponse{}, fmt.Errorf("collection name cannot be empty")
	}
	path := fmt.Sprintf("/collections/%s", url.PathEscape(name))
	return doRequest[CollectionResponse](ctx, s.client, http.MethodPatch, path, schema)
}
