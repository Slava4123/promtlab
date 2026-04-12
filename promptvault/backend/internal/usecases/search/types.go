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

type Suggestion struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type SuggestOutput struct {
	Suggestions []Suggestion `json:"suggestions"`
}
