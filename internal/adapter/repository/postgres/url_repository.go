package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"

	"urlshortner/internal/domain/entity"
	"urlshortner/internal/domain/repository"
)

// PostgresURLRepository implements the URLRepository interface using PostgreSQL
type PostgresURLRepository struct {
	writeDB *sql.DB // Primary database for writes
	readDB  *sql.DB // Replica database for reads (can be same as writeDB in dev)
}

// NewPostgresURLRepository creates a new PostgreSQL URL repository
// writeDB should point to the primary (master) database
// readDB should point to a read replica (or primary if no replicas available)
func NewPostgresURLRepository(writeDB, readDB *sql.DB) repository.URLRepository {
	return &PostgresURLRepository{
		writeDB: writeDB,
		readDB:  readDB,
	}
}

// Save persists a URL to PostgreSQL
// Uses INSERT ON CONFLICT to handle potential duplicate IDs (upsert)
func (r *PostgresURLRepository) Save(ctx context.Context, url entity.URL) error {
	query := `
		INSERT INTO urls (id, short_code, original_url, clicks, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id, created_at) DO UPDATE SET
			clicks = EXCLUDED.clicks,
			original_url = EXCLUDED.original_url
	`

	_, err := r.writeDB.ExecContext(ctx, query,
		url.ID,
		url.ShortCode,
		url.OriginalURL,
		url.Clicks,
		url.CreatedAt,
		url.ExpiresAt,
	)

	// Handle unique constraint violation on short_code
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == "23505" { // unique_violation
			return repository.ErrDuplicateKey
		}
	}

	return err
}

// FindByShortCode retrieves a URL by its short code from a read replica
// Returns repository.ErrNotFound if the URL doesn't exist or has expired
func (r *PostgresURLRepository) FindByShortCode(ctx context.Context, shortCode string) (*entity.URL, error) {
	query := `
		SELECT id, short_code, original_url, clicks, created_at, expires_at
		FROM urls
		WHERE short_code = $1 AND expires_at > NOW()
	`

	var url entity.URL
	err := r.readDB.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.Clicks,
		&url.CreatedAt,
		&url.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &url, nil
}

// IncrementClicks atomically increments the click counter for a single URL
func (r *PostgresURLRepository) IncrementClicks(ctx context.Context, shortCode string) error {
	query := `UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1`
	_, err := r.writeDB.ExecContext(ctx, query, shortCode)
	return err
}

// BatchIncrementClicks atomically increments click counters for multiple URLs in a transaction
// This is used by the background job to flush buffered clicks from Redis
// The map key is the short code, the value is the increment amount
func (r *PostgresURLRepository) BatchIncrementClicks(ctx context.Context, clicks map[string]int64) error {
	if len(clicks) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback if we don't commit

	// Prepare statement for reuse
	stmt, err := tx.PrepareContext(ctx, `UPDATE urls SET clicks = clicks + $1 WHERE short_code = $2`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute update for each URL
	for shortCode, count := range clicks {
		if _, err := stmt.ExecContext(ctx, count, shortCode); err != nil {
			return err // Transaction will rollback
		}
	}

	// Commit transaction
	return tx.Commit()
}

// DeleteExpired removes all URLs that expired before the given time
// Returns the number of URLs deleted
// This should be called periodically by a background job
func (r *PostgresURLRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM urls WHERE expires_at < $1`
	result, err := r.writeDB.ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
