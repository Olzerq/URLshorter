package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"shortURL/internal/repository"
)

type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository создает новый реп
func NewPostgresRepository(connString string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// тест конекта
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// Save сохранияет новый shortURL
func (r *PostgresRepository) Save(ctx context.Context, shortCode string, originalURL string) error {
	query := `
		INSERT INTO urls (short_code, original_url, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (short_code) DO NOTHING
	`

	result, err := r.db.ExecContext(ctx, query, shortCode, originalURL)
	if err != nil {
		// проверяем уникальность ориг юрлаа
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrDuplicate
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// shortURL уже есть смотрим на его оригURl
		existingURL, err := r.Get(ctx, shortCode)
		if err != nil {
			return fmt.Errorf("failed to check existing URL: %w", err)
		}
		if existingURL != originalURL {
			return repository.ErrAlreadyExists
		}
		return nil
	}

	return nil
}

// Get получает оригинальный url по shortURL
func (r *PostgresRepository) Get(ctx context.Context, shortCode string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_code = $1`

	var originalURL string
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repository.ErrNotFound
		}
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	return originalURL, nil
}

// GetByOriginal получает shortURL по оригу
func (r *PostgresRepository) GetByOriginal(ctx context.Context, originalURL string) (string, error) {
	query := `SELECT short_code FROM urls WHERE original_url = $1`

	var shortCode string
	err := r.db.QueryRowContext(ctx, query, originalURL).Scan(&shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repository.ErrNotFound
		}
		return "", fmt.Errorf("failed to get short code: %w", err)
	}

	return shortCode, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// InitSchema схема бд
func (r *PostgresRepository) InitSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_code VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);
		
		CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
		CREATE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to init schema: %w", err)
	}

	return nil
}
