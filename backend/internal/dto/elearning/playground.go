package dto

// Bentuk PlaygroundExample di playgroundService.ts.
type PlaygroundExample struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Language    string  `json:"language"`
	Code        string  `json:"code"`
	Description *string `json:"description"`
	SortOrder   int     `json:"sortOrder"`
}

type PlaygroundInput struct {
	Title       *string `json:"title"`
	Language    *string `json:"language"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sortOrder"`
}

// Persis updatePlaygroundExample: title/language/code hanya bila tidak kosong, description/sortOrder bila dikirim.
type PlaygroundUpdate struct {
	Title       string           `json:"title"`
	Language    string           `json:"language"`
	Code        string           `json:"code"`
	Description Optional[string] `json:"description"`
	SortOrder   Optional[int]    `json:"sortOrder"`
}
