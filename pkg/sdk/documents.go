package sdk

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
)

type SearchParams struct {
	Q                   string `json:"q"`
	QueryBy             string `json:"query_by"`
	FilterBy            string `json:"filter_by,omitempty"`
	SortBy              string `json:"sort_by,omitempty"`
	FacetBy             string `json:"facet_by,omitempty"`
	Page                int    `json:"page,omitempty"`
	PerPage             int    `json:"per_page,omitempty"`
	GroupBy             string `json:"group_by,omitempty"`
	GroupLimit          int    `json:"group_limit,omitempty"`
	VectorQuery         string `json:"vector_query,omitempty"`
	IncludeFields       string `json:"include_fields,omitempty"`
	ExcludeFields       string `json:"exclude_fields,omitempty"`
	HighlightFullFields string `json:"highlight_full_fields,omitempty"`
}

type SearchHit[T any] struct {
	Document   T                `json:"document"`
	Highlights []map[string]any `json:"highlights,omitempty"`
	TextMatch  int64            `json:"text_match,omitempty"`
}

type SearchResult[T any] struct {
	Found        int            `json:"found"`
	OutOf        int            `json:"out_of"`
	Page         int            `json:"page"`
	SearchTimeMs int            `json:"search_time_ms"`
	Hits         []SearchHit[T] `json:"hits"`
}

type DeleteByQueryResponse struct {
	NumDeleted int `json:"num_deleted"`
}

type documentsService[T any] struct {
	client         *Client
	collectionName string
}

// Constructor returns the DocumentsService[T] interface
func NewDocumentsService[T any](client *Client, collectionName string) DocumentsService[T] {
	return &documentsService[T]{
		client:         client,
		collectionName: collectionName,
	}
}

func (d *documentsService[T]) Create(ctx context.Context, doc T) (T, error) {
	path := fmt.Sprintf("/collections/%s/documents", url.PathEscape(d.collectionName))
	return doRequest[T](ctx, d.client, http.MethodPost, path, doc)
}

func (d *documentsService[T]) Upsert(ctx context.Context, doc T) (T, error) {
	path := fmt.Sprintf("/collections/%s/documents?action=upsert", url.PathEscape(d.collectionName))
	return doRequest[T](ctx, d.client, http.MethodPost, path, doc)
}

func (d *documentsService[T]) Update(ctx context.Context, id string, doc T) (T, error) {
	path := fmt.Sprintf("/collections/%s/documents/%s", url.PathEscape(d.collectionName), url.PathEscape(id))
	return doRequest[T](ctx, d.client, http.MethodPatch, path, doc)
}

func (d *documentsService[T]) Retrieve(ctx context.Context, id string) (T, error) {
	path := fmt.Sprintf("/collections/%s/documents/%s", url.PathEscape(d.collectionName), url.PathEscape(id))
	return doRequest[T](ctx, d.client, http.MethodGet, path, nil)
}

func (d *documentsService[T]) Delete(ctx context.Context, id string) (T, error) {
	path := fmt.Sprintf("/collections/%s/documents/%s", url.PathEscape(d.collectionName), url.PathEscape(id))
	return doRequest[T](ctx, d.client, http.MethodDelete, path, nil)
}

func (d *documentsService[T]) DeleteByQuery(ctx context.Context, filterBy string) (DeleteByQueryResponse, error) {
	v := url.Values{}
	v.Set("filter_by", filterBy)
	path := fmt.Sprintf("/collections/%s/documents?%s", url.PathEscape(d.collectionName), v.Encode())
	return doRequest[DeleteByQueryResponse](ctx, d.client, http.MethodDelete, path, nil)
}

func (d *documentsService[T]) Search(ctx context.Context, params SearchParams) (SearchResult[T], error) {
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
	return doRequest[SearchResult[T]](ctx, d.client, http.MethodGet, path, nil)
}

// ImportBatch streams JSONL documents into Typesense in bulk.
func (d *documentsService[T]) ImportBatch(ctx context.Context, docs []T, action string) ([]byte, error) {
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
	nodeURL := d.client.nodes.GetNextNode()
	fullURL := strings.TrimRight(nodeURL, "/") + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-TYPESENSE-API-KEY", d.client.config.APIKey)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := d.client.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	return respBody, nil
}
