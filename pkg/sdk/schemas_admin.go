package sdk

type OverrideMatchCondition struct {
	Query string `json:"query"`
	Match string `json:"match"`
}

type OverrideInclude struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}

type OverrideExclude struct {
	ID string `json:"id"`
}

type OverrideSchema struct {
	Rule     OverrideMatchCondition `json:"rule"`
	Includes []OverrideInclude      `json:"includes,omitempty"`
	Excludes []OverrideExclude      `json:"excludes,omitempty"`
	FilterBy string                 `json:"filter_by,omitempty"`
	SortBy   string                 `json:"sort_by,omitempty"`
}

type Override struct {
	ID string `json:"id"`
	OverrideSchema
}

type SynonymSchema struct {
	Synonyms []string `json:"synonyms"`
	Root     string   `json:"root,omitempty"`
}

type Synonym struct {
	ID string `json:"id"`
	SynonymSchema
}

type AliasSchema struct {
	CollectionName string `json:"collection_name"`
}

type Alias struct {
	Name string `json:"name"`
	AliasSchema
}

type KeySchema struct {
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
	Collections []string `json:"collections"`
	ExpiresAt   int64    `json:"expires_at,omitempty"`
}

type Key struct {
	ID    int64  `json:"id"`
	Value string `json:"value,omitempty"`
	KeySchema
}

type AnalyticsRuleParameters struct {
	Source      map[string]any `json:"source"`
	Destination map[string]any `json:"destination"`
	Limit       int            `json:"limit,omitempty"`
}

type AnalyticsRuleSchema struct {
	Type   string                  `json:"type"`
	Params AnalyticsRuleParameters `json:"params"`
}

type AnalyticsRule struct {
	Name string `json:"name"`
	AnalyticsRuleSchema
}

type HealthResponse struct {
	Ok bool `json:"ok"`
}

type MetricsResponse struct {
	SystemCPUActivePercentage string `json:"system_cpu_active_percentage"`
	SystemDiskTotalBytes      string `json:"system_disk_total_bytes"`
	SystemDiskUsedBytes       string `json:"system_disk_used_bytes"`
	SystemMemoryTotalBytes    string `json:"system_memory_total_bytes"`
	SystemMemoryUsedBytes     string `json:"system_memory_used_bytes"`
}
