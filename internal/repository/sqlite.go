package repository

import (
	"context"
	"database/sql"
	"urlshortner/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id TEXT PRIMARY KEY,
			long_url TEXT NOT NULL,
			clicks INTEGER DEFAULT 0,
			created_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) Save(ctx context.Context, url model.URL) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT OR REPLACE INTO urls (id, long_url, clicks, created_at) VALUES (?, ?, ?, ?)",
		url.ID, url.LongURL, url.Clicks, url.CreatedAt,
	)
	return err
}

func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*model.URL, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, long_url, clicks, created_at FROM urls WHERE id = ?", id)

	var url model.URL
	err := row.Scan(&url.ID, &url.LongURL, &url.Clicks, &url.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Increment clicks
	_, _ = r.db.ExecContext(ctx, "UPDATE urls SET clicks = clicks + 1 WHERE id = ?", id)
	url.Clicks++

	return &url, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
