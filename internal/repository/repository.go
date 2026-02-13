package repository

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("URL not found")
	ErrAlreadyExists = errors.New("short code already exists")
	ErrDuplicate     = errors.New("URL already shortened")
)

type URLRepository interface {
	// Save сохраняет новый shortURL
	Save(ctx context.Context, shortCode string, originalURL string) error

	// Get получает оригинальный URL по ShortURL
	Get(ctx context.Context, shortCode string) (string, error)

	// GetByOriginal получает shortURL по оригинальному
	GetByOriginal(ctx context.Context, originalURL string) (string, error)

	// Close закрывает соединение
	Close() error
}
