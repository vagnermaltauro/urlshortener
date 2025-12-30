package dto

// CreateShortURLRequest represents the HTTP request body for creating a short URL
type CreateShortURLRequest struct {
	URL string `json:"url"`
}
