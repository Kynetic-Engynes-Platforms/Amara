package documents_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/connection"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/documents"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/types"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/types/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDoc is a strictly typed struct to validate generic [T] behavior.
type TestDoc struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Views int    `json:"views"`
}

// setupTestEnv initializes an httptest.Server and a wired SDK Client pointing to it.
func setupTestEnv(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *types.Client) {
	server := httptest.NewServer(handler)

	cfg := types.Config{
		APIKey:         "test-secret-key",
		Nodes:          []string{server.URL},
		NumRetries:     1,
		RetryInterval:  time.Millisecond,
		HealthWaitTime: time.Second,
	}

	client, err := connection.NewClient(cfg)
	require.NoError(t, err)

	return server, client
}

func TestDocumentsService_Create(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/collections/books/documents", r.URL.Path)
		assert.Equal(t, "test-secret-key", r.Header.Get("X-TYPESENSE-API-KEY"))

		body, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, `{"id":"1","title":"Go Programming","views":100}`, string(body))

		w.WriteHeader(http.StatusCreated)
		w.Write(body) // Echo back
	}

	server, client := setupTestEnv(t, handler)
	defer server.Close()

	svc := documents.NewDocumentsService[TestDoc](client, "books")
	doc := TestDoc{ID: "1", Title: "Go Programming", Views: 100}

	res, err := svc.Create(context.Background(), doc)

	require.NoError(t, err)
	assert.Equal(t, "1", res.ID)
	assert.Equal(t, "Go Programming", res.Title)
}

func TestDocumentsService_Upsert(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/collections/books/documents", r.URL.Path)
		assert.Equal(t, "upsert", r.URL.Query().Get("action"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"2","title":"Upserted","views":50}`))
	}

	server, client := setupTestEnv(t, handler)
	defer server.Close()

	svc := documents.NewDocumentsService[TestDoc](client, "books")

	res, err := svc.Upsert(context.Background(), TestDoc{ID: "2", Title: "Upserted", Views: 50})

	require.NoError(t, err)
	assert.Equal(t, "Upserted", res.Title)
}

func TestDocumentsService_Update(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/collections/books/documents/123", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","title":"Updated","views":10}`))
	}

	server, client := setupTestEnv(t, handler)
	defer server.Close()

	svc := documents.NewDocumentsService[TestDoc](client, "books")

	res, err := svc.Update(context.Background(), "123", TestDoc{Title: "Updated"})

	require.NoError(t, err)
	assert.Equal(t, "Updated", res.Title)
}

func TestDocumentsService_Retrieve(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/collections/books/documents/abc", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"abc","title":"Found","views":1}`))
	}

	server, client := setupTestEnv(t, handler)
	defer server.Close()

	svc := documents.NewDocumentsService[TestDoc](client, "books")

	res, err := svc.Retrieve(context.Background(), "abc")

	require.NoError(t, err)
	assert.Equal(t, "abc", res.ID)
}

func TestDocumentsService_Retrieve_NotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Document not found"}`))
	}

	server, client := setupTestEnv(t, handler)
	defer server.Close()

	svc := documents.NewDocumentsService[TestDoc](client, "books")

	_, err := svc.Retrieve(context.Background(), "missing")

	require.Error(t, err)
	var apiErr *types.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "Document not found", apiErr.Message)
}

func TestDocumentsService_Delete(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/collections/books/documents/del-1", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"del-1","title":"Deleted"}`))
	}

	server, client := setupTestEnv(t, handler)
	defer server.Close()

	svc := documents.NewDocumentsService[TestDoc](client, "books")

	res, err := svc.Delete(context.Background(), "del-1")

	require.NoError(t, err)
	assert.Equal(t, "del-1", res.ID)
}

func TestDocumentsService_DeleteByQuery(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/collections/books/documents", r.URL.Path)
		assert.Equal(t, "views:>10", r.URL.Query().Get("filter_by"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"num_deleted":5}`))
	}

	server, client := setupTestEnv(t, handler)
	defer server.Close()

	svc := documents.NewDocumentsService[TestDoc](client, "books")

	res, err := svc.DeleteByQuery(context.Background(), "views:>10")

	require.NoError(t, err)
	assert.Equal(t, 5, res.NumDeleted)
}

func TestDocumentsService_Search(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/collections/books/documents/search", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, "golang", q.Get("q"))
		assert.Equal(t, "title", q.Get("query_by"))
		assert.Equal(t, "views:desc", q.Get("sort_by"))
		assert.Equal(t, "10", q.Get("per_page"))

		mockResponse := schemas.SearchResult[TestDoc]{
			Found: 1,
			Hits: []schemas.SearchHit[TestDoc]{
				{Document: TestDoc{ID: "1", Title: "golang basics"}},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}

	server, client := setupTestEnv(t, handler)
	defer server.Close()

	svc := documents.NewDocumentsService[TestDoc](client, "books")

	params := schemas.SearchParams{
		Q:       "golang",
		QueryBy: "title",
		SortBy:  "views:desc",
		PerPage: 10,
	}

	res, err := svc.Search(context.Background(), params)

	require.NoError(t, err)
	assert.Equal(t, 1, res.Found)
	assert.Len(t, res.Hits, 1)
	assert.Equal(t, "golang basics", res.Hits[0].Document.Title)
}

func TestDocumentsService_ImportBatch(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/collections/books/documents/import", r.URL.Path)
		assert.Equal(t, "upsert", r.URL.Query().Get("action"))
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

		bodyBytes, _ := io.ReadAll(r.Body)
		bodyStr := string(bodyBytes)

		// Assert strictly JSONL format (newline separated)
		lines := strings.Split(strings.TrimSuffix(bodyStr, "\n"), "\n")
		assert.Len(t, lines, 2)
		assert.Contains(t, lines[0], `"id":"1"`)
		assert.Contains(t, lines[1], `"id":"2"`)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}\n{"success":true}`))
	}

	server, client := setupTestEnv(t, handler)
	defer server.Close()

	svc := documents.NewDocumentsService[TestDoc](client, "books")

	docs := []TestDoc{
		{ID: "1", Title: "Doc 1"},
		{ID: "2", Title: "Doc 2"},
	}

	res, err := svc.ImportBatch(context.Background(), docs, "upsert")

	require.NoError(t, err)
	assert.Contains(t, string(res), "success")
}
