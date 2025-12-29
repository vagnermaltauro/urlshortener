package model

type URL struct {
    ID        string
    LongURL   string
    Clicks    int64
    CreatedAt string // ISO 8601 format: 2025-12-06T14:30:00-03:00
}

type CreateRequest struct {
    URL string `json:"url"`
}

