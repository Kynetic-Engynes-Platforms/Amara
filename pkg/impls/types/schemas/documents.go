package schemas

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
