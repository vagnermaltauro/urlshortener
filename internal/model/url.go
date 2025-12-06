package model

type URL struct {
    ID        string
    LongURL   string
    Clicks    int64
    CreatedAt int64
}

type CreateRequest struct {
    URL string `json:"url"`
}
