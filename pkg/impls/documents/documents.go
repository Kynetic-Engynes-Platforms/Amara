package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	dorequest "github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/do_request"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/ops"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/types"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/types/schemas"
)

type DocumentsService[T any] struct {
	client         *types.Client
	collectionName string
}

// Constructor returns the DocumentsService[T] interface
func NewDocumentsService[T any](client *types.Client, collectionName string) ops.DocumentsService[T] {
	return &DocumentsService[T]{
		client:         client,
		collectionName: collectionName,
	}
}

func (d *DocumentsService[T]) Create(ctx context.Context, doc T) (T, error) {
	path := fmt.Sprintf("/collections/%s/documents", url.PathEscape(d.collectionName))
	return dorequest.DoRequest[T](ctx, d.client, http.MethodPost, path, doc)
}

func (d *DocumentsService[T]) Upsert(ctx context.Context, doc T) (T, error) {
	path := fmt.Sprintf("/collections/%s/documents?action=upsert", url.PathEscape(d.collectionName))
	return dorequest.DoRequest[T](ctx, d.client, http.MethodPost, path, doc)
}

func (d *DocumentsService[T]) Update(ctx context.Context, id string, doc T) (T, error) {
	path := fmt.Sprintf("/collections/%s/documents/%s", url.PathEscape(d.collectionName), url.PathEscape(id))
	return dorequest.DoRequest[T](ctx, d.client, http.MethodPatch, path, doc)
}

func (d *DocumentsService[T]) Retrieve(ctx context.Context, id string) (T, error) {
	path := fmt.Sprintf("/collections/%s/documents/%s", url.PathEscape(d.collectionName), url.PathEscape(id))
	return dorequest.DoRequest[T](ctx, d.client, http.MethodGet, path, nil)
}

func (d *DocumentsService[T]) Delete(ctx context.Context, id string) (T, error) {
	path := fmt.Sprintf("/collections/%s/documents/%s", url.PathEscape(d.collectionName), url.PathEscape(id))
	return dorequest.DoRequest[T](ctx, d.client, http.MethodDelete, path, nil)
}

func (d *DocumentsService[T]) DeleteByQuery(ctx context.Context, filterBy string) (schemas.DeleteByQueryResponse, error) {
	v := url.Values{}
	v.Set("filter_by", filterBy)
	path := fmt.Sprintf("/collections/%s/documents?%s", url.PathEscape(d.collectionName), v.Encode())
	return dorequest.DoRequest[schemas.DeleteByQueryResponse](ctx, d.client, http.MethodDelete, path, nil)
}

func (d *DocumentsService[T]) Search(ctx context.Context, params schemas.SearchParams) (schemas.SearchResult[T], error) {
	v := url.Values{}
	if params.Q != "" {
		v.Set("q", params.Q)
	}
	if params.QueryBy != "" {
		v.Set("query_by", params.QueryBy)
	}
	if params.FilterBy != "" {
		v.Set("filter_by", params.FilterBy)
	}
	if params.SortBy != "" {
		v.Set("sort_by", params.SortBy)
	}
	if params.FacetBy != "" {
		v.Set("facet_by", params.FacetBy)
	}
	if params.Page > 0 {
		v.Set("page", strconv.Itoa(params.Page))
	}
	if params.PerPage > 0 {
		v.Set("per_page", strconv.Itoa(params.PerPage))
	}
	if params.GroupBy != "" {
		v.Set("group_by", params.GroupBy)
	}
	if params.VectorQuery != "" {
		v.Set("vector_query", params.VectorQuery)
	}

	path := fmt.Sprintf("/collections/%s/documents/search?%s", url.PathEscape(d.collectionName), v.Encode())
	return dorequest.DoRequest[schemas.SearchResult[T]](ctx, d.client, http.MethodGet, path, nil)
}

// ImportBatch streams JSONL documents into Typesense in bulk.
func (d *DocumentsService[T]) ImportBatch(ctx context.Context, docs []T, action string) ([]byte, error) {
	if len(docs) == 0 {
		return nil, nil
	}
	if action == "" {
		action = "create"
	}

	var buf bytes.Buffer
	for _, doc := range docs {
		b, err := json.Marshal(doc)
		if err != nil {
			return nil, fmt.Errorf("failed marshaling batch item: %w", err)
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}

	path := fmt.Sprintf("/collections/%s/documents/import?action=%s", url.PathEscape(d.collectionName), url.QueryEscape(action))
	nodeURL := d.client.Nodes.GetNextNode()
	fullURL := strings.TrimRight(nodeURL, "/") + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-TYPESENSE-API-KEY", d.client.Config.APIKey)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := d.client.Http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, &types.APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	return respBody, nil
}

func (d *DocumentsService[T]) ImportStream(ctx context.Context, body io.Reader, action string) ([]byte, error) {
	if action == "" {
		action = "create"
	}

	path := fmt.Sprintf("/collections/%s/documents/import?action=%s", url.PathEscape(d.collectionName), url.QueryEscape(action))
	nodeURL := d.client.Nodes.GetNextNode()
	fullURL := strings.TrimRight(nodeURL, "/") + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-TYPESENSE-API-KEY", d.client.Config.APIKey)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := d.client.Http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, &types.APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	return respBody, nil
}
