package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"

	"urlshortner/internal/domain/entity"
	"urlshortner/internal/domain/repository"
)

type PostgresURLRepository struct {
	writeDB *sql.DB
	readDB  *sql.DB
}

func NewPostgresURLRepository(writeDB, readDB *sql.DB) repository.URLRepository {
	return &PostgresURLRepository{
		writeDB: writeDB,
		readDB:  readDB,
	}
}

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

	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == "23505" {
			return repository.ErrDuplicateKey
		}
	}

	return err
}

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

func (r *PostgresURLRepository) IncrementClicks(ctx context.Context, shortCode string) error {
	query := `UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1`
	_, err := r.writeDB.ExecContext(ctx, query, shortCode)
	return err
}

func (r *PostgresURLRepository) BatchIncrementClicks(ctx context.Context, clicks map[string]int64) error {
	if len(clicks) == 0 {
		return nil
	}

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `UPDATE urls SET clicks = clicks + $1 WHERE short_code = $2`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for shortCode, count := range clicks {
		if _, err := stmt.ExecContext(ctx, count, shortCode); err != nil {
			return err
		}
	}

	return tx.Commit()
}

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
