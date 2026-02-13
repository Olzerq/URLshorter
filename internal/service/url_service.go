package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"shortURL/internal/repository"
	"shortURL/pkg/shortener"
)

var (
	ErrInvalidURL = errors.New("invalid URL format")
)

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{
		repo: repo,
	}
}

// Create создаем shortURL
func (s *URLService) Create(ctx context.Context, originalURL string) (string, error) {
	if err := s.validateURL(originalURL); err != nil {
		return "", err
	}

	// Проверяем есть ли у юрл шортюрл
	existingShort, err := s.repo.GetByOriginal(ctx, originalURL)
	if err == nil {
		// URL already shortened, return existing code
		return existingShort, nil
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return "", fmt.Errorf("failed to check existing URL: %w", err)
	}

	// генерируем shortUrl
	shortCode := shortener.Generate(originalURL)

	// сохраняем
	err = s.repo.Save(ctx, shortCode, originalURL)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return "", fmt.Errorf("short code collision detected: %w", err)
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	return shortCode, nil
}

func (s *URLService) Resolve(ctx context.Context, shortCode string) (string, error) {
	if !shortener.Validate(shortCode) {
		return "", repository.ErrNotFound
	}

	originalURL, err := s.repo.Get(ctx, shortCode)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

// validateURL проверяем валидность URL
func (s *URLService) validateURL(urlStr string) error {
	if urlStr == "" {
		return ErrInvalidURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrInvalidURL
	}

	if parsedURL.Host == "" {
		return ErrInvalidURL
	}

	return nil
}
