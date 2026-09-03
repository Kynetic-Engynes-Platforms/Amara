package schemas

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
