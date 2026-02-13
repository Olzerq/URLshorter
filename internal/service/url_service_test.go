package service

import (
	"context"
	"testing"

	"shortURL/internal/repository/memory"
)

func TestURLService_Create(t *testing.T) {
	repo := memory.NewMemoryRepository()
	service := NewURLService(repo)
	ctx := context.Background()

	tests := []struct {
		name        string
		url         string
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "valid HTTP URL",
			url:     "http://example.com/test",
			wantErr: false,
		},
		{
			name:    "valid HTTPS URL",
			url:     "https://example.com/test",
			wantErr: false,
		},
		{
			name:    "valid URL with path and query",
			url:     "https://example.com/path/to/page?param=value",
			wantErr: false,
		},
		{
			name:        "empty URL",
			url:         "",
			wantErr:     true,
			expectedErr: ErrInvalidURL,
		},
		{
			name:        "invalid URL without scheme",
			url:         "example.com",
			wantErr:     true,
			expectedErr: ErrInvalidURL,
		},
		{
			name:        "invalid URL with ftp scheme",
			url:         "ftp://example.com",
			wantErr:     true,
			expectedErr: ErrInvalidURL,
		},
		{
			name:        "invalid URL without host",
			url:         "https://",
			wantErr:     true,
			expectedErr: ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortCode, err := service.Create(ctx, tt.url)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error, got nil")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Create() error = %v, want %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("Create() unexpected error: %v", err)
				}
				if shortCode == "" {
					t.Errorf("Create() returned empty short code")
				}
				if len(shortCode) != 10 {
					t.Errorf("Create() short code length = %d, want 10", len(shortCode))
				}
			}
		})
	}
}

func TestURLService_CreateIdempotency(t *testing.T) {
	repo := memory.NewMemoryRepository()
	service := NewURLService(repo)
	ctx := context.Background()

	url := "https://example.com/test"

	// создаем первый раз
	shortCode1, err := service.Create(ctx, url)
	if err != nil {
		t.Fatalf("Create() first call failed: %v", err)
	}

	// создаем еще раз с тем же URL
	shortCode2, err := service.Create(ctx, url)
	if err != nil {
		t.Fatalf("Create() second call failed: %v", err)
	}

	// должно вернуть тот же шорт URL
	if shortCode1 != shortCode2 {
		t.Errorf("Create() not idempotent: got %s and %s", shortCode1, shortCode2)
	}
}

func TestURLService_Resolve(t *testing.T) {
	repo := memory.NewMemoryRepository()
	service := NewURLService(repo)
	ctx := context.Background()

	originalURL := "https://example.com/test"
	shortCode, err := service.Create(ctx, originalURL)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	resolvedURL, err := service.Resolve(ctx, shortCode)
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	if resolvedURL != originalURL {
		t.Errorf("Resolve() = %s, want %s", resolvedURL, originalURL)
	}
}

func TestURLService_ResolveNotFound(t *testing.T) {
	repo := memory.NewMemoryRepository()
	service := NewURLService(repo)
	ctx := context.Background()

	_, err := service.Resolve(ctx, "notexist__")
	if err == nil {
		t.Errorf("Resolve() expected error, got nil")
	}
}

func TestURLService_ResolveInvalidShortCode(t *testing.T) {
	repo := memory.NewMemoryRepository()
	service := NewURLService(repo)
	ctx := context.Background()

	tests := []struct {
		name      string
		shortCode string
	}{
		{
			name:      "too short",
			shortCode: "abc",
		},
		{
			name:      "too long",
			shortCode: "abc123456789",
		},
		{
			name:      "invalid characters",
			shortCode: "abc@123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Resolve(ctx, tt.shortCode)
			if err == nil {
				t.Errorf("Resolve() with invalid short code expected error, got nil")
			}
		})
	}
}

func TestURLService_MultipleDifferentURLs(t *testing.T) {
	repo := memory.NewMemoryRepository()
	service := NewURLService(repo)
	ctx := context.Background()

	url1 := "https://example.com/page1"
	url2 := "https://example.com/page2"

	shortCode1, err := service.Create(ctx, url1)
	if err != nil {
		t.Fatalf("Create() url1 failed: %v", err)
	}

	shortCode2, err := service.Create(ctx, url2)
	if err != nil {
		t.Fatalf("Create() url2 failed: %v", err)
	}

	if shortCode1 == shortCode2 {
		t.Errorf("Different URLs produced same short code: %s", shortCode1)
	}

	resolved1, _ := service.Resolve(ctx, shortCode1)
	resolved2, _ := service.Resolve(ctx, shortCode2)

	if resolved1 != url1 {
		t.Errorf("Resolve() url1 = %s, want %s", resolved1, url1)
	}
	if resolved2 != url2 {
		t.Errorf("Resolve() url2 = %s, want %s", resolved2, url2)
	}
}
