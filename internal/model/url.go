package model

type URL struct {
	ID        string
	LongURL   string
	Clicks    int64
	CreatedAt string
}

type CreateRequest struct {
	URL string `json:"url"`
}
