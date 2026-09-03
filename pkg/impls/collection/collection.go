package collection

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	adminservices "github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/admin_services"
	dorequest "github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/do_request"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/documents"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/types"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/types/schemas"
)

type CollectionsService struct {
	Client *types.Client
}

func NewCollection[T any](client *types.Client, name string) *types.Collection[T] {
	return &types.Collection[T]{
		Name:      name,
		Documents: documents.NewDocumentsService[T](client, name),
		Overrides: &adminservices.OverridesService{Client: client, CollectionName: name},
		Synonyms:  &adminservices.SynonymsService{Client: client, CollectionName: name},
	}
}

func (s *CollectionsService) Create(ctx context.Context, schema schemas.Schema) (schemas.CollectionResponse, error) {
	return dorequest.DoRequest[schemas.CollectionResponse](ctx, s.Client, http.MethodPost, "/collections", schema)
}

func (s *CollectionsService) Retrieve(ctx context.Context, name string) (schemas.CollectionResponse, error) {
	if name == "" {
		return schemas.CollectionResponse{}, fmt.Errorf("collection name cannot be empty")
	}
	path := fmt.Sprintf("/collections/%s", url.PathEscape(name))
	return dorequest.DoRequest[schemas.CollectionResponse](ctx, s.Client, http.MethodGet, path, nil)
}

func (s *CollectionsService) RetrieveAll(ctx context.Context) ([]schemas.CollectionResponse, error) {
	return dorequest.DoRequest[[]schemas.CollectionResponse](ctx, s.Client, http.MethodGet, "/collections", nil)
}

func (s *CollectionsService) Delete(ctx context.Context, name string) (schemas.CollectionResponse, error) {
	if name == "" {
		return schemas.CollectionResponse{}, fmt.Errorf("collection name cannot be empty")
	}
	path := fmt.Sprintf("/collections/%s", url.PathEscape(name))
	return dorequest.DoRequest[schemas.CollectionResponse](ctx, s.Client, http.MethodDelete, path, nil)
}

func (s *CollectionsService) Update(ctx context.Context, name string, schema schemas.Schema) (schemas.CollectionResponse, error) {
	if name == "" {
		return schemas.CollectionResponse{}, fmt.Errorf("collection name cannot be empty")
	}
	path := fmt.Sprintf("/collections/%s", url.PathEscape(name))
	return dorequest.DoRequest[schemas.CollectionResponse](ctx, s.Client, http.MethodPatch, path, schema)
}
