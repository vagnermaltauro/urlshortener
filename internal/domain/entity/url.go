package entity

import "time"

// URL represents the core business entity for a shortened URL.
// This is a pure domain entity with zero external dependencies.
type URL struct {
	ID          int64     // Snowflake ID (unique across all machines)
	ShortCode   string    // Base62 encoded ID (7 characters)
	OriginalURL string    // The original long URL
	Clicks      int64     // Number of times this URL was accessed
	CreatedAt   time.Time // When the URL was created
	ExpiresAt   time.Time // When the URL expires (5 years from creation)
}

// IsExpired checks if the URL has expired
func (u *URL) IsExpired() bool {
	return time.Now().After(u.ExpiresAt)
}

// IsValid performs basic validation on the URL entity
func (u *URL) IsValid() bool {
	return u.ID > 0 &&
		u.ShortCode != "" &&
		u.OriginalURL != "" &&
		u.ExpiresAt.After(u.CreatedAt)
}
