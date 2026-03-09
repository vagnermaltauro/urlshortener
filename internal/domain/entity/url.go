package entity

import "time"

type URL struct {
	ID          int64
	ShortCode   string
	OriginalURL string
	Clicks      int64
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

func (u *URL) IsExpired() bool {
	return time.Now().After(u.ExpiresAt)
}

func (u *URL) IsValid() bool {
	return u.ID > 0 &&
		u.ShortCode != "" &&
		u.OriginalURL != "" &&
		u.ExpiresAt.After(u.CreatedAt)
}
