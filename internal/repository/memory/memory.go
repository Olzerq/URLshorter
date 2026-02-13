package memory

import (
	"context"
	"sync"

	"shortURL/internal/repository"
)

type MemoryRepository struct {
	mu              sync.RWMutex
	shortToOriginal map[string]string
	originalToShort map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		shortToOriginal: make(map[string]string),
		originalToShort: make(map[string]string),
	}
}

// Save сохраняет новый shortURL
func (r *MemoryRepository) Save(ctx context.Context, shortCode string, originalURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// проверка на существование этого URL
	if existingURL, exists := r.shortToOriginal[shortCode]; exists {
		if existingURL == originalURL {
			return nil
		}
		return repository.ErrAlreadyExists
	}

	// проверяем есть ли у ориг URL shortURL
	if existingShort, exists := r.originalToShort[originalURL]; exists {
		if existingShort != shortCode {
			return repository.ErrDuplicate
		}
		return nil
	}

	r.shortToOriginal[shortCode] = originalURL
	r.originalToShort[originalURL] = shortCode

	return nil
}

// Get Получиет ориг URL по shortURL
func (r *MemoryRepository) Get(ctx context.Context, shortCode string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	originalURL, exists := r.shortToOriginal[shortCode]
	if !exists {
		return "", repository.ErrNotFound
	}

	return originalURL, nil
}

// GetByOriginal получает shortURL по оригу
func (r *MemoryRepository) GetByOriginal(ctx context.Context, originalURL string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortCode, exists := r.originalToShort[originalURL]
	if !exists {
		return "", repository.ErrNotFound
	}

	return shortCode, nil
}

func (r *MemoryRepository) Close() error {
	return nil
}

func (r *MemoryRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.shortToOriginal = make(map[string]string)
	r.originalToShort = make(map[string]string)
}
