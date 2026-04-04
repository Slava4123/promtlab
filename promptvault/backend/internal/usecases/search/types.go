package search

type SearchResult struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

type SearchOutput struct {
	Prompts     []SearchResult `json:"prompts"`
	Collections []SearchResult `json:"collections"`
	Tags        []SearchResult `json:"tags"`
}
