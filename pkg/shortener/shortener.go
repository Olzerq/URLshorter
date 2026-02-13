package shortener

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

const (
	ShortURLLength = 10
	base62Chars    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

// Generate создает код из оригЮРЛ
func Generate(originalURL string) string {

	hash := sha256.Sum256([]byte(originalURL))

	encoded := base64.RawURLEncoding.EncodeToString(hash[:])

	encoded = strings.ReplaceAll(encoded, "-", "_")

	if len(encoded) > ShortURLLength {
		encoded = encoded[:ShortURLLength]
	}

	return encoded
}

// Validate проверка валидности шорткода
func Validate(shortCode string) bool {
	if len(shortCode) != ShortURLLength {
		return false
	}

	for _, ch := range shortCode {
		if !isValidChar(ch) {
			return false
		}
	}

	return true
}

func isValidChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}
