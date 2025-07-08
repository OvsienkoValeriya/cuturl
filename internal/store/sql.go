package store

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/lib/pq"
)

type SQLRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(dsn string) (Repository, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	schema := `
CREATE TABLE IF NOT EXISTS urls (
    uuid TEXT PRIMARY KEY,
    short_url TEXT NOT NULL,
    original_url TEXT NOT NULL UNIQUE
);`
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &SQLRepository{db: db}, nil
}

func (r *SQLRepository) Ping() error {
	return r.db.Ping()
}

func (r *SQLRepository) Close() error {
	return r.db.Close()
}

func (r *SQLRepository) Load() ([]StoredURL, error) {
	queryBuilder := sq.Select("uuid", "short_url", "original_url").
		From("urls").
		PlaceholderFormat(sq.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []StoredURL
	for rows.Next() {
		var url StoredURL
		if err := rows.Scan(&url.UUID, &url.ShortURL, &url.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *SQLRepository) Save(entry StoredURL) error {
	queryBuilder := sq.Insert("urls").
		Columns("uuid", "short_url", "original_url").
		Values(entry.UUID, entry.ShortURL, entry.OriginalURL).
		PlaceholderFormat(sq.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(query, args...)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("conflict: %w", err)
		}
		return err
	}
	return nil
}

func (r *SQLRepository) FindByShortID(id string) (*StoredURL, error) {
	queryBuilder := sq.
		Select("uuid", "short_url", "original_url").
		From("urls").
		Where(sq.Eq{"short_url": id}).
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var result StoredURL
	err = r.db.Get(&result, query, args...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *SQLRepository) FindByOriginalURL(original string) (*StoredURL, error) {
	queryBuilder := sq.
		Select("uuid", "short_url", "original_url").
		From("urls").
		Where(sq.Eq{"original_url": original}).
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var result StoredURL
	err = r.db.Get(&result, query, args...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *SQLRepository) BatchSave(ctx context.Context, urls []StoredURL) error {

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmtStr := "INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)"
	stmt, err := tx.PrepareContext(ctx, stmtStr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, u := range urls {
		if _, err := stmt.ExecContext(ctx, u.UUID, u.ShortURL, u.OriginalURL); err != nil {
			return err
		}
	}

	return tx.Commit()
}
