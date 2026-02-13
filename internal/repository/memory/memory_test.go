package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"shortURL/internal/repository"
)

func TestMemoryRepository_SaveAndGet(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	shortCode := "abc123XYZ_"
	originalURL := "https://example.com/test"

	err := repo.Save(ctx, shortCode, originalURL)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// по шортюрл
	gotURL, err := repo.Get(ctx, shortCode)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if gotURL != originalURL {
		t.Errorf("Get() = %s, want %s", gotURL, originalURL)
	}

	// по оригюрл
	gotShort, err := repo.GetByOriginal(ctx, originalURL)
	if err != nil {
		t.Fatalf("GetByOriginal() failed: %v", err)
	}
	if gotShort != shortCode {
		t.Errorf("GetByOriginal() = %s, want %s", gotShort, shortCode)
	}
}

func TestMemoryRepository_GetNotFound(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// попытка получить несуществ шортюрл
	_, err := repo.Get(ctx, "notexists")
	if err != repository.ErrNotFound {
		t.Errorf("Get() error = %v, want %v", err, repository.ErrNotFound)
	}

	// попытка получить несуществующ оригюрл
	_, err = repo.GetByOriginal(ctx, "https://notexists.com")
	if err != repository.ErrNotFound {
		t.Errorf("GetByOriginal() error = %v, want %v", err, repository.ErrNotFound)
	}
}

func TestMemoryRepository_SaveDuplicate(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	shortCode1 := "abc123XYZ_"
	shortCode2 := "xyz789ABC_"
	originalURL := "https://example.com/test"

	err := repo.Save(ctx, shortCode1, originalURL)
	if err != nil {
		t.Fatalf("Save() first failed: %v", err)
	}

	// пытаемся сохр тот же оригюрл с другим шортюрл
	err = repo.Save(ctx, shortCode2, originalURL)
	if err != repository.ErrDuplicate {
		t.Errorf("Save() error = %v, want %v", err, repository.ErrDuplicate)
	}

	err = repo.Save(ctx, shortCode1, originalURL)
	if err != nil {
		t.Errorf("Save() same mapping failed: %v", err)
	}
}

func TestMemoryRepository_SaveShortCodeExists(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	shortCode := "abc123XYZ_"
	url1 := "https://example.com/page1"
	url2 := "https://example.com/page2"

	err := repo.Save(ctx, shortCode, url1)
	if err != nil {
		t.Fatalf("Save() first failed: %v", err)
	}

	err = repo.Save(ctx, shortCode, url2)
	if err != repository.ErrAlreadyExists {
		t.Errorf("Save() error = %v, want %v", err, repository.ErrAlreadyExists)
	}
}

func TestMemoryRepository_Concurrency(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	// запускаем 100 функций
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			shortCode := fmt.Sprintf("s%08d__", id)
			originalURL := fmt.Sprintf("https://example.com/page%d", id)
			err := repo.Save(ctx, shortCode, originalURL)
			if err != nil && err != repository.ErrDuplicate && err != repository.ErrAlreadyExists {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent Save() failed: %v", err)
	}
}

func TestMemoryRepository_ConcurrentReadWrite(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		shortCode := fmt.Sprintf("test%d_____", i)
		originalURL := fmt.Sprintf("https://example.com/page%d", i)
		_ = repo.Save(ctx, shortCode, originalURL)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 200)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			shortCode := fmt.Sprintf("test%d_____", id%10)
			_, err := repo.Get(ctx, shortCode)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			shortCode := fmt.Sprintf("new%07d", id)
			originalURL := fmt.Sprintf("https://example.com/new%d", id)
			err := repo.Save(ctx, shortCode, originalURL)
			if err != nil && err != repository.ErrDuplicate && err != repository.ErrAlreadyExists {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent operation failed: %v", err)
	}
}

func TestMemoryRepository_Clear(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	_ = repo.Save(ctx, "abc123XYZ_", "https://example.com/test")

	repo.Clear()

	_, err := repo.Get(ctx, "abc123XYZ_")
	if err != repository.ErrNotFound {
		t.Errorf("After Clear(), Get() error = %v, want %v", err, repository.ErrNotFound)
	}
}
